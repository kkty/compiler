package mir

import (
	"fmt"
	"log"

	"github.com/kkty/compiler/ast"
)

var nextTemporaryId = 0

func temporary() string {
	defer func() { nextTemporaryId++ }()
	return fmt.Sprintf("_knormalize_%d", nextTemporaryId)
}

func Generate(node ast.Node) Node {
	insertTemporaries := func(
		nodes []ast.Node,
		constructor func(names []string) Node,
	) Node {
		names := []string{}
		bindings := map[string]Node{}
		for _, node := range nodes {
			switch node.(type) {
			case *ast.Variable:
				names = append(names, node.(*ast.Variable).Name)
			default:
				t := temporary()
				names = append(names, t)
				bindings[t] = Generate(node)
			}
		}

		ret := constructor(names)
		for _, name := range names {
			if value, ok := bindings[name]; ok {
				ret = &ValueBinding{name, value, ret}
			}
		}

		return ret
	}

	switch node.(type) {
	case *ast.Variable:
		return &Variable{node.(*ast.Variable).Name}
	case *ast.Unit:
		return &Unit{}
	case *ast.Int:
		return &Int{node.(*ast.Int).Value}
	case *ast.Bool:
		return &Bool{node.(*ast.Bool).Value}
	case *ast.Float:
		return &Float{node.(*ast.Float).Value}
	case *ast.Add:
		n := node.(*ast.Add)
		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) Node { return &Add{names[0], names[1]} })
	case *ast.Sub:
		n := node.(*ast.Sub)
		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) Node { return &Sub{names[0], names[1]} })
	case *ast.FloatAdd:
		n := node.(*ast.FloatAdd)
		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) Node { return &FloatAdd{names[0], names[1]} })
	case *ast.FloatSub:
		n := node.(*ast.FloatSub)
		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) Node { return &FloatSub{names[0], names[1]} })
	case *ast.FloatDiv:
		n := node.(*ast.FloatDiv)
		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) Node { return &FloatDiv{names[0], names[1]} })
	case *ast.FloatMul:
		n := node.(*ast.FloatMul)
		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) Node { return &FloatMul{names[0], names[1]} })
	case *ast.Equal:
		n := node.(*ast.Equal)

		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) Node {
				return &Equal{names[0], names[1]}
			})
	case *ast.LessThan:
		n := node.(*ast.LessThan)

		return insertTemporaries([]ast.Node{n.Left, n.Right},
			func(names []string) Node {
				return &LessThan{names[0], names[1]}
			})
	case *ast.Neg:
		n := node.(*ast.Neg)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) Node { return &Neg{names[0]} })
	case *ast.FloatNeg:
		n := node.(*ast.FloatNeg)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) Node {
				t := temporary()
				return &ValueBinding{t, &Float{0},
					&FloatSub{t, names[0]}}
			})
	case *ast.Not:
		n := node.(*ast.Not)

		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) Node {
				return &Not{names[0]}
			})
	case *ast.If:
		n := node.(*ast.If)

		switch n.Condition.(type) {
		case *ast.Equal:
			c := n.Condition.(*ast.Equal)
			return insertTemporaries([]ast.Node{c.Left, c.Right},
				func(names []string) Node {
					return &IfEqual{names[0], names[1], Generate(n.True), Generate(n.False)}
				})
		case *ast.LessThan:
			c := n.Condition.(*ast.LessThan)
			return insertTemporaries([]ast.Node{c.Left, c.Right},
				func(names []string) Node {
					return &IfLessThan{names[0], names[1], Generate(n.True), Generate(n.False)}
				})
		case *ast.Not:
			c := n.Condition.(*ast.Not)
			return Generate(&ast.If{c.Inner, n.False, n.True})
		}

		return insertTemporaries([]ast.Node{n.Condition},
			func(names []string) Node {
				t := temporary()
				return &ValueBinding{t, &Bool{true},
					&IfEqual{t, names[0], Generate(n.True), Generate(n.False)}}
			})
	case *ast.ValueBinding:
		n := node.(*ast.ValueBinding)

		return &ValueBinding{n.Name, Generate(n.Body), Generate(n.Next)}
	case *ast.FunctionBinding:
		n := node.(*ast.FunctionBinding)

		return &FunctionBinding{n.Name, n.Args, Generate(n.Body), Generate(n.Next)}
	case *ast.Application:
		n := node.(*ast.Application)

		return insertTemporaries(n.Args,
			func(names []string) Node { return &Application{n.Function, names} })
	case *ast.Tuple:
		n := node.(*ast.Tuple)

		return insertTemporaries(n.Elements,
			func(names []string) Node {
				return &Tuple{names}
			})
	case *ast.TupleBinding:
		n := node.(*ast.TupleBinding)

		return insertTemporaries([]ast.Node{n.Tuple},
			func(names []string) Node {
				return &TupleBinding{n.Names, names[0], Generate(n.Next)}
			})
	case *ast.ArrayCreate:
		n := node.(*ast.ArrayCreate)

		return insertTemporaries([]ast.Node{n.Size, n.Value},
			func(names []string) Node {
				return &ArrayCreate{names[0], names[1]}
			})
	case *ast.ArrayGet:
		n := node.(*ast.ArrayGet)

		return insertTemporaries([]ast.Node{n.Array, n.Index},
			func(names []string) Node {
				return &ArrayGet{names[0], names[1]}
			})
	case *ast.ArrayPut:
		n := node.(*ast.ArrayPut)

		return insertTemporaries([]ast.Node{n.Array, n.Index, n.Value},
			func(names []string) Node {
				return &ArrayPut{names[0], names[1], names[2]}
			})
	case *ast.ReadInt:
		return &ReadInt{}
	case *ast.ReadFloat:
		return &ReadFloat{}
	case *ast.PrintInt:
		n := node.(*ast.PrintInt)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) Node { return &PrintInt{names[0]} })
	case *ast.WriteByte:
		n := node.(*ast.WriteByte)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) Node { return &WriteByte{names[0]} })
	case *ast.IntToFloat:
		n := node.(*ast.IntToFloat)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) Node { return &IntToFloat{names[0]} })
	case *ast.FloatToInt:
		n := node.(*ast.FloatToInt)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) Node { return &FloatToInt{names[0]} })
	case *ast.Sqrt:
		n := node.(*ast.Sqrt)
		return insertTemporaries([]ast.Node{n.Inner},
			func(names []string) Node { return &Sqrt{names[0]} })
	default:
		log.Fatal("invalid ast node")
	}

	return nil
}
