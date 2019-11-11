package alpha

import (
	"fmt"
	"log"

	"github.com/kkty/mincaml-go/ast"
)

// AlphaTransform renames all the names in a program so that they are different
// from each other, without changing the program's behaviour.
func AlphaTransform(node ast.Node) ast.Node {
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

	var transform func(node ast.Node, mapping map[string]string) ast.Node
	transform = func(node ast.Node, mapping map[string]string) ast.Node {
		switch node.(type) {
		case ast.Variable:
			return ast.Variable{mapping[node.(ast.Variable).Name]}
		case ast.Unit:
			return node
		case ast.Int:
			return node
		case ast.Bool:
			return node
		case ast.Float:
			return node
		case ast.Add:
			n := node.(ast.Add)
			return ast.Add{transform(n.Left, mapping), transform(n.Right, mapping)}
		case ast.Sub:
			n := node.(ast.Sub)
			return ast.Sub{transform(n.Left, mapping), transform(n.Right, mapping)}
		case ast.FloatAdd:
			n := node.(ast.FloatAdd)
			return ast.FloatAdd{transform(n.Left, mapping), transform(n.Right, mapping)}
		case ast.FloatSub:
			n := node.(ast.FloatSub)
			return ast.FloatSub{transform(n.Left, mapping), transform(n.Right, mapping)}
		case ast.FloatDiv:
			n := node.(ast.FloatDiv)
			return ast.FloatDiv{transform(n.Left, mapping), transform(n.Right, mapping)}
		case ast.FloatMul:
			n := node.(ast.FloatMul)
			return ast.FloatMul{transform(n.Left, mapping), transform(n.Right, mapping)}
		case ast.Equal:
			n := node.(ast.Equal)
			return ast.Equal{transform(n.Left, mapping), transform(n.Right, mapping)}
		case ast.LessThanOrEqual:
			n := node.(ast.LessThanOrEqual)
			return ast.LessThanOrEqual{transform(n.Left, mapping), transform(n.Right, mapping)}
		case ast.Neg:
			n := node.(ast.Neg)
			return ast.Neg{transform(n.Inner, mapping)}
		case ast.FloatNeg:
			n := node.(ast.FloatNeg)
			return ast.FloatNeg{transform(n.Inner, mapping)}
		case ast.Not:
			n := node.(ast.Not)
			return ast.Not{transform(n.Inner, mapping)}
		case ast.If:
			n := node.(ast.If)
			return ast.If{transform(n.Condition, mapping), transform(n.True, mapping), transform(n.False, mapping)}
		case ast.ValueBinding:
			n := node.(ast.ValueBinding)
			newMapping := copyMapping(mapping)
			newName := getNewName(n.Name)
			newMapping[n.Name] = newName
			return ast.ValueBinding{newName, transform(n.Body, mapping), transform(n.Next, newMapping)}
		case ast.FunctionBinding:
			n := node.(ast.FunctionBinding)
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
			return ast.FunctionBinding{newName, newArgNames, transform(n.Body, newMappingForFunction), transform(n.Next, newMapping)}
		case ast.Application:
			n := node.(ast.Application)
			args := []ast.Node{}

			for _, arg := range n.Args {
				args = append(args, transform(arg, mapping))
			}

			return ast.Application{mapping[n.Function], args}
		case ast.Tuple:
			n := node.(ast.Tuple)
			elements := []ast.Node{}

			for _, element := range n.Elements {
				elements = append(elements, transform(element, mapping))
			}

			return ast.Tuple{elements}
		case ast.TupleBinding:
			n := node.(ast.TupleBinding)
			newMapping := copyMapping(mapping)
			newNames := []string{}
			for _, name := range n.Names {
				newName := getNewName(name)
				newNames = append(newNames, newName)
				newMapping[name] = newName
			}
			return ast.TupleBinding{newNames, transform(n.Tuple, mapping), transform(n.Next, newMapping)}
		case ast.ArrayCreate:
			n := node.(ast.ArrayCreate)
			return ast.ArrayCreate{transform(n.Size, mapping), transform(n.Value, mapping)}
		case ast.ArrayGet:
			n := node.(ast.ArrayGet)
			return ast.ArrayGet{transform(n.Array, mapping), transform(n.Index, mapping)}
		case ast.ArrayPut:
			n := node.(ast.ArrayPut)
			return ast.ArrayPut{transform(n.Array, mapping), transform(n.Index, mapping), transform(n.Value, mapping)}
		case ast.ReadInt:
			return node
		case ast.ReadFloat:
			return node
		case ast.PrintInt:
			n := node.(ast.PrintInt)
			return ast.PrintInt{transform(n.Inner, mapping)}
		case ast.PrintChar:
			n := node.(ast.PrintChar)
			return ast.PrintChar{transform(n.Inner, mapping)}
		}

		log.Fatal("invalid ast node")
		return nil
	}

	return transform(node, map[string]string{})
}
