package ir

import (
	"github.com/thoas/go-funk"
)

func RemoveRedundantVariables(main Node, functions []*Function) Node {
	var removeRedundantVariables func(node Node) Node
	removeRedundantVariables = func(node Node) Node {
		switch node.(type) {
		case *IfEqual:
			n := node.(*IfEqual)
			n.True = removeRedundantVariables(n.True)
			n.False = removeRedundantVariables(n.False)
			return n
		case *IfEqualZero:
			n := node.(*IfEqualZero)
			n.True = removeRedundantVariables(n.True)
			n.False = removeRedundantVariables(n.False)
			return n
		case *IfLessThan:
			n := node.(*IfLessThan)
			n.True = removeRedundantVariables(n.True)
			n.False = removeRedundantVariables(n.False)
			return n
		case *IfLessThanZero:
			n := node.(*IfLessThanZero)
			n.True = removeRedundantVariables(n.True)
			n.False = removeRedundantVariables(n.False)
			return n
		case *ValueBinding:
			n := node.(*ValueBinding)
			if n.Value.HasSideEffects() || funk.ContainsString(n.Next.FreeVariables(map[string]struct{}{}), n.Name) {
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
