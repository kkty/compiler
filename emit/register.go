package emit

import (
	"fmt"
	"github.com/kkty/compiler/ir"
	"github.com/kkty/compiler/stringset"
	"github.com/kkty/compiler/typing"
	"sort"
)

// colorGraph colors a graph with k colors (0, 1, ... k - 1).
// When failed, the second return value is set to false.
func colorGraph(graph map[string]stringset.Set, k int) (map[string]int, bool) {
	nodes := []string{}

	for node := range graph {
		nodes = append(nodes, node)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return len(graph[nodes[i]].Slice()) > len(graph[nodes[j]].Slice())
	})

	colorMap := map[string]int{}

	for _, node := range nodes {
		unavailable := map[int]struct{}{}
		for _, adjacent := range graph[node].Slice() {
			if c, exists := colorMap[adjacent]; exists {
				unavailable[c] = struct{}{}
			}
		}
		mapped := false
		for i := 0; i < k; i++ {
			if _, exists := unavailable[i]; !exists {
				colorMap[node] = i
				mapped = true
				break
			}
		}
		if !mapped {
			return nil, false
		}
	}

	return colorMap, true
}

// AllocateRegisters does register allocation with graph coloring.
// Variable names in ir.Node are replaced with register names like $i1.
// If a variable could not be assigned to any registers, it will be kept unchanged
// and should be treated accordingly in the later stage.
func AllocateRegisters(main ir.Node, functions []*ir.Function, types map[string]typing.Type) {
	allocate := func(function *ir.Function) {
		intGraph := map[string]stringset.Set{}
		floatGraph := map[string]stringset.Set{}

		addEdges := func(variables stringset.Set) {
			for _, i := range variables.Slice() {
				for _, j := range variables.Slice() {
					if i != j {
						if types[i] == typing.FloatType && types[j] == typing.FloatType {
							if floatGraph[i] == nil {
								floatGraph[i] = stringset.New()
							}
							floatGraph[i].Add(j)
						}
						if types[i] != typing.FloatType && types[j] != typing.FloatType {
							if intGraph[i] == nil {
								intGraph[i] = stringset.New()
							}
							intGraph[i].Add(j)
						}
					}
				}
			}
		}

		var liveVariables func(ir.Node, stringset.Set) stringset.Set
		liveVariables = func(node ir.Node, variablesToKeep stringset.Set) stringset.Set {
			switch node.(type) {
			case *ir.IfEqual:
				n := node.(*ir.IfEqual)
				v := stringset.NewFromSlice([]string{n.Left, n.Right})
				v.Join(liveVariables(n.True, variablesToKeep))
				v.Join(liveVariables(n.False, variablesToKeep))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			case *ir.IfEqualZero:
				n := node.(*ir.IfEqualZero)
				v := stringset.NewFromSlice([]string{n.Inner})
				v.Join(liveVariables(n.True, variablesToKeep))
				v.Join(liveVariables(n.False, variablesToKeep))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			case *ir.IfEqualTrue:
				n := node.(*ir.IfEqualTrue)
				v := stringset.NewFromSlice([]string{n.Inner})
				v.Join(liveVariables(n.True, variablesToKeep))
				v.Join(liveVariables(n.False, variablesToKeep))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			case *ir.IfLessThan:
				n := node.(*ir.IfLessThan)
				v := stringset.NewFromSlice([]string{n.Left, n.Right})
				v.Join(liveVariables(n.True, variablesToKeep))
				v.Join(liveVariables(n.False, variablesToKeep))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			case *ir.IfLessThanZero:
				n := node.(*ir.IfLessThanZero)
				v := stringset.NewFromSlice([]string{n.Inner})
				v.Join(liveVariables(n.True, variablesToKeep))
				v.Join(liveVariables(n.False, variablesToKeep))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			case *ir.ValueBinding:
				n := node.(*ir.ValueBinding)
				v := stringset.New()
				v.Join(liveVariables(n.Next, variablesToKeep))
				v.Remove(n.Name)
				restore := v.Join(variablesToKeep)
				v.Join(liveVariables(n.Value, v))
				addEdges(v)
				restore(v)
				return v
			default:
				v := stringset.Set(node.FreeVariables(stringset.New()))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			}
		}

		addEdges(liveVariables(function.Body, stringset.New()))

		getNodes := func(graph map[string]stringset.Set) []string {
			nodes := []string{}
			for node := range graph {
				nodes = append(nodes, node)
			}
			sort.Slice(nodes, func(i, j int) bool {
				return len(graph[nodes[i]].Slice()) > len(graph[nodes[j]].Slice())
			})
			return nodes
		}

		removeNode := func(node string, graph map[string]stringset.Set) {
			for _, adjacent := range graph[node].Slice() {
				graph[adjacent].Remove(node)
				if len(graph[adjacent].Slice()) == 0 {
					delete(graph, adjacent)
				}
			}
			delete(graph, node)
		}

		updateArgs := func(mapping map[string]string) {
			for i, arg := range function.Args {
				if updated, exists := mapping[arg]; exists {
					function.Args[i] = updated
				}
			}
		}

		for _, i := range getNodes(intGraph) {
			if colorMap, ok := colorGraph(intGraph, 21); ok {
				mapping := map[string]string{}
				for variable, color := range colorMap {
					mapping[variable] = fmt.Sprintf("$i%d", color+1)
				}
				function.Body.UpdateNames(mapping)
				updateArgs(mapping)
				break
			}
			removeNode(i, intGraph)
		}

		for _, i := range getNodes(floatGraph) {
			if colorMap, ok := colorGraph(floatGraph, 25); ok {
				mapping := map[string]string{}
				for variable, color := range colorMap {
					mapping[variable] = fmt.Sprintf("$f%d", color+1)
				}
				function.Body.UpdateNames(mapping)
				updateArgs(mapping)
				break
			}
			removeNode(i, floatGraph)
		}
	}

	for _, function := range append(functions, &ir.Function{
		Name: "main",
		Args: nil,
		Body: main,
	}) {
		allocate(function)
	}
}
