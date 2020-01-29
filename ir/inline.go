package ir

import (
	"fmt"
	"github.com/kkty/compiler/typing"
	"os"
)

// Inline does inline expansions for all non-recursive functions and n recursive functions.
func Inline(main Node, functions []*Function, n int, types map[string]typing.Type, debug bool) (Node, []*Function) {
	nextTemporaryId := 0

	temporary := func() string {
		defer func() { nextTemporaryId++ }()
		return fmt.Sprintf("_inline_%d", nextTemporaryId)
	}

	// rename all names in node
	rename := func(node Node) {
		assignments := []*Assignment{}

		// find all assignments using bfs

		queue := []Node{node}

		for len(queue) > 0 {
			node := queue[0]
			queue = queue[1:]

			switch node.(type) {
			case *IfEqual:
				n := node.(*IfEqual)
				queue = append(queue, n.True, n.False)
			case *IfEqualZero:
				n := node.(*IfEqualZero)
				queue = append(queue, n.True, n.False)
			case *IfEqualTrue:
				n := node.(*IfEqualTrue)
				queue = append(queue, n.True, n.False)
			case *IfLessThan:
				n := node.(*IfLessThan)
				queue = append(queue, n.True, n.False)
			case *IfLessThanZero:
				n := node.(*IfLessThanZero)
				queue = append(queue, n.True, n.False)
			case *Assignment:
				n := node.(*Assignment)
				assignments = append(assignments, n)
				queue = append(queue, n.Value, n.Next)
			}
		}

		mapping := map[string]string{}
		for _, assignment := range assignments {
			if assignment.Name != "" {
				t := temporary()
				types[t] = types[assignment.Name]
				mapping[assignment.Name] = t
			}
		}

		node.UpdateNames(mapping)
	}

	// replace the applications of the given function with (a copy of) the function body
	var replaceApplications func(Node, *Function) Node
	replaceApplications = func(node Node, function *Function) Node {
		switch node.(type) {
		case *IfEqual:
			n := node.(*IfEqual)
			n.True = replaceApplications(n.True, function)
			n.False = replaceApplications(n.False, function)
			return n
		case *IfEqualZero:
			n := node.(*IfEqualZero)
			n.True = replaceApplications(n.True, function)
			n.False = replaceApplications(n.False, function)
			return n
		case *IfEqualTrue:
			n := node.(*IfEqualTrue)
			n.True = replaceApplications(n.True, function)
			n.False = replaceApplications(n.False, function)
			return n
		case *IfLessThan:
			n := node.(*IfLessThan)
			n.True = replaceApplications(n.True, function)
			n.False = replaceApplications(n.False, function)
			return n
		case *IfLessThanZero:
			n := node.(*IfLessThanZero)
			n.True = replaceApplications(n.True, function)
			n.False = replaceApplications(n.False, function)
			return n
		case *Assignment:
			n := node.(*Assignment)
			n.Value = replaceApplications(n.Value, function)
			n.Next = replaceApplications(n.Next, function)
			return n
		case *Application:
			n := node.(*Application)
			if n.Function != function.Name {
				return n
			}
			f := function.Body.Clone()
			rename(f)
			mapping := map[string]string{}
			for i, arg := range function.Args {
				mapping[arg] = n.Args[i]
			}
			f.UpdateNames(mapping)
			return f
		default:
			return node
		}
	}

	inline := func(function *Function) {
		if debug {
			fmt.Fprintf(os.Stderr, "inlining %s\n", function.Name)
		}

		main = replaceApplications(main, function)
		for _, f := range functions {
			if f.Name != function.Name {
				f.Body = replaceApplications(f.Body, function)
			}
		}

		// when a non-recursive function is inlined, it can be removed from the function list.
		if !function.IsRecursive() {
			updated := []*Function{}
			for _, f := range functions {
				if f.Name != function.Name {
					updated = append(updated, f)
				}
			}
			functions = updated
		}
	}

	// non-recursive functions

	for {
		updated := false
		for _, function := range functions {
			if !function.IsRecursive() {
				inline(function)
				updated = true
			}
		}
		if !updated {
			break
		}
	}

	// recursive functions

	cnt := map[string]int{}

	priority := func(f *Function) int {
		return -(f.Body.Size() + cnt[f.Name]*5)
	}

	for i := 0; i < n; i++ {
		if len(functions) == 0 {
			break
		}

		// select the function with the highest priority
		var target *Function
		for _, function := range functions {
			if target == nil || priority(function) > priority(target) {
				target = function
			}
		}

		inline(target)
		cnt[target.Name]++
	}

	return main, functions
}
