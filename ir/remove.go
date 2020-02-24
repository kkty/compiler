package ir

import (
	"github.com/kkty/compiler/stringset"
)

func RemoveRedundantAssignments(main Node, functions []*Function) Node {
	var remove func(node Node) Node
	remove = func(node Node) Node {
		switch n := node.(type) {
		case *IfEqual:
			n.True = remove(n.True)
			n.False = remove(n.False)
			return n
		case *IfEqualZero:
			n.True = remove(n.True)
			n.False = remove(n.False)
			return n
		case *IfEqualTrue:
			n.True = remove(n.True)
			n.False = remove(n.False)
			return n
		case *IfLessThan:
			n.True = remove(n.True)
			n.False = remove(n.False)
			return n
		case *IfLessThanFloat:
			n.True = remove(n.True)
			n.False = remove(n.False)
			return n
		case *IfLessThanZero:
			n.True = remove(n.True)
			n.False = remove(n.False)
			return n
		case *IfLessThanZeroFloat:
			n.True = remove(n.True)
			n.False = remove(n.False)
			return n
		case *Assignment:
			n.Value = remove(n.Value)
			n.Next = remove(n.Next)
			if next, ok := n.Next.(*Variable); ok {
				if next.Name == n.Name {
					return n.Value
				}
			}
			return n
		default:
			return n
		}
	}

	for _, function := range functions {
		function.Body = remove(function.Body)
	}

	return remove(main)
}

func RemoveRedundantVariables(main Node, functions []*Function) Node {
	functionsWithoutSideEffects := FunctionsWithoutSideEffects(functions)

	var removeRedundantVariables func(node Node) Node
	removeRedundantVariables = func(node Node) Node {
		switch n := node.(type) {
		case *IfEqual:
			n.True = removeRedundantVariables(n.True)
			n.False = removeRedundantVariables(n.False)
			return n
		case *IfEqualZero:
			n.True = removeRedundantVariables(n.True)
			n.False = removeRedundantVariables(n.False)
			return n
		case *IfEqualTrue:
			n.True = removeRedundantVariables(n.True)
			n.False = removeRedundantVariables(n.False)
			return n
		case *IfLessThan:
			n.True = removeRedundantVariables(n.True)
			n.False = removeRedundantVariables(n.False)
			return n
		case *IfLessThanFloat:
			n.True = removeRedundantVariables(n.True)
			n.False = removeRedundantVariables(n.False)
			return n
		case *IfLessThanZero:
			n.True = removeRedundantVariables(n.True)
			n.False = removeRedundantVariables(n.False)
			return n
		case *IfLessThanZeroFloat:
			n.True = removeRedundantVariables(n.True)
			n.False = removeRedundantVariables(n.False)
			return n
		case *Assignment:
			if _, hasFreeVariable := n.Next.FreeVariables(stringset.New())[n.Name]; n.Value.HasSideEffects(functionsWithoutSideEffects) || hasFreeVariable {
				n.Value = removeRedundantVariables(n.Value)
				n.Next = removeRedundantVariables(n.Next)
				return n
			}
			return removeRedundantVariables(n.Next)
		default:
			return node
		}
	}

	for _, function := range functions {
		function.Body = removeRedundantVariables(function.Body)
	}

	return removeRedundantVariables(main)
}
