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
// Returns the number of evaluated nodes grouped by type, and the number of calls
// for each function.
func Execute(functions []*Function, main Node, globals map[string]Node, w io.Writer, r io.Reader) (map[string]int, map[string]int) {
	findFunction := func(name string) *Function {
		for _, function := range functions {
			if function.Name == name {
				return function
			}
		}

		log.Fatal("function not found")
		return nil
	}

	evaluated := map[string]int{}
	called := map[string]int{}

	globalValues := map[string]interface{}{}

	var evaluate func(Node, map[string]interface{}) interface{}
	evaluate = func(node Node, values map[string]interface{}) interface{} {
		{
			op := reflect.TypeOf(node).String()
			op = op[strings.LastIndex(op, ".")+1:]
			evaluated[op]++
		}

		switch n := node.(type) {
		case *Variable:
			if v, ok := globalValues[n.Name]; ok {
				return v
			}
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
			return values[n.Left].(int32) + values[n.Right].(int32)
		case *AddImmediate:
			return values[n.Left].(int32) + n.Right
		case *Sub:
			return values[n.Left].(int32) - values[n.Right].(int32)
		case *SubFromZero:
			return -values[n.Inner].(int32)
		case *FloatAdd:
			return values[n.Left].(float32) + values[n.Right].(float32)
		case *FloatSub:
			return values[n.Left].(float32) - values[n.Right].(float32)
		case *FloatSubFromZero:
			return -values[n.Inner].(float32)
		case *FloatDiv:
			return values[n.Left].(float32) / values[n.Right].(float32)
		case *FloatMul:
			return values[n.Left].(float32) * values[n.Right].(float32)
		case *Not:
			return !values[n.Inner].(bool)
		case *Equal:
			if values[n.Left] == values[n.Right] {
				return true
			} else {
				return false
			}
		case *EqualZero:
			return values[n.Inner] == int32(0) || values[n.Inner] == float32(0)
		case *LessThan:
			return values[n.Left].(int32) < values[n.Right].(int32)
		case *LessThanFloat:
			return values[n.Left].(float32) < values[n.Right].(float32)
		case *LessThanZero:
			return values[n.Inner].(int32) < 0
		case *LessThanZeroFloat:
			return values[n.Inner].(float32) < 0
		case *GreaterThanZero:
			return values[n.Inner].(int32) > 0
		case *GreaterThanZeroFloat:
			return values[n.Inner].(float32) > 0
		case *IfEqual:
			if values[n.Left] == values[n.Right] {
				return evaluate(n.True, values)
			} else {
				return evaluate(n.False, values)
			}
		case *IfEqualZero:
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
			if values[n.Inner].(bool) {
				return evaluate(n.True, values)
			} else {
				return evaluate(n.False, values)
			}
		case *IfLessThan:
			if values[n.Left].(int32) < values[n.Right].(int32) {
				return evaluate(n.True, values)
			}
			return evaluate(n.False, values)
		case *IfLessThanFloat:
			if values[n.Left].(float32) < values[n.Right].(float32) {
				return evaluate(n.True, values)
			}
			return evaluate(n.False, values)
		case *IfLessThanZero:
			if values[n.Inner].(int32) < 0 {
				return evaluate(n.True, values)
			} else {
				return evaluate(n.False, values)
			}
		case *IfLessThanZeroFloat:
			if values[n.Inner].(float32) < 0 {
				return evaluate(n.True, values)
			} else {
				return evaluate(n.False, values)
			}
		case *Assignment:
			values[n.Name] = evaluate(n.Value, values)
			ret := evaluate(n.Next, values)
			delete(values, n.Name)
			return ret
		case *Application:
			f := findFunction(n.Function)
			called[f.Name]++
			updated := map[string]interface{}{}
			for i, arg := range f.Args {
				if value, ok := globalValues[n.Args[i]]; ok {
					updated[arg] = value
				} else {
					updated[arg] = values[n.Args[i]]
				}
			}
			return evaluate(f.Body, updated)
		case *Tuple:
			tuple := []interface{}{}
			for _, element := range n.Elements {
				tuple = append(tuple, values[element])
			}
			return tuple
		case *ArrayCreate:
			length := values[n.Length].(int32)
			value := values[n.Value]
			array := []interface{}{}
			for i := 0; i < int(length); i++ {
				array = append(array, value)
			}
			return array
		case *ArrayCreateImmediate:
			value := values[n.Value]
			array := []interface{}{}
			for i := 0; i < int(n.Length); i++ {
				array = append(array, value)
			}
			return array
		case *ArrayGet:
			array := values[n.Array].([]interface{})
			index := values[n.Index].(int32)
			return array[index]
		case *ArrayGetImmediate:
			array := values[n.Array].([]interface{})
			return array[n.Index]
		case *ArrayPut:
			array := values[n.Array].([]interface{})
			index := values[n.Index].(int32)
			value := values[n.Value]
			array[index] = value
			return nil
		case *ArrayPutImmediate:
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
			w.Write([]byte{byte(values[n.Arg].(int32) % 256)})
			return nil
		case *IntToFloat:
			return float32(values[n.Arg].(int32))
		case *FloatToInt:
			return int32(math.Round(float64(values[n.Arg].(float32))))
		case *Sqrt:
			return float32(math.Sqrt(float64(values[n.Arg].(float32))))
		case *TupleGet:
			tuple := values[n.Tuple].([]interface{})
			return tuple[n.Index]
		default:
			log.Fatal("invalid ir node")
		}

		return nil
	}

	for name, node := range globals {
		globalValues[name] = evaluate(node, globalValues)
	}

	evaluate(main, map[string]interface{}{})

	return evaluated, called
}
