package ir

func RemoveRedundantVariables(main Node, functions []*Function) Node {
	functionsWithoutSideEffects := FunctionsWithoutSideEffects(functions)

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
		case *IfEqualTrue:
			n := node.(*IfEqualTrue)
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
		case *Assignment:
			n := node.(*Assignment)
			if _, hasFreeVariable := n.Next.FreeVariables(map[string]struct{}{})[n.Name]; n.Value.HasSideEffects(functionsWithoutSideEffects) || hasFreeVariable {
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
