package ir

import (
	"fmt"
	"github.com/emicklei/dot"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func GenerateGraph(main Node, functions []*Function) error {
	nextID := 0
	newID := func() string {
		defer func() { nextID++ }()
		return strconv.Itoa(nextID)
	}
	var generate func(Node, *dot.Graph) dot.Node
	generate = func(node Node, g *dot.Graph) dot.Node {
		switch node.(type) {
		case *Variable:
			n := node.(*Variable)
			return g.Node(newID()).Label(fmt.Sprintf("Variable(%s)", n.Name))
		case *Unit:
			return g.Node(newID()).Label("Unit")
		case *Int:
			return g.Node(newID()).Label(fmt.Sprintf("Int(%d)", node.(*Int).Value))
		case *Bool:
			return g.Node(newID()).Label(fmt.Sprintf("Bool(%v)", node.(*Bool).Value))
		case *Float:
			return g.Node(newID()).Label(fmt.Sprintf("Float(%v)", node.(*Float).Value))
		case *Add:
			n := node.(*Add)
			return g.Node(newID()).Label(fmt.Sprintf("Add(%v, %v)", n.Left, n.Right))
		case *AddImmediate:
			n := node.(*AddImmediate)
			return g.Node(newID()).Label(fmt.Sprintf("AddImmediate(%v, %v)", n.Left, n.Right))
		case *Sub:
			n := node.(*Sub)
			return g.Node(newID()).Label(fmt.Sprintf("Sub(%v, %v)", n.Left, n.Right))
		case *SubFromZero:
			n := node.(*SubFromZero)
			return g.Node(newID()).Label(fmt.Sprintf("SubFromZero(%v)", n.Inner))
		case *FloatAdd:
			n := node.(*FloatAdd)
			return g.Node(newID()).Label(fmt.Sprintf("FloatAdd(%v, %v)", n.Left, n.Right))
		case *FloatSub:
			n := node.(*FloatSub)
			return g.Node(newID()).Label(fmt.Sprintf("FloatSub(%v, %v)", n.Left, n.Right))
		case *FloatSubFromZero:
			n := node.(*FloatSubFromZero)
			return g.Node(newID()).Label(fmt.Sprintf("FloatSubFromZero(%v)", n.Inner))
		case *FloatDiv:
			n := node.(*FloatDiv)
			return g.Node(newID()).Label(fmt.Sprintf("FloatDiv(%v, %v)", n.Left, n.Right))
		case *FloatMul:
			n := node.(*FloatMul)
			return g.Node(newID()).Label(fmt.Sprintf("FloatMul(%v, %v)", n.Left, n.Right))
		case *Not:
			return g.Node(newID()).Label(fmt.Sprintf("Not(%v)", node.(*Not).Inner))
		case *Equal:
			n := node.(*Equal)
			return g.Node(newID()).Label(fmt.Sprintf("Equal(%v, %v)", n.Left, n.Right))
		case *LessThan:
			n := node.(*LessThan)
			return g.Node(newID()).Label(fmt.Sprintf("LessThan(%v, %v)", n.Left, n.Right))
		case *IfEqual:
			n := node.(*IfEqual)
			gn := g.Node(newID()).Label(fmt.Sprintf("IfEqual(%v, %v)", n.Left, n.Right))
			gn.Edge(generate(n.True, g), "True")
			gn.Edge(generate(n.False, g), "False")
			return gn
		case *IfEqualZero:
			n := node.(*IfEqualZero)
			gn := g.Node(newID()).Label(fmt.Sprintf("IfEqualZero(%v)", n.Inner))
			gn.Edge(generate(n.True, g), "True")
			gn.Edge(generate(n.False, g), "False")
			return gn
		case *IfEqualTrue:
			n := node.(*IfEqualTrue)
			gn := g.Node(newID()).Label(fmt.Sprintf("IfEqualTrue(%v)", n.Inner))
			gn.Edge(generate(n.True, g), "True")
			gn.Edge(generate(n.False, g), "False")
			return gn
		case *IfLessThan:
			n := node.(*IfLessThan)
			gn := g.Node(newID()).Label(fmt.Sprintf("IfLessThan(%v, %v)", n.Left, n.Right))
			gn.Edge(generate(n.True, g), "True")
			gn.Edge(generate(n.False, g), "False")
			return gn
		case *IfLessThanZero:
			n := node.(*IfLessThanZero)
			gn := g.Node(newID()).Label(fmt.Sprintf("IfLessThanZero(%v)", n.Inner))
			gn.Edge(generate(n.True, g), "True")
			gn.Edge(generate(n.False, g), "False")
			return gn
		case *ValueBinding:
			n := node.(*ValueBinding)
			gn := g.Node(newID()).Label(fmt.Sprintf("ValueBinding(%v)", n.Name))
			gn.Edge(generate(n.Value, g), "Value")
			gn.Edge(generate(n.Next, g), "Next")
			return gn
		case *Application:
			n := node.(*Application)
			return g.Node(newID()).Label(
				fmt.Sprintf("Application(%v, [%v])", n.Function,
					strings.Join(n.Args, ", ")))
		case *Tuple:
			return g.Node(newID()).Label(fmt.Sprintf("Tuple([%v])", strings.Join(node.(*Tuple).Elements, ", ")))
		case *TupleGet:
			n := node.(*TupleGet)
			return g.Node(newID()).Label(fmt.Sprintf("TupleGet(%v, %v)", n.Tuple, n.Index))
		case *ArrayCreate:
			n := node.(*ArrayCreate)
			return g.Node(newID()).Label(fmt.Sprintf("ArrayCreate(%v, %v)", n.Value, n.Length))
		case *ArrayCreateImmediate:
			n := node.(*ArrayCreateImmediate)
			return g.Node(newID()).Label(fmt.Sprintf("ArrayCreateImmediate(%v, %v)", n.Value, n.Length))
		case *ArrayGet:
			n := node.(*ArrayGet)
			return g.Node(newID()).Label(fmt.Sprintf("ArrayGet(%v, %v)", n.Array, n.Index))
		case *ArrayGetImmediate:
			n := node.(*ArrayGetImmediate)
			return g.Node(newID()).Label(fmt.Sprintf("ArrayGetImmediate(%v, %v)", n.Array, n.Index))
		case *ArrayPut:
			n := node.(*ArrayPut)
			return g.Node(newID()).Label(fmt.Sprintf("ArrayPut(%v, %v, %v", n.Array, n.Index, n.Value))
		case *ArrayPutImmediate:
			n := node.(*ArrayPutImmediate)
			return g.Node(newID()).Label(fmt.Sprintf("ArrayPutImmediate(%v, %v, %v", n.Array, n.Index, n.Value))
		case *ReadInt:
			return g.Node(newID()).Label("ReadInt")
		case *ReadFloat:
			return g.Node(newID()).Label("ReadFloat")
		case *PrintInt:
			return g.Node(newID()).Label("PrintInt")
		case *WriteByte:
			return g.Node(newID()).Label("WriteByte")
		case *IntToFloat:
			return g.Node(newID()).Label("IntToFloat")
		case *FloatToInt:
			return g.Node(newID()).Label("FloatToInt")
		case *Sqrt:
			return g.Node(newID()).Label("Sqrt")
		}

		panic("invalid node")
	}

	if err := os.MkdirAll("graphs", 0777); err != nil {
		return err
	}
	for _, function := range append(functions, &Function{
		Name: "main",
		Args: []string{},
		Body: main,
	}) {
		g := dot.NewGraph(dot.Directed)
		generate(function.Body, g)
		if err := ioutil.WriteFile(
			fmt.Sprintf("graphs/%s.dot", function.Name),
			[]byte(g.String()), 0666); err != nil {
			return err
		}
	}

	return nil
}
