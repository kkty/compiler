package ir

import (
	"fmt"
	"os"
	"sort"

	"github.com/kkty/compiler/typing"
)

var nextTemporaryId = 0

func temporary() string {
	defer func() { nextTemporaryId++ }()
	return fmt.Sprintf("_inline_%d", nextTemporaryId)
}

func rename(node Node, types map[string]typing.Type) {
	valueBindings := []*ValueBinding{}

	// Finds application bindings in the subtree.

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
		case *ValueBinding:
			n := node.(*ValueBinding)
			valueBindings = append(valueBindings, n)
			queue = append(queue, n.Value, n.Next)
		}
	}

	mapping := map[string]string{}
	for _, valueBinding := range valueBindings {
		t := temporary()
		types[t] = types[valueBinding.Name]
		mapping[valueBinding.Name] = t
	}

	node.UpdateNames(mapping)
}

// Replaces the applications of the given function with (a copy of) the function body.
func replaceApplications(node Node, function *Function, types map[string]typing.Type) Node {
	switch node.(type) {
	case *IfEqual:
		n := node.(*IfEqual)
		n.True = replaceApplications(n.True, function, types)
		n.False = replaceApplications(n.False, function, types)
		return n
	case *IfEqualZero:
		n := node.(*IfEqualZero)
		n.True = replaceApplications(n.True, function, types)
		n.False = replaceApplications(n.False, function, types)
		return n
	case *IfEqualTrue:
		n := node.(*IfEqualTrue)
		n.True = replaceApplications(n.True, function, types)
		n.False = replaceApplications(n.False, function, types)
		return n
	case *IfLessThan:
		n := node.(*IfLessThan)
		n.True = replaceApplications(n.True, function, types)
		n.False = replaceApplications(n.False, function, types)
		return n
	case *IfLessThanZero:
		n := node.(*IfLessThanZero)
		n.True = replaceApplications(n.True, function, types)
		n.False = replaceApplications(n.False, function, types)
		return n
	case *ValueBinding:
		n := node.(*ValueBinding)
		n.Value = replaceApplications(n.Value, function, types)
		n.Next = replaceApplications(n.Next, function, types)
		return n
	case *Application:
		n := node.(*Application)
		if n.Function != function.Name {
			return n
		}
		f := function.Body.Clone()
		rename(f, types)
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

// Inline does inline expansions for n times.
func Inline(main Node, functions []*Function, n int, types map[string]typing.Type, debug bool) (Node, []*Function) {
	cnt := map[string]int{}
	for i := 0; i < n; i++ {
		if len(functions) == 0 {
			break
		}

		sort.Slice(functions, func(i, j int) bool {
			return len(functions[i].Args)+cnt[functions[i].Name]*10 < len(functions[j].Args)+cnt[functions[j].Name]*10
		})

		function := functions[0]

		if debug {
			fmt.Fprintf(os.Stderr, "inlining %s\n", function.Name)
		}

		main = replaceApplications(main, function, types)
		for _, f := range functions {
			if f.Name != function.Name {
				f.Body = replaceApplications(f.Body, function, types)
			}
		}

		if function.IsRecursive() {
			cnt[function.Name]++
		} else {
			updated := []*Function{}
			for _, f := range functions {
				if f.Name != function.Name {
					updated = append(updated, f)
				}
			}
			functions = updated
		}
	}

	return main, functions
}
