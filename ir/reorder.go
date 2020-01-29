package ir

func Reorder(main Node, functions []*Function) Node {
	functionsWithoutSideEffects := FunctionsWithoutSideEffects(functions)

	var reorder func(node Node) Node
	reorder = func(node Node) Node {
		switch node.(type) {
		case *IfEqual:
			n := node.(*IfEqual)
			n.True = reorder(n.True)
			n.False = reorder(n.False)
			return n
		case *IfEqualZero:
			n := node.(*IfEqualZero)
			n.True = reorder(n.True)
			n.False = reorder(n.False)
			return n
		case *IfEqualTrue:
			n := node.(*IfEqualTrue)
			n.True = reorder(n.True)
			n.False = reorder(n.False)
			return n
		case *IfLessThan:
			n := node.(*IfLessThan)
			n.True = reorder(n.True)
			n.False = reorder(n.False)
			return n
		case *IfLessThanZero:
			n := node.(*IfLessThanZero)
			n.True = reorder(n.True)
			n.False = reorder(n.False)
			return n
		case *Assignment:
			n := node.(*Assignment)

			if n.Value.HasSideEffects(functionsWithoutSideEffects) {
				n.Value = reorder(n.Value)
				n.Next = reorder(n.Next)
				return n
			}

			switch n.Next.(type) {
			case *IfEqual:
				next := n.Next.(*IfEqual)

				if next.Left == n.Name || next.Right == n.Name {
					return n
				}

				if _, exists := next.True.FreeVariables(map[string]struct{}{})[n.Name]; !exists {
					next.False = &Assignment{n.Name, reorder(n.Value), reorder(next.False)}
					return next
				}

				if _, exists := next.False.FreeVariables(map[string]struct{}{})[n.Name]; !exists {
					next.True = &Assignment{n.Name, reorder(n.Value), reorder(next.True)}
					return next
				}
			case *IfEqualZero:
				next := n.Next.(*IfEqualZero)

				if next.Inner == n.Name {
					return n
				}

				if _, exists := next.True.FreeVariables(map[string]struct{}{})[n.Name]; !exists {
					next.False = &Assignment{n.Name, reorder(n.Value), reorder(next.False)}
					return next
				}

				if _, exists := next.False.FreeVariables(map[string]struct{}{})[n.Name]; !exists {
					next.True = &Assignment{n.Name, reorder(n.Value), reorder(next.True)}
					return next
				}
			case *IfEqualTrue:
				next := n.Next.(*IfEqualTrue)

				if next.Inner == n.Name {
					return n
				}

				if _, exists := next.True.FreeVariables(map[string]struct{}{})[n.Name]; !exists {
					next.False = &Assignment{n.Name, reorder(n.Value), reorder(next.False)}
					return next
				}

				if _, exists := next.False.FreeVariables(map[string]struct{}{})[n.Name]; !exists {
					next.True = &Assignment{n.Name, reorder(n.Value), reorder(next.True)}
					return next
				}
			case *IfLessThan:
				next := n.Next.(*IfLessThan)

				if next.Left == n.Name || next.Right == n.Name {
					return n
				}

				if _, exists := next.True.FreeVariables(map[string]struct{}{})[n.Name]; !exists {
					next.False = &Assignment{n.Name, reorder(n.Value), reorder(next.False)}
					return next
				}

				if _, exists := next.False.FreeVariables(map[string]struct{}{})[n.Name]; !exists {
					next.True = &Assignment{n.Name, reorder(n.Value), reorder(next.True)}
					return next
				}
			case *IfLessThanZero:
				next := n.Next.(*IfLessThanZero)

				if next.Inner == n.Name {
					return n
				}

				if _, exists := next.True.FreeVariables(map[string]struct{}{})[n.Name]; !exists {
					next.False = &Assignment{n.Name, reorder(n.Value), reorder(next.False)}
					return next
				}

				if _, exists := next.False.FreeVariables(map[string]struct{}{})[n.Name]; !exists {
					next.True = &Assignment{n.Name, reorder(n.Value), reorder(next.True)}
					return next
				}
			case *Assignment:
				next := n.Next.(*Assignment)

				if _, exists := next.Value.FreeVariables(map[string]struct{}{})[n.Name]; !exists {
					next.Next = &Assignment{n.Name, reorder(n.Value), reorder(next.Next)}
					return next
				}

				if _, exists := next.Next.FreeVariables(map[string]struct{}{})[n.Name]; !exists {
					next.Value = &Assignment{n.Name, reorder(n.Value), reorder(next.Value)}
					return next
				}
			}

			n.Value = reorder(n.Value)
			n.Next = reorder(n.Next)

			return n
		default:
			return node
		}
	}

	for _, function := range functions {
		function.Body = reorder(function.Body)
	}

	return reorder(main)
}
