package ast

import (
	"fmt"

	"github.com/kkty/mincaml-go/stringmap"
)

// AlphaTransform renames all the names in a program so that they are different
// from each other, without changing the program's behaviour.
func AlphaTransform(node Node) {
	nextId := 0

	getNewName := func(name string) string {
		defer func() { nextId++ }()
		return fmt.Sprintf("%s_%d", name, nextId)
	}

	var transform func(node Node, mapping stringmap.Map)
	transform = func(node Node, mapping stringmap.Map) {
		switch node.(type) {
		case *Variable:
			n := node.(*Variable)
			n.Name = mapping[n.Name]
		case *Add:
			n := node.(*Add)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *Sub:
			n := node.(*Sub)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *FloatAdd:
			n := node.(*FloatAdd)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *FloatSub:
			n := node.(*FloatSub)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *FloatDiv:
			n := node.(*FloatDiv)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *FloatMul:
			n := node.(*FloatMul)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *Equal:
			n := node.(*Equal)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *LessThan:
			n := node.(*LessThan)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *Neg:
			n := node.(*Neg)
			transform(n.Inner, mapping)
		case *FloatNeg:
			n := node.(*FloatNeg)
			transform(n.Inner, mapping)
		case *Not:
			n := node.(*Not)
			transform(n.Inner, mapping)
		case *If:
			n := node.(*If)
			transform(n.Condition, mapping)
			transform(n.True, mapping)
			transform(n.False, mapping)
		case *ValueBinding:
			n := node.(*ValueBinding)

			transform(n.Body, mapping)

			newName := getNewName(n.Name)

			{
				restore := mapping.Join(stringmap.Map{n.Name: newName})
				transform(n.Next, mapping)
				restore(mapping)
			}

			n.Name = newName
		case *FunctionBinding:
			n := node.(*FunctionBinding)

			newName := getNewName(n.Name)
			newArgNames := []string{}
			for _, argName := range n.Args {
				newArgNames = append(newArgNames, getNewName(argName))
			}

			// Updates n.Body.
			{
				newMapping := stringmap.New()
				newMapping[n.Name] = newName
				for i, arg := range n.Args {
					newMapping[arg] = newArgNames[i]
				}
				restore := mapping.Join(newMapping)
				transform(n.Body, mapping)
				restore(mapping)
			}

			// Updates n.Next.
			{
				newMapping := stringmap.New()
				newMapping[n.Name] = newName
				restore := mapping.Join(newMapping)
				transform(n.Next, mapping)
				restore(mapping)
			}

			n.Name, n.Args = newName, newArgNames
		case *Application:
			n := node.(*Application)

			for i := range n.Args {
				transform(n.Args[i], mapping)
			}

			n.Function = mapping[n.Function]
		case *Tuple:
			n := node.(*Tuple)

			for i := range n.Elements {
				transform(n.Elements[i], mapping)
			}
		case *TupleBinding:
			n := node.(*TupleBinding)

			transform(n.Tuple, mapping)

			newNames := []string{}

			for _, name := range n.Names {
				newNames = append(newNames, getNewName(name))
			}

			{
				newMapping := stringmap.New()
				for i, name := range n.Names {
					newMapping[name] = newNames[i]
				}
				restore := mapping.Join(newMapping)
				transform(n.Next, mapping)
				restore(mapping)
			}

			n.Names = newNames
		case *ArrayCreate:
			n := node.(*ArrayCreate)
			transform(n.Size, mapping)
			transform(n.Value, mapping)
		case *ArrayGet:
			n := node.(*ArrayGet)
			transform(n.Array, mapping)
			transform(n.Index, mapping)
		case *ArrayPut:
			n := node.(*ArrayPut)
			transform(n.Array, mapping)
			transform(n.Index, mapping)
			transform(n.Value, mapping)
		case *PrintInt:
			n := node.(*PrintInt)
			transform(n.Inner, mapping)
		case *PrintChar:
			n := node.(*PrintChar)
			transform(n.Inner, mapping)
		case *IntToFloat:
			n := node.(*IntToFloat)
			transform(n.Inner, mapping)
		case *FloatToInt:
			n := node.(*FloatToInt)
			transform(n.Inner, mapping)
		case *Sqrt:
			n := node.(*Sqrt)
			transform(n.Inner, mapping)
		}
	}

	transform(node, map[string]string{})
}
