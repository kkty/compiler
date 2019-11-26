package ast

import (
	"fmt"
)

// AlphaTransform renames all the names in a program so that they are different
// from each other, without changing the program's behaviour.
func AlphaTransform(node Node) {
	nextId := 0

	getNewName := func(name string) string {
		defer func() { nextId++ }()
		return fmt.Sprintf("%s_%d", name, nextId)
	}

	copyMapping := func(original map[string]string) map[string]string {
		m := map[string]string{}

		for k, v := range original {
			m[k] = v
		}

		return m
	}

	var transform func(node Node, mapping map[string]string)
	transform = func(node Node, mapping map[string]string) {
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
			newMapping := copyMapping(mapping)
			newName := getNewName(n.Name)
			newMapping[n.Name] = newName
			n.Name = newName
			transform(n.Body, mapping)
			transform(n.Next, newMapping)
		case *FunctionBinding:
			n := node.(*FunctionBinding)
			newMapping := copyMapping(mapping)
			newMappingForFunction := copyMapping(mapping)
			newName := getNewName(n.Name)
			newMapping[n.Name] = newName
			newMappingForFunction[n.Name] = newName
			newArgNames := []string{}
			for _, argName := range n.Args {
				newArgName := getNewName(argName)
				newMappingForFunction[argName] = newArgName
				newArgNames = append(newArgNames, newArgName)
			}
			n.Name, n.Args = newName, newArgNames
			transform(n.Body, newMappingForFunction)
			transform(n.Next, newMapping)
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
			newMapping := copyMapping(mapping)
			newNames := []string{}
			for _, name := range n.Names {
				newName := getNewName(name)
				newNames = append(newNames, newName)
				newMapping[name] = newName
			}
			n.Names = newNames
			transform(n.Tuple, mapping)
			transform(n.Next, newMapping)
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
