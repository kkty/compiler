package ir

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/emicklei/dot"
)

func GenerateGraph(main Node, functions []*Function) error {
	nextID := 0
	newID := func() string {
		defer func() { nextID++ }()
		return strconv.Itoa(nextID)
	}
	var generate func(Node, *dot.Graph) dot.Node
	generate = func(node Node, g *dot.Graph) dot.Node {
		switch n := node.(type) {
		case *Variable:
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
			return g.Node(newID()).Label(fmt.Sprintf("Add(%v, %v)", n.Left, n.Right))
		case *AddImmediate:
			return g.Node(newID()).Label(fmt.Sprintf("AddImmediate(%v, %v)", n.Left, n.Right))
		case *Sub:
			return g.Node(newID()).Label(fmt.Sprintf("Sub(%v, %v)", n.Left, n.Right))
		case *SubFromZero:
			return g.Node(newID()).Label(fmt.Sprintf("SubFromZero(%v)", n.Inner))
		case *FloatAdd:
			return g.Node(newID()).Label(fmt.Sprintf("FloatAdd(%v, %v)", n.Left, n.Right))
		case *FloatSub:
			return g.Node(newID()).Label(fmt.Sprintf("FloatSub(%v, %v)", n.Left, n.Right))
		case *FloatSubFromZero:
			return g.Node(newID()).Label(fmt.Sprintf("FloatSubFromZero(%v)", n.Inner))
		case *FloatDiv:
			return g.Node(newID()).Label(fmt.Sprintf("FloatDiv(%v, %v)", n.Left, n.Right))
		case *FloatMul:
			return g.Node(newID()).Label(fmt.Sprintf("FloatMul(%v, %v)", n.Left, n.Right))
		case *Not:
			return g.Node(newID()).Label(fmt.Sprintf("Not(%v)", node.(*Not).Inner))
		case *Equal:
			return g.Node(newID()).Label(fmt.Sprintf("Equal(%v, %v)", n.Left, n.Right))
		case *LessThan:
			return g.Node(newID()).Label(fmt.Sprintf("LessThan(%v, %v)", n.Left, n.Right))
		case *LessThanFloat:
			return g.Node(newID()).Label(fmt.Sprintf("LessThanFloat(%v, %v)", n.Left, n.Right))
		case *IfEqual:
			gn := g.Node(newID()).Label(fmt.Sprintf("IfEqual(%v, %v)", n.Left, n.Right))
			gn.Edge(generate(n.True, g), "True")
			gn.Edge(generate(n.False, g), "False")
			return gn
		case *IfEqualZero:
			gn := g.Node(newID()).Label(fmt.Sprintf("IfEqualZero(%v)", n.Inner))
			gn.Edge(generate(n.True, g), "True")
			gn.Edge(generate(n.False, g), "False")
			return gn
		case *IfEqualTrue:
			gn := g.Node(newID()).Label(fmt.Sprintf("IfEqualTrue(%v)", n.Inner))
			gn.Edge(generate(n.True, g), "True")
			gn.Edge(generate(n.False, g), "False")
			return gn
		case *IfLessThan:
			gn := g.Node(newID()).Label(fmt.Sprintf("IfLessThan(%v, %v)", n.Left, n.Right))
			gn.Edge(generate(n.True, g), "True")
			gn.Edge(generate(n.False, g), "False")
			return gn
		case *IfLessThanFloat:
			gn := g.Node(newID()).Label(fmt.Sprintf("IfLessThanFloat(%v, %v)", n.Left, n.Right))
			gn.Edge(generate(n.True, g), "True")
			gn.Edge(generate(n.False, g), "False")
			return gn
		case *IfLessThanZero:
			gn := g.Node(newID()).Label(fmt.Sprintf("IfLessThanZero(%v)", n.Inner))
			gn.Edge(generate(n.True, g), "True")
			gn.Edge(generate(n.False, g), "False")
			return gn
		case *IfLessThanZeroFloat:
			gn := g.Node(newID()).Label(fmt.Sprintf("IfLessThanZeroFloat(%v)", n.Inner))
			gn.Edge(generate(n.True, g), "True")
			gn.Edge(generate(n.False, g), "False")
			return gn
		case *Assignment:
			gn := g.Node(newID()).Label(fmt.Sprintf("Assignment(%v)", n.Name))
			gn.Edge(generate(n.Value, g), "Value")
			gn.Edge(generate(n.Next, g), "Next")
			return gn
		case *Application:
			return g.Node(newID()).Label(
				fmt.Sprintf("Application(%v, [%v])", n.Function,
					strings.Join(n.Args, ", ")))
		case *Tuple:
			return g.Node(newID()).Label(fmt.Sprintf("Tuple([%v])", strings.Join(node.(*Tuple).Elements, ", ")))
		case *TupleGet:
			return g.Node(newID()).Label(fmt.Sprintf("TupleGet(%v, %v)", n.Tuple, n.Index))
		case *ArrayCreate:
			return g.Node(newID()).Label(fmt.Sprintf("ArrayCreate(%v, %v)", n.Value, n.Length))
		case *ArrayCreateImmediate:
			return g.Node(newID()).Label(fmt.Sprintf("ArrayCreateImmediate(%v, %v)", n.Value, n.Length))
		case *ArrayGet:
			return g.Node(newID()).Label(fmt.Sprintf("ArrayGet(%v, %v)", n.Array, n.Index))
		case *ArrayGetImmediate:
			return g.Node(newID()).Label(fmt.Sprintf("ArrayGetImmediate(%v, %v)", n.Array, n.Index))
		case *ArrayPut:
			return g.Node(newID()).Label(fmt.Sprintf("ArrayPut(%v, %v, %v)", n.Array, n.Index, n.Value))
		case *ArrayPutImmediate:
			return g.Node(newID()).Label(fmt.Sprintf("ArrayPutImmediate(%v, %v, %v)", n.Array, n.Index, n.Value))
		case *ReadInt:
			return g.Node(newID()).Label("ReadInt")
		case *ReadFloat:
			return g.Node(newID()).Label("ReadFloat")
		case *WriteByte:
			return g.Node(newID()).Label(fmt.Sprintf("WriteByte(%v)", n.Arg))
		case *IntToFloat:
			return g.Node(newID()).Label(fmt.Sprintf("IntToFloat(%v)", n.Arg))
		case *FloatToInt:
			return g.Node(newID()).Label(fmt.Sprintf("FloatToInt(%v)", n.Arg))
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
		g.Node("args").Label(fmt.Sprintf("args = [%s]", strings.Join(function.Args, ", ")))
		generate(function.Body, g)
		if err := ioutil.WriteFile(
			fmt.Sprintf("graphs/%s.dot", function.Name),
			[]byte(g.String()), 0666); err != nil {
			return err
		}
	}

	return nil
}
