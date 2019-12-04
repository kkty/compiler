package ir

import (
	"fmt"
	"log"

	"github.com/kkty/compiler/mir"
	"github.com/kkty/compiler/typing"
	"github.com/thoas/go-funk"
)

// Generate separates function definitions and the main program.
// Functions (and function applications) are modified so that they do not have free variables.
func Generate(
	root mir.Node,
	types map[string]typing.Type,
) (Node, []*Function, map[string]typing.Type) {
	functions := map[string]*Function{}

	// Constructs Node from mir.Node.
	// Function bindings are removed and function applications are modified.
	var construct func(node mir.Node) Node
	construct = func(node mir.Node) Node {
		switch node.(type) {
		case *mir.Variable:
			return &Variable{node.(*mir.Variable).Name}
		case *mir.Unit:
			return &Unit{}
		case *mir.Int:
			return &Int{node.(*mir.Int).Value}
		case *mir.Bool:
			return &Bool{node.(*mir.Bool).Value}
		case *mir.Float:
			return &Float{node.(*mir.Float).Value}
		case *mir.Add:
			n := node.(*mir.Add)
			return &Add{n.Left, n.Right}
		case *mir.Sub:
			n := node.(*mir.Sub)
			return &Sub{n.Left, n.Right}
		case *mir.FloatAdd:
			n := node.(*mir.FloatAdd)
			return &FloatAdd{n.Left, n.Right}
		case *mir.FloatSub:
			n := node.(*mir.FloatSub)
			return &FloatSub{n.Left, n.Right}
		case *mir.FloatDiv:
			n := node.(*mir.FloatDiv)
			return &FloatDiv{n.Left, n.Right}
		case *mir.FloatMul:
			n := node.(*mir.FloatMul)
			return &FloatMul{n.Left, n.Right}
		case *mir.Not:
			n := node.(*mir.Not)
			return &Not{n.Arg}
		case *mir.Equal:
			n := node.(*mir.Equal)
			return &Equal{n.Left, n.Right}
		case *mir.LessThan:
			n := node.(*mir.LessThan)
			return &LessThan{n.Left, n.Right}
		case *mir.IfEqual:
			n := node.(*mir.IfEqual)
			return &IfEqual{n.Left, n.Right, construct(n.True), construct(n.False)}
		case *mir.IfLessThan:
			n := node.(*mir.IfLessThan)
			return &IfLessThan{n.Left, n.Right, construct(n.True), construct(n.False)}
		case *mir.ValueBinding:
			n := node.(*mir.ValueBinding)
			return &ValueBinding{n.Name, construct(n.Value), construct(n.Next)}
		case *mir.FunctionBinding:
			n := node.(*mir.FunctionBinding)
			functions[n.Name] = &Function{n.Name, n.Args, construct(n.Body)}
			return construct(n.Next)
		case *mir.Application:
			n := node.(*mir.Application)
			return &Application{n.Function, n.Args}
		case *mir.Tuple:
			return &Tuple{node.(*mir.Tuple).Elements}
		case *mir.TupleBinding:
			n := node.(*mir.TupleBinding)
			ret := construct(n.Next)
			for i, name := range n.Names {
				ret = &ValueBinding{name, &TupleGet{n.Tuple, int32(i)}, ret}
			}
			return ret
		case *mir.ArrayCreate:
			n := node.(*mir.ArrayCreate)
			return &ArrayCreate{n.Size, n.Value}
		case *mir.ArrayGet:
			n := node.(*mir.ArrayGet)
			return &ArrayGet{n.Array, n.Index}
		case *mir.ArrayPut:
			n := node.(*mir.ArrayPut)
			return &ArrayPut{n.Array, n.Index, n.Value}
		case *mir.ReadInt:
			return &ReadInt{}
		case *mir.ReadFloat:
			return &ReadFloat{}
		case *mir.PrintInt:
			return &PrintInt{node.(*mir.PrintInt).Arg}
		case *mir.PrintChar:
			return &PrintChar{node.(*mir.PrintChar).Arg}
		case *mir.IntToFloat:
			return &IntToFloat{node.(*mir.IntToFloat).Arg}
		case *mir.FloatToInt:
			return &FloatToInt{node.(*mir.FloatToInt).Arg}
		case *mir.Sqrt:
			return &Sqrt{node.(*mir.Sqrt).Arg}
		case *mir.Neg:
			n := node.(*mir.Neg)

			if types[n.Arg] == typing.IntType {
				return &SubFromZero{n.Arg}
			}

			return &FloatSubFromZero{n.Arg}
		}

		log.Fatal("invalid mir node")
		return nil
	}

	constructed := construct(root)

	functionToApplications := map[string][]*Application{}

	for _, function := range functions {
		functionToApplications[function.Name] = function.Body.Applications()
	}

	applicationsInMain := constructed.Applications()

	nextVarId := 0
	newVar := func() string {
		defer func() { nextVarId++ }()
		return fmt.Sprintf("_lifting_%d", nextVarId)
	}

	appended := map[string]struct{}{}

	for {
		for _, function := range functions {
			freeVariables := funk.Keys(function.FreeVariables()).([]string)

			functionToApplications["main"] = applicationsInMain
			for _, applications := range functionToApplications {
				for _, application := range applications {
					if application.Function == function.Name {
						application.Args = append(application.Args, freeVariables...)
					}
				}
			}
			delete(functionToApplications, "main")

			for _, freeVariable := range freeVariables {
				appended[freeVariable] = struct{}{}
				function.Args = append(function.Args, freeVariable)
				types[function.Name] = typing.FunctionType{
					append(types[function.Name].(typing.FunctionType).Args, types[freeVariable]),
					types[function.Name].(typing.FunctionType).Return}
			}
		}

		// Ends if all the free variables are removed from every function.

		ok := true
		for _, function := range functions {
			if len(function.FreeVariables()) > 0 {
				ok = false
			}
		}

		if ok {
			// Updates functions so that they do not share the same names.
			for _, function := range functions {
				mapping := map[string]string{}
				for i, arg := range function.Args {
					if _, shouldReplace := appended[arg]; shouldReplace {
						v := newVar()
						mapping[arg] = v
						types[v] = types[arg]
						function.Args[i] = v
					}
				}
				function.Body.UpdateNames(mapping)
			}
			break
		}
	}

	functionsAsSlice := []*Function{}
	for _, function := range functions {
		functionsAsSlice = append(functionsAsSlice, function)
	}

	return constructed, functionsAsSlice, types
}
