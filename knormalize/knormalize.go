package knormalize

import (
	"fmt"
	"log"

	"github.com/kkty/mincaml-go/ast"
	"github.com/kkty/mincaml-go/mir"
)

var nextTemporaryId = 1

func temporary() string {
	defer func() { nextTemporaryId++ }()
	return fmt.Sprintf("_%d", nextTemporaryId)
}

func KNormalize(node ast.Node) mir.Node {
	switch node.(type) {
	case ast.Variable:
		return mir.Variable{node.(ast.Variable).Name}
	case ast.Unit:
		return mir.Unit{}
	case ast.Int:
		return mir.Int{node.(ast.Int).Value}
	case ast.Bool:
		return mir.Bool{node.(ast.Bool).Value}
	case ast.Float:
		return mir.Float{node.(ast.Float).Value}
	case ast.Add:
		n := node.(ast.Add)

		left := temporary()
		right := temporary()

		return mir.ValueBinding{left, KNormalize(n.Left),
			mir.ValueBinding{right, KNormalize(n.Right), mir.Add{left, right}}}
	case ast.Sub:
		n := node.(ast.Sub)

		left := temporary()
		right := temporary()

		return mir.ValueBinding{left, KNormalize(n.Left),
			mir.ValueBinding{right, KNormalize(n.Right), mir.Sub{left, right}}}
	case ast.FloatAdd:
		n := node.(ast.FloatAdd)

		left := temporary()
		right := temporary()

		return mir.ValueBinding{left, KNormalize(n.Left),
			mir.ValueBinding{right, KNormalize(n.Right), mir.FloatAdd{left, right}}}
	case ast.FloatSub:
		n := node.(ast.FloatSub)

		left := temporary()
		right := temporary()

		return mir.ValueBinding{left, KNormalize(n.Left),
			mir.ValueBinding{right, KNormalize(n.Right), mir.FloatSub{left, right}}}
	case ast.FloatDiv:
		n := node.(ast.FloatDiv)

		left := temporary()
		right := temporary()

		return mir.ValueBinding{left, KNormalize(n.Left),
			mir.ValueBinding{right, KNormalize(n.Right), mir.FloatDiv{left, right}}}
	case ast.FloatMul:
		n := node.(ast.FloatMul)

		left := temporary()
		right := temporary()

		return mir.ValueBinding{left, KNormalize(n.Left),
			mir.ValueBinding{right, KNormalize(n.Right), mir.FloatMul{left, right}}}
	case ast.Equal:
		n := node.(ast.Equal)

		left := temporary()
		right := temporary()

		return mir.ValueBinding{left, KNormalize(n.Left),
			mir.ValueBinding{right, KNormalize(n.Right),
				mir.IfEqual{left, right, mir.Bool{true}, mir.Bool{false}}}}
	case ast.LessThanOrEqual:
		n := node.(ast.LessThanOrEqual)

		left := temporary()
		right := temporary()

		return mir.ValueBinding{left, KNormalize(n.Left),
			mir.ValueBinding{right, KNormalize(n.Right),
				mir.IfLessThanOrEqual{left, right, mir.Bool{true}, mir.Bool{false}}}}
	case ast.Neg:
		n := node.(ast.Neg)

		left := temporary()
		right := temporary()

		return mir.ValueBinding{left, mir.Int{0},
			mir.ValueBinding{right, KNormalize(n.Inner), mir.Sub{left, right}}}
	case ast.FloatNeg:
		n := node.(ast.FloatNeg)

		left := temporary()
		right := temporary()

		return mir.ValueBinding{left, mir.Float{0},
			mir.ValueBinding{right, KNormalize(n.Inner), mir.FloatSub{left, right}}}
	case ast.Not:
		n := node.(ast.Not)

		left := temporary()
		right := temporary()

		return mir.ValueBinding{left, KNormalize(n.Inner),
			mir.ValueBinding{right, mir.Bool{true},
				mir.IfEqual{left, right, mir.Bool{false}, mir.Bool{true}}}}
	case ast.If:
		n := node.(ast.If)

		left := temporary()
		right := temporary()

		return mir.ValueBinding{left, KNormalize(n.Condition),
			mir.ValueBinding{right, mir.Bool{true},
				mir.IfEqual{left, right, KNormalize(n.True), KNormalize(n.False)}}}
	case ast.ValueBinding:
		n := node.(ast.ValueBinding)

		return mir.ValueBinding{n.Name, KNormalize(n.Body), KNormalize(n.Next)}
	case ast.FunctionBinding:
		n := node.(ast.FunctionBinding)

		return mir.FunctionBinding{n.Name, n.Args, KNormalize(n.Body), KNormalize(n.Next)}
	case ast.Application:
		n := node.(ast.Application)

		args := []string{}
		for _ = range n.Args {
			args = append(args, temporary())
		}

		var ret mir.Node = mir.Application{n.Function, args}
		for i := len(n.Args) - 1; i >= 0; i-- {
			ret = mir.ValueBinding{args[i], KNormalize(n.Args[i]), ret}
		}

		return ret
	case ast.Tuple:
		n := node.(ast.Tuple)

		elements := []string{}
		for i := 0; i < len(n.Elements); i++ {
			elements = append(elements, temporary())
		}

		var ret mir.Node = mir.Tuple{elements}
		for i := len(n.Elements) - 1; i >= 0; i-- {
			ret = mir.ValueBinding{elements[i], KNormalize(n.Elements[i]), ret}
		}

		return ret
	case ast.TupleBinding:
		n := node.(ast.TupleBinding)

		tuple := temporary()

		return mir.ValueBinding{tuple, KNormalize(n.Tuple),
			mir.TupleBinding{n.Names, tuple, KNormalize(n.Next)}}
	case ast.ArrayCreate:
		n := node.(ast.ArrayCreate)

		size := temporary()
		value := temporary()

		return mir.ValueBinding{size, KNormalize(n.Size),
			mir.ValueBinding{value, KNormalize(n.Value),
				mir.ArrayCreate{size, value}}}
	case ast.ArrayGet:
		n := node.(ast.ArrayGet)

		array := temporary()
		index := temporary()

		return mir.ValueBinding{array, KNormalize(n.Array),
			mir.ValueBinding{index, KNormalize(n.Index),
				mir.ArrayGet{array, index}}}
	case ast.ArrayPut:
		n := node.(ast.ArrayPut)

		array := temporary()
		index := temporary()
		value := temporary()

		return mir.ValueBinding{array, KNormalize(n.Array),
			mir.ValueBinding{index, KNormalize(n.Index),
				mir.ValueBinding{value, KNormalize(n.Value),
					mir.ArrayPut{array, index, value}}}}
	case ast.ReadInt:
		return mir.ReadInt{}
	case ast.ReadFloat:
		return mir.ReadFloat{}
	case ast.PrintInt:
		n := node.(ast.PrintInt)
		arg := temporary()
		return mir.ValueBinding{arg, KNormalize(n.Inner),
			mir.PrintInt{arg}}
	case ast.PrintChar:
		n := node.(ast.PrintChar)
		arg := temporary()
		return mir.ValueBinding{arg, KNormalize(n.Inner),
			mir.PrintChar{arg}}
	default:
		log.Fatal("invalid ast node")
	}

	return nil
}
