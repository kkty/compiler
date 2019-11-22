package alpha

import (
	"fmt"

	"github.com/kkty/mincaml-go/ast"
)

// AlphaTransform renames all the names in a program so that they are different
// from each other, without changing the program's behaviour.
func AlphaTransform(node ast.Node) {
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

	var transform func(node ast.Node, mapping map[string]string)
	transform = func(node ast.Node, mapping map[string]string) {
		switch node.(type) {
		case *ast.Variable:
			n := node.(*ast.Variable)
			n.Name = mapping[n.Name]
		case *ast.Add:
			n := node.(*ast.Add)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *ast.Sub:
			n := node.(*ast.Sub)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *ast.FloatAdd:
			n := node.(*ast.FloatAdd)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *ast.FloatSub:
			n := node.(*ast.FloatSub)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *ast.FloatDiv:
			n := node.(*ast.FloatDiv)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *ast.FloatMul:
			n := node.(*ast.FloatMul)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *ast.Equal:
			n := node.(*ast.Equal)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *ast.LessThan:
			n := node.(*ast.LessThan)
			transform(n.Left, mapping)
			transform(n.Right, mapping)
		case *ast.Neg:
			n := node.(*ast.Neg)
			transform(n.Inner, mapping)
		case *ast.FloatNeg:
			n := node.(*ast.FloatNeg)
			transform(n.Inner, mapping)
		case *ast.Not:
			n := node.(*ast.Not)
			transform(n.Inner, mapping)
		case *ast.If:
			n := node.(*ast.If)
			transform(n.Condition, mapping)
			transform(n.True, mapping)
			transform(n.False, mapping)
		case *ast.ValueBinding:
			n := node.(*ast.ValueBinding)
			newMapping := copyMapping(mapping)
			newName := getNewName(n.Name)
			newMapping[n.Name] = newName
			n.Name = newName
			transform(n.Body, mapping)
			transform(n.Next, newMapping)
		case *ast.FunctionBinding:
			n := node.(*ast.FunctionBinding)
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
		case *ast.Application:
			n := node.(*ast.Application)

			for i := range n.Args {
				transform(n.Args[i], mapping)
			}

			n.Function = mapping[n.Function]
		case *ast.Tuple:
			n := node.(*ast.Tuple)

			for i := range n.Elements {
				transform(n.Elements[i], mapping)
			}
		case *ast.TupleBinding:
			n := node.(*ast.TupleBinding)
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
		case *ast.ArrayCreate:
			n := node.(*ast.ArrayCreate)
			transform(n.Size, mapping)
			transform(n.Value, mapping)
		case *ast.ArrayGet:
			n := node.(*ast.ArrayGet)
			transform(n.Array, mapping)
			transform(n.Index, mapping)
		case *ast.ArrayPut:
			n := node.(*ast.ArrayPut)
			transform(n.Array, mapping)
			transform(n.Index, mapping)
			transform(n.Value, mapping)
		case *ast.PrintInt:
			n := node.(*ast.PrintInt)
			transform(n.Inner, mapping)
		case *ast.PrintChar:
			n := node.(*ast.PrintChar)
			transform(n.Inner, mapping)
		case *ast.IntToFloat:
			n := node.(*ast.IntToFloat)
			transform(n.Inner, mapping)
		case *ast.FloatToInt:
			n := node.(*ast.FloatToInt)
			transform(n.Inner, mapping)
		case *ast.Sqrt:
			n := node.(*ast.Sqrt)
			transform(n.Inner, mapping)
		}
	}

	transform(node, map[string]string{})
}
