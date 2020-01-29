package ir

import (
	"fmt"
	"github.com/kkty/compiler/ast"
	"github.com/kkty/compiler/typing"
	"github.com/thoas/go-funk"
)

// Generate generates Node from ast.Node.
// K-normalization is performed and functions are separated from the main program.
// Also, functions (and function applications) are modified so that they
// do not have free variables.
func Generate(root ast.Node, nameToType map[string]typing.Type) (Node, []*Function, map[string]typing.Type) {
	functions := map[string]*Function{}

	nextNameId := 0
	newName := func() string {
		defer func() { nextNameId++ }()
		return fmt.Sprintf("_irgen_%d", nextNameId)
	}

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

		switch node.(type) {
		case *ast.Variable:
			return &Variable{Name: node.(*ast.Variable).Name}
		case *ast.Unit:
			return &Unit{}
		case *ast.Int:
			return &Int{Value: node.(*ast.Int).Value}
		case *ast.Float:
			return &Float{Value: node.(*ast.Float).Value}
		case *ast.Bool:
			return &Bool{Value: node.(*ast.Bool).Value}
		case *ast.Add:
			n := node.(*ast.Add)
			return insert([]ast.Node{n.Left, n.Right}, func(names []string) Node {
				return &Add{Left: names[0], Right: names[1]}
			})
		case *ast.Sub:
			n := node.(*ast.Sub)
			return insert([]ast.Node{n.Left, n.Right}, func(names []string) Node {
				return &Sub{Left: names[0], Right: names[1]}
			})
		case *ast.FloatAdd:
			n := node.(*ast.FloatAdd)
			return insert([]ast.Node{n.Left, n.Right}, func(names []string) Node {
				return &FloatAdd{Left: names[0], Right: names[1]}
			})
		case *ast.FloatSub:
			n := node.(*ast.FloatSub)
			return insert([]ast.Node{n.Left, n.Right}, func(names []string) Node {
				return &FloatSub{Left: names[0], Right: names[1]}
			})
		case *ast.FloatDiv:
			n := node.(*ast.FloatDiv)
			return insert([]ast.Node{n.Left, n.Right}, func(names []string) Node {
				return &FloatDiv{Left: names[0], Right: names[1]}
			})
		case *ast.FloatMul:
			n := node.(*ast.FloatMul)
			return insert([]ast.Node{n.Left, n.Right}, func(names []string) Node {
				return &FloatMul{Left: names[0], Right: names[1]}
			})
		case *ast.Equal:
			n := node.(*ast.Equal)
			return insert([]ast.Node{n.Left, n.Right}, func(names []string) Node {
				return &Equal{Left: names[0], Right: names[1]}
			})
		case *ast.LessThan:
			n := node.(*ast.LessThan)
			return insert([]ast.Node{n.Left, n.Right}, func(names []string) Node {
				return &LessThan{Left: names[0], Right: names[1]}
			})
		case *ast.Neg:
			n := node.(*ast.Neg)
			if _, ok := n.GetType(nameToType).(*typing.IntType); ok {
				return insert([]ast.Node{&ast.Int{Value: 0}, n.Inner}, func(names []string) Node {
					return &Sub{Left: names[0], Right: names[1]}
				})
			} else {
				return insert([]ast.Node{&ast.Float{Value: 0}, n.Inner}, func(names []string) Node {
					return &FloatSub{Left: names[0], Right: names[1]}
				})
			}
		case *ast.FloatNeg:
			n := node.(*ast.FloatNeg)
			return insert([]ast.Node{&ast.Float{Value: 0}, n.Inner}, func(names []string) Node {
				return &FloatSub{Left: names[0], Right: names[1]}
			})
		case *ast.Not:
			n := node.(*ast.Not)
			return insert([]ast.Node{n.Inner}, func(names []string) Node {
				return &Not{Inner: names[0]}
			})
		case *ast.If:
			n := node.(*ast.If)
			if c, ok := n.Condition.(*ast.LessThan); ok {
				return insert([]ast.Node{c.Left, c.Right}, func(names []string) Node {
					return &IfLessThan{Left: names[0], Right: names[1], True: construct(n.True), False: construct(n.False)}
				})
			}
			if c, ok := n.Condition.(*ast.Equal); ok {
				return insert([]ast.Node{c.Left, c.Right}, func(names []string) Node {
					return &IfEqual{Left: names[0], Right: names[1], True: construct(n.True), False: construct(n.False)}
				})
			}
			if c, ok := n.Condition.(*ast.Not); ok {
				return construct(&ast.If{Condition: c.Inner, True: n.False, False: n.True})
			}
			return insert([]ast.Node{n.Condition, &ast.Bool{Value: true}}, func(names []string) Node {
				return &IfEqual{Left: names[0], Right: names[1], True: construct(n.True), False: construct(n.False)}
			})
		case *ast.Assignment:
			n := node.(*ast.Assignment)
			return &Assignment{Name: n.Name, Value: construct(n.Body), Next: construct(n.Next)}
		case *ast.FunctionAssignment:
			n := node.(*ast.FunctionAssignment)
			functions[n.Name] = &Function{Name: n.Name, Args: n.Args, Body: construct(n.Body)}
			return construct(n.Next)
		case *ast.Application:
			n := node.(*ast.Application)
			return insert(n.Args, func(names []string) Node {
				return &Application{Function: n.Function, Args: names}
			})
		case *ast.Tuple:
			n := node.(*ast.Tuple)
			return insert(n.Elements, func(names []string) Node {
				return &Tuple{Elements: names}
			})
		case *ast.TupleAssignment:
			n := node.(*ast.TupleAssignment)
			return insert([]ast.Node{n.Tuple}, func(names []string) Node {
				tupleName := names[0]
				var ret Node = construct(n.Next)
				for i, name := range n.Names {
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
			n := node.(*ast.ArrayCreate)
			return insert([]ast.Node{n.Size, n.Value}, func(names []string) Node {
				return &ArrayCreate{Length: names[0], Value: names[1]}
			})
		case *ast.ArrayGet:
			n := node.(*ast.ArrayGet)
			return insert([]ast.Node{n.Array, n.Index}, func(names []string) Node {
				return &ArrayGet{Array: names[0], Index: names[1]}
			})
		case *ast.ArrayPut:
			n := node.(*ast.ArrayPut)
			return insert([]ast.Node{n.Array, n.Index, n.Value}, func(names []string) Node {
				return &ArrayPut{Array: names[0], Index: names[1], Value: names[2]}
			})
		case *ast.ReadInt:
			return &ReadInt{}
		case *ast.ReadFloat:
			return &ReadFloat{}
		case *ast.WriteByte:
			n := node.(*ast.WriteByte)
			return insert([]ast.Node{n.Inner}, func(names []string) Node {
				return &WriteByte{Arg: names[0]}
			})
		case *ast.IntToFloat:
			n := node.(*ast.IntToFloat)
			return insert([]ast.Node{n.Inner}, func(names []string) Node {
				return &IntToFloat{Arg: names[0]}
			})
		case *ast.FloatToInt:
			n := node.(*ast.FloatToInt)
			return insert([]ast.Node{n.Inner}, func(names []string) Node {
				return &FloatToInt{Arg: names[0]}
			})
		case *ast.Sqrt:
			n := node.(*ast.Sqrt)
			return insert([]ast.Node{n.Inner}, func(names []string) Node {
				return &Sqrt{Arg: names[0]}
			})
		default:
			panic("invalid node")
		}
	}

	constructed := construct(root)

	functionToApplications := map[string][]*Application{}

	for _, function := range functions {
		functionToApplications[function.Name] = function.Body.Applications()
	}

	applicationsInMain := constructed.Applications()

	appended := map[string]struct{}{}

	// removes free variables in functions (lambda lifting)
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
				nameToType[function.Name] = &typing.FunctionType{
					append(nameToType[function.Name].(*typing.FunctionType).Args, nameToType[freeVariable]),
					nameToType[function.Name].(*typing.FunctionType).Return}
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

	return constructed, functionsAsSlice, nameToType
}
