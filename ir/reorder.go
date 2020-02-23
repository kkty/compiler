package ir

import "github.com/kkty/compiler/stringset"

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
		case *IfLessThanFloat:
			n := node.(*IfLessThanFloat)
			n.True = reorder(n.True)
			n.False = reorder(n.False)
			return n
		case *IfLessThanZero:
			n := node.(*IfLessThanZero)
			n.True = reorder(n.True)
			n.False = reorder(n.False)
			return n
		case *IfLessThanZeroFloat:
			n := node.(*IfLessThanZeroFloat)
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

			switch next := n.Next.(type) {
			case *IfEqual:
				if next.Left == n.Name || next.Right == n.Name {
					return n
				}

				if !next.True.FreeVariables(stringset.New()).Has(n.Name) {
					next.False = &Assignment{n.Name, reorder(n.Value), reorder(next.False)}
					return next
				}

				if !next.False.FreeVariables(stringset.New()).Has(n.Name) {
					next.True = &Assignment{n.Name, reorder(n.Value), reorder(next.True)}
					return next
				}
			case *IfEqualZero:
				if next.Inner == n.Name {
					return n
				}

				if !next.True.FreeVariables(stringset.New()).Has(n.Name) {
					next.False = &Assignment{n.Name, reorder(n.Value), reorder(next.False)}
					return next
				}

				if !next.False.FreeVariables(stringset.New()).Has(n.Name) {
					next.True = &Assignment{n.Name, reorder(n.Value), reorder(next.True)}
					return next
				}
			case *IfEqualTrue:
				if next.Inner == n.Name {
					return n
				}

				if !next.True.FreeVariables(stringset.New()).Has(n.Name) {
					next.False = &Assignment{n.Name, reorder(n.Value), reorder(next.False)}
					return next
				}

				if !next.False.FreeVariables(stringset.New()).Has(n.Name) {
					next.True = &Assignment{n.Name, reorder(n.Value), reorder(next.True)}
					return next
				}
			case *IfLessThan:
				if next.Left == n.Name || next.Right == n.Name {
					return n
				}

				if !next.True.FreeVariables(stringset.New()).Has(n.Name) {
					next.False = &Assignment{n.Name, reorder(n.Value), reorder(next.False)}
					return next
				}

				if !next.False.FreeVariables(stringset.New()).Has(n.Name) {
					next.True = &Assignment{n.Name, reorder(n.Value), reorder(next.True)}
					return next
				}
			case *IfLessThanFloat:
				if next.Left == n.Name || next.Right == n.Name {
					return n
				}

				if !next.True.FreeVariables(stringset.New()).Has(n.Name) {
					next.False = &Assignment{n.Name, reorder(n.Value), reorder(next.False)}
					return next
				}

				if !next.False.FreeVariables(stringset.New()).Has(n.Name) {
					next.True = &Assignment{n.Name, reorder(n.Value), reorder(next.True)}
					return next
				}
			case *IfLessThanZero:
				if next.Inner == n.Name {
					return n
				}

				if !next.True.FreeVariables(stringset.New()).Has(n.Name) {
					next.False = &Assignment{n.Name, reorder(n.Value), reorder(next.False)}
					return next
				}

				if !next.False.FreeVariables(stringset.New()).Has(n.Name) {
					next.True = &Assignment{n.Name, reorder(n.Value), reorder(next.True)}
					return next
				}
			case *IfLessThanZeroFloat:
				if next.Inner == n.Name {
					return n
				}

				if !next.True.FreeVariables(stringset.New()).Has(n.Name) {
					next.False = &Assignment{n.Name, reorder(n.Value), reorder(next.False)}
					return next
				}

				if !next.False.FreeVariables(stringset.New()).Has(n.Name) {
					next.True = &Assignment{n.Name, reorder(n.Value), reorder(next.True)}
					return next
				}
			case *Assignment:
				if !next.Value.FreeVariables(stringset.New()).Has(n.Name) {
					next.Next = &Assignment{n.Name, reorder(n.Value), reorder(next.Next)}
					return next
				}

				if !next.Next.FreeVariables(stringset.New()).Has(n.Name) {
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
