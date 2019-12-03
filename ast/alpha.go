package ast

import (
	"fmt"

	"github.com/kkty/compiler/stringmap"
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
		default:
			for _, n := range node.Children() {
				transform(n, mapping)
			}
		}
	}

	transform(node, stringmap.New())
}
