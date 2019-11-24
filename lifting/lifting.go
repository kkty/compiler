package lifting

import (
	"fmt"
	"log"

	"github.com/kkty/mincaml-go/ir"
	"github.com/kkty/mincaml-go/mir"
	"github.com/kkty/mincaml-go/typing"
)

func findApplications(node ir.Node) []*ir.Application {
	applications := []*ir.Application{}

	queue := []ir.Node{node}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		switch node.(type) {
		case *ir.IfEqual:
			n := node.(*ir.IfEqual)
			queue = append(queue, n.True, n.False)
		case *ir.IfLessThan:
			n := node.(*ir.IfLessThan)
			queue = append(queue, n.True, n.False)
		case *ir.ValueBinding:
			n := node.(*ir.ValueBinding)
			queue = append(queue, n.Value, n.Next)
		case *ir.Application:
			applications = append(applications, node.(*ir.Application))
		}
	}

	return applications
}

// Lift separates function definitions and the main program.
// Functions (and function applications) are modified so that they do not have free variables.
func Lift(
	root mir.Node,
	types map[string]typing.Type,
) (ir.Node, []*ir.Function, map[string]typing.Type) {
	nextTemporaryId := 0
	temporary := func() string {
		defer func() { nextTemporaryId++ }()
		return fmt.Sprintf("_lifting_%d", nextTemporaryId)
	}

	functions := map[string]*ir.Function{}

	// Constructs ir.Node from mir.Node.
	// Function bindings are removed and function applications are modified.
	var construct func(node mir.Node) ir.Node
	construct = func(node mir.Node) ir.Node {
		switch node.(type) {
		case *mir.Variable:
			return &ir.Variable{node.(*mir.Variable).Name}
		case *mir.Unit:
			return &ir.Unit{}
		case *mir.Int:
			return &ir.Int{node.(*mir.Int).Value}
		case *mir.Bool:
			return &ir.Bool{node.(*mir.Bool).Value}
		case *mir.Float:
			return &ir.Float{node.(*mir.Float).Value}
		case *mir.Add:
			n := node.(*mir.Add)
			return &ir.Add{n.Left, n.Right}
		case *mir.Sub:
			n := node.(*mir.Sub)
			return &ir.Sub{n.Left, n.Right}
		case *mir.FloatAdd:
			n := node.(*mir.FloatAdd)
			return &ir.FloatAdd{n.Left, n.Right}
		case *mir.FloatSub:
			n := node.(*mir.FloatSub)
			return &ir.FloatSub{n.Left, n.Right}
		case *mir.FloatDiv:
			n := node.(*mir.FloatDiv)
			return &ir.FloatDiv{n.Left, n.Right}
		case *mir.FloatMul:
			n := node.(*mir.FloatMul)
			return &ir.FloatMul{n.Left, n.Right}
		case *mir.IfEqual:
			n := node.(*mir.IfEqual)
			return &ir.IfEqual{n.Left, n.Right, construct(n.True), construct(n.False)}
		case *mir.IfLessThan:
			n := node.(*mir.IfLessThan)
			return &ir.IfLessThan{n.Left, n.Right, construct(n.True), construct(n.False)}
		case *mir.ValueBinding:
			n := node.(*mir.ValueBinding)
			return &ir.ValueBinding{n.Name, construct(n.Value), construct(n.Next)}
		case *mir.FunctionBinding:
			n := node.(*mir.FunctionBinding)
			functions[n.Name] = &ir.Function{n.Name, n.Args, construct(n.Body)}
			return construct(n.Next)
		case *mir.Application:
			n := node.(*mir.Application)
			return &ir.Application{n.Function, n.Args}
		case *mir.Tuple:
			return &ir.Tuple{node.(*mir.Tuple).Elements}
		case *mir.TupleBinding:
			n := node.(*mir.TupleBinding)
			ret := construct(n.Next)
			for i, name := range n.Names {
				ret = &ir.ValueBinding{name, &ir.TupleGet{n.Tuple, int32(i)}, ret}
			}
			return ret
		case *mir.ArrayCreate:
			n := node.(*mir.ArrayCreate)
			return &ir.ArrayCreate{n.Size, n.Value}
		case *mir.ArrayGet:
			n := node.(*mir.ArrayGet)
			return &ir.ArrayGet{n.Array, n.Index}
		case *mir.ArrayPut:
			n := node.(*mir.ArrayPut)
			return &ir.ArrayPut{n.Array, n.Index, n.Value}
		case *mir.ReadInt:
			return &ir.ReadInt{}
		case *mir.ReadFloat:
			return &ir.ReadFloat{}
		case *mir.PrintInt:
			return &ir.PrintInt{node.(*mir.PrintInt).Arg}
		case *mir.PrintChar:
			return &ir.PrintChar{node.(*mir.PrintChar).Arg}
		case *mir.IntToFloat:
			return &ir.IntToFloat{node.(*mir.IntToFloat).Arg}
		case *mir.FloatToInt:
			return &ir.FloatToInt{node.(*mir.FloatToInt).Arg}
		case *mir.Sqrt:
			return &ir.Sqrt{node.(*mir.Sqrt).Arg}
		case *mir.Neg:
			n := node.(*mir.Neg)
			zero := temporary()

			if types[n.Arg] == typing.IntType {
				types[zero] = typing.IntType
				return &ir.ValueBinding{zero, &ir.Int{0},
					&ir.Sub{zero, n.Arg}}
			}

			types[zero] = typing.FloatType
			return &ir.ValueBinding{zero, &ir.Float{0},
				&ir.FloatSub{zero, n.Arg}}
		}

		log.Fatal("invalid mir node")
		return nil
	}

	constructed := construct(root)

	functionToApplications := map[string][]*ir.Application{}

	for _, function := range functions {
		functionToApplications[function.Name] = findApplications(function.Body)
	}

	applicationsInMain := findApplications(constructed)

	nextVarId := 0
	newVar := func() string {
		defer func() { nextVarId++ }()
		return fmt.Sprintf("_lifting_%d", nextVarId)
	}

	for {
		for _, function := range functions {
			freeVariables := function.FreeVariables()

			// Adds arguments to the applications.
			functionToApplications["main"] = applicationsInMain
			for _, applications := range functionToApplications {
				for _, application := range applications {
					if application.Function == function.Name {
						application.Args = append(application.Args, freeVariables...)
					}
				}
			}
			delete(functionToApplications, "main")

			// Adds arguments to the function.
			mapping := map[string]string{}
			for _, freeVariable := range freeVariables {
				v := newVar()
				mapping[freeVariable] = v
				types[v] = types[freeVariable]
				function.Args = append(function.Args, v)
				types[function.Name] = typing.FunctionType{
					append(types[function.Name].(typing.FunctionType).Args, types[v]),
					types[function.Name].(typing.FunctionType).Return,
				}
			}

			function.Body.UpdateNames(mapping)
		}

		ok := true
		for _, function := range functions {
			if len(function.FreeVariables()) > 0 {
				ok = false
			}
		}

		if ok {
			break
		}
	}

	functionsAsSlice := []*ir.Function{}
	for _, function := range functions {
		functionsAsSlice = append(functionsAsSlice, function)
	}

	return constructed, functionsAsSlice, types
}
