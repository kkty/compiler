package knormalize

import (
	"fmt"
	"log"

	"github.com/kkty/mincaml-go/ast"
	"github.com/kkty/mincaml-go/mir"
)

var nextTemporaryId = 0

func temporary() string {
	defer func() { nextTemporaryId++ }()
	return fmt.Sprintf("_knormalize_%d", nextTemporaryId)
}

func KNormalize(node ast.Node) mir.Node {
	insertTemporaries := func(
		nodes []ast.Node,
		constructor func(names []string) mir.Node,
	) mir.Node {
		names := []string{}
		bindings := map[string]mir.Node{}
		for _, node := range nodes {
			switch node.(type) {
			case *ast.Variable:
				names = append(names, node.(*ast.Variable).Name)
			default:
				t := temporary()
				names = append(names, t)
				bindings[t] = KNormalize(node)
			}
		}

		ret := constructor(names)
		for _, name := range names {
			if value, ok := bindings[name]; ok {
				ret = &mir.ValueBinding{name, value, ret}
			}
		}

		return ret
	}

	switch node.(type) {
	case *ast.Variable:
		return &mir.Variable{node.(*ast.Variable).Name}
	case *ast.Unit:
		return &mir.Unit{}
	case *ast.Int:
		return &mir.Int{node.(*ast.Int).Value}
	case *ast.Bool:
		return &mir.Bool{node.(*ast.Bool).Value}
	case *ast.Float:
		return &mir.Float{node.(*ast.Float).Value}
	case *ast.Add:
		n := node.(*ast.Add)
		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) mir.Node { return &mir.Add{names[0], names[1]} })
	case *ast.Sub:
		n := node.(*ast.Sub)
		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) mir.Node { return &mir.Sub{names[0], names[1]} })
	case *ast.FloatAdd:
		n := node.(*ast.FloatAdd)
		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) mir.Node { return &mir.FloatAdd{names[0], names[1]} })
	case *ast.FloatSub:
		n := node.(*ast.FloatSub)
		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) mir.Node { return &mir.FloatSub{names[0], names[1]} })
	case *ast.FloatDiv:
		n := node.(*ast.FloatDiv)
		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) mir.Node { return &mir.FloatDiv{names[0], names[1]} })
	case *ast.FloatMul:
		n := node.(*ast.FloatMul)
		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) mir.Node { return &mir.FloatMul{names[0], names[1]} })
	case *ast.Equal:
		n := node.(*ast.Equal)

		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) mir.Node {
				return &mir.IfEqual{names[0], names[1], &mir.Bool{true}, &mir.Bool{false}}
			})
	case *ast.LessThan:
		n := node.(*ast.LessThan)

		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) mir.Node {
				return &mir.IfLessThan{names[0], names[1], &mir.Bool{true}, &mir.Bool{false}}
			})
	case *ast.Neg:
		n := node.(*ast.Neg)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) mir.Node { return &mir.Neg{names[0]} })
	case *ast.FloatNeg:
		n := node.(*ast.FloatNeg)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) mir.Node {
				t := temporary()
				return &mir.ValueBinding{t, &mir.Float{0},
					&mir.FloatSub{t, names[0]}}
			})
	case *ast.Not:
		n := node.(*ast.Not)

		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) mir.Node {
				t := temporary()
				return &mir.ValueBinding{t, &mir.Bool{true},
					&mir.IfEqual{t, names[0], &mir.Bool{false}, &mir.Bool{true}}}
			})
	case *ast.If:
		n := node.(*ast.If)

		switch n.Condition.(type) {
		case *ast.Equal:
			c := n.Condition.(*ast.Equal)
			return insertTemporaries([]ast.Node{c.Left, c.Right},
				func(names []string) mir.Node {
					return &mir.IfEqual{names[0], names[1], KNormalize(n.True), KNormalize(n.False)}
				})
		case *ast.LessThan:
			c := n.Condition.(*ast.LessThan)
			return insertTemporaries([]ast.Node{c.Left, c.Right},
				func(names []string) mir.Node {
					return &mir.IfLessThan{names[0], names[1], KNormalize(n.True), KNormalize(n.False)}
				})
		case *ast.Not:
			c := n.Condition.(*ast.Not)
			return KNormalize(&ast.If{c.Inner, n.False, n.True})
		}

		return insertTemporaries([]ast.Node{n.Condition},
			func(names []string) mir.Node {
				t := temporary()
				return &mir.ValueBinding{t, &mir.Bool{true},
					&mir.IfEqual{t, names[0], KNormalize(n.True), KNormalize(n.False)}}
			})
	case *ast.ValueBinding:
		n := node.(*ast.ValueBinding)

		return &mir.ValueBinding{n.Name, KNormalize(n.Body), KNormalize(n.Next)}
	case *ast.FunctionBinding:
		n := node.(*ast.FunctionBinding)

		return &mir.FunctionBinding{n.Name, n.Args, KNormalize(n.Body), KNormalize(n.Next)}
	case *ast.Application:
		n := node.(*ast.Application)

		return insertTemporaries(n.Args,
			func(names []string) mir.Node { return &mir.Application{n.Function, names} })
	case *ast.Tuple:
		n := node.(*ast.Tuple)

		return insertTemporaries(n.Elements,
			func(names []string) mir.Node {
				return &mir.Tuple{names}
			})
	case *ast.TupleBinding:
		n := node.(*ast.TupleBinding)

		return insertTemporaries([]ast.Node{n.Tuple},
			func(names []string) mir.Node {
				return &mir.TupleBinding{n.Names, names[0], KNormalize(n.Next)}
			})
	case *ast.ArrayCreate:
		n := node.(*ast.ArrayCreate)

		return insertTemporaries([]ast.Node{n.Size, n.Value},
			func(names []string) mir.Node {
				return &mir.ArrayCreate{names[0], names[1]}
			})
	case *ast.ArrayGet:
		n := node.(*ast.ArrayGet)

		return insertTemporaries([]ast.Node{n.Array, n.Index},
			func(names []string) mir.Node {
				return &mir.ArrayGet{names[0], names[1]}
			})
	case *ast.ArrayPut:
		n := node.(*ast.ArrayPut)

		return insertTemporaries([]ast.Node{n.Array, n.Index, n.Value},
			func(names []string) mir.Node {
				return &mir.ArrayPut{names[0], names[1], names[2]}
			})
	case *ast.ReadInt:
		return &mir.ReadInt{}
	case *ast.ReadFloat:
		return &mir.ReadFloat{}
	case *ast.PrintInt:
		n := node.(*ast.PrintInt)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) mir.Node { return &mir.PrintInt{names[0]} })
	case *ast.PrintChar:
		n := node.(*ast.PrintChar)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) mir.Node { return &mir.PrintChar{names[0]} })
	case *ast.IntToFloat:
		n := node.(*ast.IntToFloat)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) mir.Node { return &mir.IntToFloat{names[0]} })
	case *ast.FloatToInt:
		n := node.(*ast.FloatToInt)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) mir.Node { return &mir.FloatToInt{names[0]} })
	case *ast.Sqrt:
		n := node.(*ast.Sqrt)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) mir.Node { return &mir.Sqrt{names[0]} })
	default:
		log.Fatal("invalid ast node")
	}

	return nil
}
