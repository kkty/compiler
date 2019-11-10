package alpha

import (
	"fmt"

	"github.com/kkty/mincaml-go/ast"
)

var nextId = 0

func AlphaTransform(node ast.Node) ast.Node {
	return alphaTransform(node, make(map[string]string))
}

func alphaTransform(node ast.Node, mapping map[string]string) ast.Node {
	getNewName := func(name string) string {
		defer func() { nextId++ }()
		return fmt.Sprintf("%s_%d", name, nextId)
	}

	copyMapping := func() map[string]string {
		m := map[string]string{}

		for k, v := range mapping {
			m[k] = v
		}

		return m
	}

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
		return ast.Add{alphaTransform(n.Left, mapping), alphaTransform(n.Right, mapping)}
	case ast.Sub:
		n := node.(ast.Sub)
		return ast.Sub{alphaTransform(n.Left, mapping), alphaTransform(n.Right, mapping)}
	case ast.FloatAdd:
		n := node.(ast.FloatAdd)
		return ast.FloatAdd{alphaTransform(n.Left, mapping), alphaTransform(n.Right, mapping)}
	case ast.FloatSub:
		n := node.(ast.FloatSub)
		return ast.FloatSub{alphaTransform(n.Left, mapping), alphaTransform(n.Right, mapping)}
	case ast.FloatDiv:
		n := node.(ast.FloatDiv)
		return ast.FloatDiv{alphaTransform(n.Left, mapping), alphaTransform(n.Right, mapping)}
	case ast.FloatMul:
		n := node.(ast.FloatMul)
		return ast.FloatMul{alphaTransform(n.Left, mapping), alphaTransform(n.Right, mapping)}
	case ast.Equal:
		n := node.(ast.Equal)
		return ast.Equal{alphaTransform(n.Left, mapping), alphaTransform(n.Right, mapping)}
	case ast.LessThanOrEqual:
		n := node.(ast.LessThanOrEqual)
		return ast.LessThanOrEqual{alphaTransform(n.Left, mapping), alphaTransform(n.Right, mapping)}
	case ast.Neg:
		n := node.(ast.Neg)
		return ast.Neg{alphaTransform(n.Inner, mapping)}
	case ast.FloatNeg:
		n := node.(ast.FloatNeg)
		return ast.FloatNeg{alphaTransform(n.Inner, mapping)}
	case ast.Not:
		n := node.(ast.Not)
		return ast.Not{alphaTransform(n.Inner, mapping)}
	case ast.If:
		n := node.(ast.If)
		return ast.If{alphaTransform(n.Condition, mapping), alphaTransform(n.True, mapping), alphaTransform(n.False, mapping)}
	case ast.ValueBinding:
		n := node.(ast.ValueBinding)
		newMapping := copyMapping()
		newName := getNewName(n.Name)
		newMapping[n.Name] = newName
		return ast.ValueBinding{newName, alphaTransform(n.Body, mapping), alphaTransform(n.Next, newMapping)}
	case ast.FunctionBinding:
		n := node.(ast.FunctionBinding)
		newMapping := copyMapping()
		newMappingForFunction := copyMapping()
		newName := getNewName(n.Name)
		newMapping[n.Name] = newName
		newMappingForFunction[n.Name] = newName
		newArgNames := []string{}
		for _, argName := range n.Args {
			newArgName := getNewName(argName)
			newMappingForFunction[argName] = newArgName
			newArgNames = append(newArgNames, newArgName)
		}
		return ast.FunctionBinding{newName, newArgNames, alphaTransform(n.Body, newMappingForFunction), alphaTransform(n.Next, newMapping)}
	case ast.Application:
		n := node.(ast.Application)
		args := []ast.Node{}

		for _, arg := range n.Args {
			args = append(args, alphaTransform(arg, mapping))
		}

		return ast.Application{alphaTransform(n.Function, mapping), args}
	case ast.Tuple:
		n := node.(ast.Tuple)
		elements := []ast.Node{}

		for _, element := range n.Elements {
			elements = append(elements, alphaTransform(element, mapping))
		}

		return ast.Tuple{elements}
	case ast.TupleBinding:
		n := node.(ast.TupleBinding)
		newMapping := copyMapping()
		newNames := []string{}
		for _, name := range n.Names {
			newName := getNewName(name)
			newNames = append(newNames, newName)
			newMapping[name] = newName
		}
		return ast.TupleBinding{newNames, alphaTransform(n.Tuple, mapping), alphaTransform(n.Next, newMapping)}
	case ast.ArrayCreate:
		n := node.(ast.ArrayCreate)
		return ast.ArrayCreate{alphaTransform(n.Size, mapping), alphaTransform(n.Value, mapping)}
	case ast.ArrayGet:
		n := node.(ast.ArrayGet)
		return ast.ArrayGet{alphaTransform(n.Array, mapping), alphaTransform(n.Index, mapping)}
	case ast.ArrayPut:
		n := node.(ast.ArrayPut)
		return ast.ArrayPut{alphaTransform(n.Array, mapping), alphaTransform(n.Index, mapping), alphaTransform(n.Value, mapping)}

	}

	return nil
}
