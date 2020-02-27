package ir

import (
	"github.com/kkty/compiler/stringset"
)

// RemoveRedundantAssignments traverses the program and removes variable
// assignments if possible.
func RemoveRedundantAssignments(main Node, functions []*Function) Node {
	functionsWithoutSideEffects := FunctionsWithoutSideEffects(functions)

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
			usedInNext := n.Next.FreeVariables(stringset.New()).Has(n.Name)
			if !usedInNext && !n.Value.HasSideEffects(functionsWithoutSideEffects) {
				return remove(n.Next)
			}
			if next, ok := n.Next.(*Variable); ok {
				if next.Name == n.Name {
					return remove(n.Value)
				}
			}
			if next, ok := n.Next.(*IfEqualTrue); ok && !usedInNext {
				switch value := n.Value.(type) {
				case *Equal:
					return &IfEqual{value.Left, value.Right, remove(next.True), remove(next.False)}
				case *EqualZero:
					return &IfEqualZero{value.Inner, remove(next.True), remove(next.False)}
				case *LessThan:
					return &IfLessThan{value.Left, value.Right, remove(next.True), remove(next.False)}
				case *LessThanFloat:
					return &IfLessThanFloat{value.Left, value.Right, remove(next.True), remove(next.False)}
				case *LessThanZero:
					return &IfLessThanZero{value.Inner, remove(next.True), remove(next.False)}
				case *LessThanZeroFloat:
					return &IfLessThanZeroFloat{value.Inner, remove(next.True), remove(next.False)}
				case *GreaterThanZero:
					return &IfLessThanZero{value.Inner, remove(next.False), remove(next.True)}
				case *GreaterThanZeroFloat:
					return &IfLessThanZeroFloat{value.Inner, remove(next.False), remove(next.True)}
				}
			}
			n.Next = remove(n.Next)
			n.Value = remove(n.Value)
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
