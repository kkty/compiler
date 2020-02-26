package ir

import (
	"fmt"

	"github.com/kkty/compiler/ast"
	"github.com/kkty/compiler/stringset"
	"github.com/kkty/compiler/typing"
)

// Generate generates Node from ast.Node.
// K-normalization is performed and functions are separated from the main program.
// Functions (and function applications) are modified so that they do not have free variables.
// Global variables are separated from the main program.
func Generate(root ast.Node, nameToType map[string]typing.Type) (Node, []*Function, map[string]Node, map[string]typing.Type) {
	functions := map[string]*Function{}

	nextNameId := 0
	newName := func() string {
		defer func() { nextNameId++ }()
		return fmt.Sprintf("_irgen_%d", nextNameId)
	}

	globals := map[string]Node{}

	// construct node recursively
	var construct func(node ast.Node) Node
	construct = func(node ast.Node) Node {
		// for K-normalization
		insert := func(nodes []ast.Node, getNext func([]string) Node) Node {
			names := []string{}
			for _, node := range nodes {
				if v, ok := node.(*ast.Variable); ok {
					names = append(names, v.Name)
				} else {
					name := newName()
					names = append(names, name)
					nameToType[name] = node.GetType(nameToType)
				}
			}
			ret := getNext(names)
			for i, node := range nodes {
				if _, ok := node.(*ast.Variable); !ok {
					ret = &Assignment{
						Name:  names[i],
						Value: construct(node),
						Next:  ret,
					}
				}
			}
			return ret
		}

		switch node := node.(type) {
		case *ast.Variable:
			return &Variable{Name: node.Name}
		case *ast.Unit:
			return &Unit{}
		case *ast.Int:
			return &Int{Value: node.Value}
		case *ast.Float:
			return &Float{Value: node.Value}
		case *ast.Bool:
			return &Bool{Value: node.Value}
		case *ast.Add:
			return insert([]ast.Node{node.Left, node.Right}, func(names []string) Node {
				return &Add{Left: names[0], Right: names[1]}
			})
		case *ast.Sub:
			return insert([]ast.Node{node.Left, node.Right}, func(names []string) Node {
				return &Sub{Left: names[0], Right: names[1]}
			})
		case *ast.FloatAdd:
			return insert([]ast.Node{node.Left, node.Right}, func(names []string) Node {
				return &FloatAdd{Left: names[0], Right: names[1]}
			})
		case *ast.FloatSub:
			return insert([]ast.Node{node.Left, node.Right}, func(names []string) Node {
				return &FloatSub{Left: names[0], Right: names[1]}
			})
		case *ast.FloatDiv:
			return insert([]ast.Node{node.Left, node.Right}, func(names []string) Node {
				return &FloatDiv{Left: names[0], Right: names[1]}
			})
		case *ast.FloatMul:
			return insert([]ast.Node{node.Left, node.Right}, func(names []string) Node {
				return &FloatMul{Left: names[0], Right: names[1]}
			})
		case *ast.Equal:
			return insert([]ast.Node{node.Left, node.Right}, func(names []string) Node {
				return &Equal{Left: names[0], Right: names[1]}
			})
		case *ast.LessThan:
			if _, ok := node.Left.GetType(nameToType).(*typing.IntType); ok {
				return insert([]ast.Node{node.Left, node.Right}, func(names []string) Node {
					return &LessThan{Left: names[0], Right: names[1]}
				})
			} else {
				return insert([]ast.Node{node.Left, node.Right}, func(names []string) Node {
					return &LessThanFloat{Left: names[0], Right: names[1]}
				})
			}
		case *ast.Neg:
			if _, ok := node.GetType(nameToType).(*typing.IntType); ok {
				return insert([]ast.Node{&ast.Int{Value: 0}, node.Inner}, func(names []string) Node {
					return &Sub{Left: names[0], Right: names[1]}
				})
			} else {
				return insert([]ast.Node{&ast.Float{Value: 0}, node.Inner}, func(names []string) Node {
					return &FloatSub{Left: names[0], Right: names[1]}
				})
			}
		case *ast.FloatNeg:
			return insert([]ast.Node{&ast.Float{Value: 0}, node.Inner}, func(names []string) Node {
				return &FloatSub{Left: names[0], Right: names[1]}
			})
		case *ast.Not:
			return insert([]ast.Node{node.Inner}, func(names []string) Node {
				return &Not{Inner: names[0]}
			})
		case *ast.If:
			if c, ok := node.Condition.(*ast.LessThan); ok {
				if _, ok := c.Left.GetType(nameToType).(*typing.IntType); ok {
					return insert([]ast.Node{c.Left, c.Right}, func(names []string) Node {
						return &IfLessThan{Left: names[0], Right: names[1], True: construct(node.True), False: construct(node.False)}
					})
				} else {
					return insert([]ast.Node{c.Left, c.Right}, func(names []string) Node {
						return &IfLessThanFloat{Left: names[0], Right: names[1], True: construct(node.True), False: construct(node.False)}
					})
				}
			}
			if c, ok := node.Condition.(*ast.Equal); ok {
				return insert([]ast.Node{c.Left, c.Right}, func(names []string) Node {
					return &IfEqual{Left: names[0], Right: names[1], True: construct(node.True), False: construct(node.False)}
				})
			}
			if c, ok := node.Condition.(*ast.Not); ok {
				return construct(&ast.If{Condition: c.Inner, True: node.False, False: node.True})
			}
			return insert([]ast.Node{node.Condition, &ast.Bool{Value: true}}, func(names []string) Node {
				return &IfEqual{Left: names[0], Right: names[1], True: construct(node.True), False: construct(node.False)}
			})
		case *ast.Assignment:
			return &Assignment{Name: node.Name, Value: construct(node.Body), Next: construct(node.Next)}
		case *ast.FunctionAssignment:
			// TODO: this might better be in parser
			args := node.Args
			if len(args) == 1 {
				if _, ok := nameToType[args[0]].(*typing.UnitType); ok {
					args = []string{}
				}
			}
			functions[node.Name] = &Function{Name: node.Name, Args: args, Body: construct(node.Body)}
			return construct(node.Next)
		case *ast.Application:
			// TODO: this might better be in parser
			if len(node.Args) == 1 {
				if _, ok := node.Args[0].GetType(nameToType).(*typing.UnitType); ok {
					return &Application{Function: node.Function, Args: nil}
				}
			}
			return insert(node.Args, func(names []string) Node {
				return &Application{Function: node.Function, Args: names}
			})
		case *ast.Tuple:
			return insert(node.Elements, func(names []string) Node {
				return &Tuple{Elements: names}
			})
		case *ast.TupleAssignment:
			return insert([]ast.Node{node.Tuple}, func(names []string) Node {
				tupleName := names[0]
				var ret Node = construct(node.Next)
				for i, name := range node.Names {
					ret = &Assignment{
						Name: name,
						Value: &TupleGet{
							Tuple: tupleName,
							Index: int32(i),
						},
						Next: ret,
					}
				}
				return ret
			})
		case *ast.ArrayCreate:
			return insert([]ast.Node{node.Size, node.Value}, func(names []string) Node {
				return &ArrayCreate{Length: names[0], Value: names[1]}
			})
		case *ast.ArrayGet:
			return insert([]ast.Node{node.Array, node.Index}, func(names []string) Node {
				return &ArrayGet{Array: names[0], Index: names[1]}
			})
		case *ast.ArrayPut:
			return insert([]ast.Node{node.Array, node.Index, node.Value}, func(names []string) Node {
				return &ArrayPut{Array: names[0], Index: names[1], Value: names[2]}
			})
		case *ast.ReadInt:
			return &ReadInt{}
		case *ast.ReadFloat:
			return &ReadFloat{}
		case *ast.WriteByte:
			return insert([]ast.Node{node.Inner}, func(names []string) Node {
				return &WriteByte{Arg: names[0]}
			})
		case *ast.IntToFloat:
			return insert([]ast.Node{node.Inner}, func(names []string) Node {
				return &IntToFloat{Arg: names[0]}
			})
		case *ast.FloatToInt:
			return insert([]ast.Node{node.Inner}, func(names []string) Node {
				return &FloatToInt{Arg: names[0]}
			})
		case *ast.Sqrt:
			return insert([]ast.Node{node.Inner}, func(names []string) Node {
				return &Sqrt{Arg: names[0]}
			})
		default:
			panic("invalid node")
		}
	}

	// If the program starts with an array assignment, it is considered a global variable assignment
	// and is removed from the program. This is repeated until the condition is met.
	for func() bool {
		if n, ok := root.(*ast.Assignment); ok {
			if _, ok := n.Body.GetType(nameToType).(*typing.ArrayType); ok {
				return true
			}
		}
		return false
	}() {
		n := root.(*ast.Assignment)
		globals[n.Name] = construct(n.Body)
		root = n.Next
	}

	constructed := construct(root)

	functionToApplications := map[string][]*Application{}

	for _, function := range functions {
		functionToApplications[function.Name] = function.Body.Applications()
	}

	// names of global variables
	globalNames := stringset.New()
	for n := range globals {
		globalNames.Add(n)
	}

	applicationsInMain := constructed.Applications()

	appended := stringset.New()

	// removes free variables in functions (lambda lifting)
	for {
		for _, function := range functions {
			freeVariables := []string{}
			for freeVariable := range function.FreeVariables() {
				if !globalNames.Has(freeVariable) {
					freeVariables = append(freeVariables, freeVariable)
				}
			}

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
				appended.Add(freeVariable)
				function.Args = append(function.Args, freeVariable)
				nameToType[function.Name] = &typing.FunctionType{
					append(nameToType[function.Name].(*typing.FunctionType).Args, nameToType[freeVariable]),
					nameToType[function.Name].(*typing.FunctionType).Return}
			}
		}

		// Ends if all the free variables are removed from every function.

		ok := true
		for _, function := range functions {
			for freeVariable := range function.FreeVariables() {
				if !globalNames.Has(freeVariable) {
					ok = false
				}
			}
		}

		if ok {
			// Updates functions so that they do not share the same names.
			for _, function := range functions {
				mapping := map[string]string{}
				for i, arg := range function.Args {
					if appended.Has(arg) {
						v := newName()
						mapping[arg] = v
						nameToType[v] = nameToType[arg]
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

	return constructed, functionsAsSlice, globals, nameToType
}
