package lifting

import (
	"fmt"
	"log"

	"github.com/kkty/mincaml-go/ir"
	"github.com/kkty/mincaml-go/mir"
	"github.com/kkty/mincaml-go/typing"
)

func Lift(root mir.Node, types map[string]typing.Type) (ir.Node, []ir.Function, map[string]typing.Type) {
	// Gathers functions in the program.
	queue := []mir.Node{root}
	functions := map[string]mir.FunctionBinding{}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		switch node.(type) {
		case mir.FunctionBinding:
			n := node.(mir.FunctionBinding)
			functions[n.Name] = n
			queue = append(queue, n.Body, n.Next)
		case mir.IfEqual:
			n := node.(mir.IfEqual)
			queue = append(queue, n.True, n.False)
		case mir.IfLessThanOrEqual:
			n := node.(mir.IfLessThanOrEqual)
			queue = append(queue, n.True, n.False)
		case mir.ValueBinding:
			n := node.(mir.ValueBinding)
			queue = append(queue, n.Value, n.Next)
		case mir.TupleBinding:
			n := node.(mir.TupleBinding)
			queue = append(queue, n.Next)
		}
	}

	// Constructs ir.Node from mir.Node.
	// Function bindings are removed and function applications are modified.
	var construct func(node mir.Node) ir.Node
	construct = func(node mir.Node) ir.Node {
		switch node.(type) {
		case mir.Variable:
			return ir.Variable{node.(mir.Variable).Name}
		case mir.Unit:
			return ir.Unit{}
		case mir.Int:
			return ir.Int{node.(mir.Int).Value}
		case mir.Bool:
			return ir.Bool{node.(mir.Bool).Value}
		case mir.Float:
			return ir.Float{node.(mir.Float).Value}
		case mir.Add:
			n := node.(mir.Add)
			return ir.Add{n.Left, n.Right}
		case mir.Sub:
			n := node.(mir.Sub)
			return ir.Sub{n.Left, n.Right}
		case mir.FloatAdd:
			n := node.(mir.FloatAdd)
			return ir.FloatAdd{n.Left, n.Right}
		case mir.FloatSub:
			n := node.(mir.FloatSub)
			return ir.FloatSub{n.Left, n.Right}
		case mir.FloatDiv:
			n := node.(mir.FloatDiv)
			return ir.FloatDiv{n.Left, n.Right}
		case mir.FloatMul:
			n := node.(mir.FloatMul)
			return ir.FloatMul{n.Left, n.Right}
		case mir.IfEqual:
			n := node.(mir.IfEqual)
			return ir.IfEqual{n.Left, n.Right, construct(n.True), construct(n.False)}
		case mir.IfLessThanOrEqual:
			n := node.(mir.IfLessThanOrEqual)
			return ir.IfLessThanOrEqual{n.Left, n.Right, construct(n.True), construct(n.False)}
		case mir.ValueBinding:
			n := node.(mir.ValueBinding)
			return ir.ValueBinding{n.Name, construct(n.Value), construct(n.Next)}
		case mir.FunctionBinding:
			n := node.(mir.FunctionBinding)
			functions[n.Name] = n
			return construct(n.Next)
		case mir.Application:
			n := node.(mir.Application)
			return ir.Application{
				n.Function,
				append(n.Args, functions[n.Function].FreeVariables(map[string]struct{}{})...),
			}
		case mir.Tuple:
			return ir.Tuple{node.(mir.Tuple).Elements}
		case mir.TupleBinding:
			n := node.(mir.TupleBinding)
			return ir.TupleBinding{n.Names, n.Tuple, construct(n.Next)}
		case mir.ArrayCreate:
			n := node.(mir.ArrayCreate)
			return ir.ArrayCreate{n.Size, n.Value}
		case mir.ArrayGet:
			n := node.(mir.ArrayGet)
			return ir.ArrayGet{n.Array, n.Index}
		case mir.ArrayPut:
			n := node.(mir.ArrayPut)
			return ir.ArrayPut{n.Array, n.Index, n.Value}
		case mir.ReadInt:
			return ir.ReadInt{}
		case mir.ReadFloat:
			return ir.ReadFloat{}
		case mir.PrintInt:
			return ir.PrintInt{node.(mir.PrintInt).Arg}
		case mir.PrintChar:
			return ir.PrintChar{node.(mir.PrintChar).Arg}
		}

		log.Fatal("invalid mir node")
		return nil
	}

	constructed := construct(root)

	nextVarId := 0
	newVar := func() string {
		defer func() { nextVarId++ }()
		return fmt.Sprintf("_l_%d", nextVarId)
	}

	// Creates ir.Function from mir.FunctionBinding and updates their types.
	newFunctions := []ir.Function{}

	for _, function := range functions {
		mapping := map[string]string{}
		args := function.Args

		for _, freeVariable := range function.FreeVariables(map[string]struct{}{}) {
			v := newVar()
			mapping[freeVariable] = v
			types[v] = types[freeVariable]
			args = append(args, v)
			types[function.Name] = typing.FunctionType{
				append(types[function.Name].(typing.FunctionType).Args, types[v]),
				types[function.Name].(typing.FunctionType).Return,
			}
		}

		body := construct(function.Body)

		newFunctions = append(newFunctions,
			ir.Function{function.Name, args, body.UpdateNames(mapping)})
	}

	return constructed, newFunctions, types
}
