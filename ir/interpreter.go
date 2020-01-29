package ir

import (
	"fmt"
	"io"
	"log"
	"math"
	"reflect"
	"strings"
)

// Execute interprets and executes the program.
// Returns the number of evaluated nodes grouped by type.
func Execute(functions []*Function, main Node, w io.Writer, r io.Reader) map[string]int {
	findFunction := func(name string) *Function {
		for _, function := range functions {
			if function.Name == name {
				return function
			}
		}

		log.Fatal("function not found")
		return nil
	}

	counter := map[string]int{}

	var evaluate func(Node, map[string]interface{}) interface{}
	evaluate = func(node Node, values map[string]interface{}) interface{} {
		{
			op := reflect.TypeOf(node).String()
			op = op[strings.LastIndex(op, ".")+1:]
			counter[op]++
		}

		switch node.(type) {
		case *Variable:
			n := node.(*Variable)
			return values[n.Name]
		case *Unit:
			return nil
		case *Int:
			return node.(*Int).Value
		case *Bool:
			return node.(*Bool).Value
		case *Float:
			return node.(*Float).Value
		case *Add:
			n := node.(*Add)
			return values[n.Left].(int32) + values[n.Right].(int32)
		case *AddImmediate:
			n := node.(*AddImmediate)
			return values[n.Left].(int32) + n.Right
		case *Sub:
			n := node.(*Sub)
			return values[n.Left].(int32) - values[n.Right].(int32)
		case *SubFromZero:
			n := node.(*SubFromZero)
			return -values[n.Inner].(int32)
		case *FloatAdd:
			n := node.(*FloatAdd)
			return values[n.Left].(float32) + values[n.Right].(float32)
		case *FloatSub:
			n := node.(*FloatSub)
			return values[n.Left].(float32) - values[n.Right].(float32)
		case *FloatSubFromZero:
			n := node.(*FloatSubFromZero)
			return -values[n.Inner].(float32)
		case *FloatDiv:
			n := node.(*FloatDiv)
			return values[n.Left].(float32) / values[n.Right].(float32)
		case *FloatMul:
			n := node.(*FloatMul)
			return values[n.Left].(float32) * values[n.Right].(float32)
		case *Not:
			n := node.(*Not)
			return !values[n.Inner].(bool)
		case *Equal:
			n := node.(*Equal)
			if values[n.Left] == values[n.Right] {
				return true
			} else {
				return false
			}
		case *LessThan:
			n := node.(*LessThan)
			if left, ok := values[n.Left].(int32); ok {
				return left < values[n.Right].(int32)
			}
			return values[n.Left].(float32) < values[n.Right].(float32)
		case *IfEqual:
			n := node.(*IfEqual)
			if values[n.Left] == values[n.Right] {
				return evaluate(n.True, values)
			} else {
				return evaluate(n.False, values)
			}
		case *IfEqualZero:
			n := node.(*IfEqualZero)

			if value, ok := values[n.Inner].(int32); ok {
				if value == 0 {
					return evaluate(n.True, values)
				} else {
					return evaluate(n.False, values)
				}
			} else if value, ok := values[n.Inner].(float32); ok {
				if value == 0 {
					return evaluate(n.True, values)
				} else {
					return evaluate(n.False, values)
				}
			}
		case *IfEqualTrue:
			n := node.(*IfEqualTrue)

			if values[n.Inner].(bool) {
				return evaluate(n.True, values)
			} else {
				return evaluate(n.False, values)
			}
		case *IfLessThan:
			n := node.(*IfLessThan)

			var condition bool
			switch values[n.Left].(type) {
			case int32:
				condition = values[n.Left].(int32) < values[n.Right].(int32)
			case float32:
				condition = values[n.Left].(float32) < values[n.Right].(float32)
			}

			if condition {
				return evaluate(n.True, values)
			}

			return evaluate(n.False, values)
		case *IfLessThanZero:
			n := node.(*IfLessThanZero)

			if value, ok := values[n.Inner].(int32); ok {
				if value < 0 {
					return evaluate(n.True, values)
				} else {
					return evaluate(n.False, values)
				}
			} else if value, ok := values[n.Inner].(float32); ok {
				if value < 0 {
					return evaluate(n.True, values)
				} else {
					return evaluate(n.False, values)
				}
			}
		case *Assignment:
			n := node.(*Assignment)
			values[n.Name] = evaluate(n.Value, values)
			ret := evaluate(n.Next, values)
			delete(values, n.Name)
			return ret
		case *Application:
			n := node.(*Application)
			f := findFunction(n.Function)
			updated := map[string]interface{}{}
			for i, arg := range f.Args {
				updated[arg] = values[n.Args[i]]
			}
			return evaluate(f.Body, updated)
		case *Tuple:
			n := node.(*Tuple)
			tuple := []interface{}{}
			for _, element := range n.Elements {
				tuple = append(tuple, values[element])
			}
			return tuple
		case *ArrayCreate:
			n := node.(*ArrayCreate)
			length := values[n.Length].(int32)
			value := values[n.Value]
			array := []interface{}{}
			for i := 0; i < int(length); i++ {
				array = append(array, value)
			}
			return array
		case *ArrayCreateImmediate:
			n := node.(*ArrayCreateImmediate)
			value := values[n.Value]
			array := []interface{}{}
			for i := 0; i < int(n.Length); i++ {
				array = append(array, value)
			}
			return array
		case *ArrayGet:
			n := node.(*ArrayGet)
			array := values[n.Array].([]interface{})
			index := values[n.Index].(int32)
			return array[index]
		case *ArrayGetImmediate:
			n := node.(*ArrayGetImmediate)
			array := values[n.Array].([]interface{})
			return array[n.Index]
		case *ArrayPut:
			n := node.(*ArrayPut)
			array := values[n.Array].([]interface{})
			index := values[n.Index].(int32)
			value := values[n.Value]
			array[index] = value
			return nil
		case *ArrayPutImmediate:
			n := node.(*ArrayPutImmediate)
			array := values[n.Array].([]interface{})
			value := values[n.Value]
			array[n.Index] = value
			return nil
		case *ReadInt:
			var value int32
			fmt.Fscan(r, &value)
			return value
		case *ReadFloat:
			var value float32
			fmt.Fscan(r, &value)
			return value
		case *WriteByte:
			n := node.(*WriteByte)
			w.Write([]byte{byte(values[n.Arg].(int32) % 256)})
			return nil
		case *IntToFloat:
			n := node.(*IntToFloat)
			return float32(values[n.Arg].(int32))
		case *FloatToInt:
			n := node.(*FloatToInt)
			return int32(math.Round(float64(values[n.Arg].(float32))))
		case *Sqrt:
			n := node.(*Sqrt)
			return float32(math.Sqrt(float64(values[n.Arg].(float32))))
		case *TupleGet:
			n := node.(*TupleGet)
			tuple := values[n.Tuple].([]interface{})
			return tuple[n.Index]
		default:
			log.Fatal("invalid ir node")
		}

		return nil
	}

	evaluate(main, map[string]interface{}{})

	return counter
}
