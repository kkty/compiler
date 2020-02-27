package ir

import (
	"fmt"
	"io"
	"log"
	"math"
	"reflect"
	"strings"

	"github.com/kkty/compiler/stringset"
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

		getValue := func(name string) interface{} {
			if v, ok := values[name]; ok {
				return v
			}
			if v, ok := globalValues[name]; ok {
				return v
			}
			panic(fmt.Sprintf("variable not found: %s", name))
		}

		switch n := node.(type) {
		case *Variable:
			return getValue(n.Name)
		case *Unit:
			return nil
		case *Int:
			return node.(*Int).Value
		case *Bool:
			return node.(*Bool).Value
		case *Float:
			return node.(*Float).Value
		case *Add:
			return getValue(n.Left).(int32) + getValue(n.Right).(int32)
		case *AddImmediate:
			return getValue(n.Left).(int32) + n.Right
		case *Sub:
			return getValue(n.Left).(int32) - getValue(n.Right).(int32)
		case *SubFromZero:
			return -getValue(n.Inner).(int32)
		case *FloatAdd:
			return getValue(n.Left).(float32) + getValue(n.Right).(float32)
		case *FloatSub:
			return getValue(n.Left).(float32) - getValue(n.Right).(float32)
		case *FloatSubFromZero:
			return -getValue(n.Inner).(float32)
		case *FloatDiv:
			return getValue(n.Left).(float32) / getValue(n.Right).(float32)
		case *FloatMul:
			return getValue(n.Left).(float32) * getValue(n.Right).(float32)
		case *Not:
			return !getValue(n.Inner).(bool)
		case *Equal:
			if getValue(n.Left) == getValue(n.Right) {
				return true
			} else {
				return false
			}
		case *EqualZero:
			return getValue(n.Inner) == int32(0) || getValue(n.Inner) == float32(0)
		case *LessThan:
			return getValue(n.Left).(int32) < getValue(n.Right).(int32)
		case *LessThanFloat:
			return getValue(n.Left).(float32) < getValue(n.Right).(float32)
		case *LessThanZero:
			return getValue(n.Inner).(int32) < 0
		case *LessThanZeroFloat:
			return getValue(n.Inner).(float32) < 0
		case *GreaterThanZero:
			return getValue(n.Inner).(int32) > 0
		case *GreaterThanZeroFloat:
			return getValue(n.Inner).(float32) > 0
		case *IfEqual:
			if getValue(n.Left) == getValue(n.Right) {
				return evaluate(n.True, values)
			} else {
				return evaluate(n.False, values)
			}
		case *IfEqualZero:
			if value, ok := getValue(n.Inner).(int32); ok {
				if value == 0 {
					return evaluate(n.True, values)
				} else {
					return evaluate(n.False, values)
				}
			} else if value, ok := getValue(n.Inner).(float32); ok {
				if value == 0 {
					return evaluate(n.True, values)
				} else {
					return evaluate(n.False, values)
				}
			}
		case *IfEqualTrue:
			if getValue(n.Inner).(bool) {
				return evaluate(n.True, values)
			} else {
				return evaluate(n.False, values)
			}
		case *IfLessThan:
			if getValue(n.Left).(int32) < getValue(n.Right).(int32) {
				return evaluate(n.True, values)
			}
			return evaluate(n.False, values)
		case *IfLessThanFloat:
			if getValue(n.Left).(float32) < getValue(n.Right).(float32) {
				return evaluate(n.True, values)
			}
			return evaluate(n.False, values)
		case *IfLessThanZero:
			if getValue(n.Inner).(int32) < 0 {
				return evaluate(n.True, values)
			} else {
				return evaluate(n.False, values)
			}
		case *IfLessThanZeroFloat:
			if getValue(n.Inner).(float32) < 0 {
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
				updated[arg] = getValue(n.Args[i])
			}
			return evaluate(f.Body, updated)
		case *Tuple:
			tuple := []interface{}{}
			for _, element := range n.Elements {
				tuple = append(tuple, getValue(element))
			}
			return tuple
		case *ArrayCreate:
			length := getValue(n.Length).(int32)
			value := getValue(n.Value)
			array := []interface{}{}
			for i := 0; i < int(length); i++ {
				array = append(array, value)
			}
			return array
		case *ArrayCreateImmediate:
			value := getValue(n.Value)
			array := []interface{}{}
			for i := 0; i < int(n.Length); i++ {
				array = append(array, value)
			}
			return array
		case *ArrayGet:
			array := getValue(n.Array).([]interface{})
			index := getValue(n.Index).(int32)
			return array[index]
		case *ArrayGetImmediate:
			array := getValue(n.Array).([]interface{})
			return array[n.Index]
		case *ArrayPut:
			array := getValue(n.Array).([]interface{})
			index := getValue(n.Index).(int32)
			value := getValue(n.Value)
			array[index] = value
			return nil
		case *ArrayPutImmediate:
			array := getValue(n.Array).([]interface{})
			value := getValue(n.Value)
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
			w.Write([]byte{byte(getValue(n.Arg).(int32) % 256)})
			return nil
		case *IntToFloat:
			return float32(getValue(n.Arg).(int32))
		case *FloatToInt:
			return int32(math.Round(float64(getValue(n.Arg).(float32))))
		case *Sqrt:
			return float32(math.Sqrt(float64(getValue(n.Arg).(float32))))
		case *TupleGet:
			tuple := getValue(n.Tuple).([]interface{})
			return tuple[n.Index]
		default:
			log.Fatal("invalid ir node")
		}

		return nil
	}

	{
		defined := stringset.New()
		for len(defined) < len(globals) {
			for name, node := range globals {
				if !defined.Has(name) && len(node.FreeVariables(defined)) == 0 {
					globalValues[name] = evaluate(node, globalValues)
					defined.Add(name)
				}
			}
		}
	}

	evaluate(main, map[string]interface{}{})

	return evaluated, called
}
