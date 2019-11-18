package interpreter

import (
	"fmt"
	"io"
	"log"
	"math"

	"github.com/kkty/mincaml-go/ir"
)

// Execute interprets and executes the program.
func Execute(functions []ir.Function, main ir.Node, w io.Writer, r io.Reader) {
	findFunction := func(name string) ir.Function {
		for _, function := range functions {
			if function.Name == name {
				return function
			}
		}

		log.Fatal("function not found")
		return ir.Function{}
	}

	copyValues := func(original map[string]interface{}) map[string]interface{} {
		copied := map[string]interface{}{}
		for k, v := range original {
			copied[k] = v
		}
		return copied
	}

	var evaluate func(ir.Node, map[string]interface{}) interface{}
	evaluate = func(node ir.Node, values map[string]interface{}) interface{} {
		switch node.(type) {
		case ir.Variable:
			n := node.(ir.Variable)
			return values[n.Name]
		case ir.Unit:
			return nil
		case ir.Int:
			return node.(ir.Int).Value
		case ir.Bool:
			return node.(ir.Bool).Value
		case ir.Float:
			return node.(ir.Float).Value
		case ir.Add:
			n := node.(ir.Add)
			return values[n.Left].(int32) + values[n.Right].(int32)
		case ir.Sub:
			n := node.(ir.Sub)
			return values[n.Left].(int32) - values[n.Right].(int32)
		case ir.FloatAdd:
			n := node.(ir.FloatAdd)
			return values[n.Left].(float32) + values[n.Right].(float32)
		case ir.FloatSub:
			n := node.(ir.FloatSub)
			return values[n.Left].(float32) - values[n.Right].(float32)
		case ir.FloatDiv:
			n := node.(ir.FloatDiv)
			return values[n.Left].(float32) / values[n.Right].(float32)
		case ir.FloatMul:
			n := node.(ir.FloatMul)
			return values[n.Left].(float32) * values[n.Right].(float32)
		case ir.IfEqual:
			n := node.(ir.IfEqual)
			if values[n.Left] == values[n.Right] {
				return evaluate(n.True, values)
			} else {
				return evaluate(n.False, values)
			}
		case ir.IfLessThanOrEqual:
			n := node.(ir.IfLessThanOrEqual)

			var condition bool
			switch values[n.Left].(type) {
			case int32:
				condition = values[n.Left].(int32) <= values[n.Right].(int32)
			case float32:
				condition = values[n.Left].(float32) <= values[n.Right].(float32)
			}

			if condition {
				return evaluate(n.True, values)
			}

			return evaluate(n.False, values)
		case ir.ValueBinding:
			n := node.(ir.ValueBinding)
			values[n.Name] = evaluate(n.Value, values)
			return evaluate(n.Next, values)
		case ir.Application:
			n := node.(ir.Application)
			f := findFunction(n.Function)
			values := copyValues(values)
			for i, arg := range f.Args {
				values[arg] = values[n.Args[i]]
			}
			return evaluate(f.Body, values)
		case ir.Tuple:
			n := node.(ir.Tuple)
			tuple := []interface{}{}
			for _, element := range n.Elements {
				tuple = append(tuple, values[element])
			}
			return tuple
		case ir.TupleBinding:
			n := node.(ir.TupleBinding)
			for i, element := range values[n.Tuple].([]interface{}) {
				values[n.Names[i]] = element
			}

			return evaluate(n.Next, values)
		case ir.ArrayCreate:
			n := node.(ir.ArrayCreate)
			size := values[n.Size].(int32)
			value := values[n.Value]
			array := []interface{}{}
			for i := 0; i < int(size); i++ {
				array = append(array, value)
			}
			return array
		case ir.ArrayGet:
			n := node.(ir.ArrayGet)
			array := values[n.Array].([]interface{})
			index := values[n.Index].(int32)
			return array[index]
		case ir.ArrayPut:
			n := node.(ir.ArrayPut)
			array := values[n.Array].([]interface{})
			index := values[n.Index].(int32)
			value := values[n.Value]
			array[index] = value
			return nil
		case ir.ReadInt:
			var value int32
			fmt.Fscan(r, &value)
			return value
		case ir.ReadFloat:
			var value float32
			fmt.Fscan(r, &value)
			return value
		case ir.PrintInt:
			n := node.(ir.PrintInt)
			fmt.Fprintf(w, "%d", values[n.Arg].(int32))
			return nil
		case ir.PrintChar:
			n := node.(ir.PrintChar)
			fmt.Fprintf(w, "%c", rune(values[n.Arg].(int32)))
			return nil
		case ir.IntToFloat:
			n := node.(ir.IntToFloat)
			return float32(values[n.Arg].(int32))
		case ir.FloatToInt:
			n := node.(ir.FloatToInt)
			return int32(math.Round(float64(values[n.Arg].(float32))))
		case ir.Sqrt:
			n := node.(ir.Sqrt)
			return float32(math.Sqrt(float64(values[n.Arg].(float32))))
		default:
			log.Fatal("invalid ir node")
		}

		return nil
	}

	evaluate(main, map[string]interface{}{})
}
