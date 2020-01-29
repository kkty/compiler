package emit

import (
	"fmt"
	"github.com/kkty/compiler/ir"
	"github.com/kkty/compiler/stringset"
	"github.com/kkty/compiler/typing"
	"sort"
)

var (
	intRegisters   []string
	floatRegisters []string
)

func init() {
	for i := 0; i < 22; i++ {
		intRegisters = append(intRegisters, fmt.Sprintf("$i%d", i+1))
	}
	for i := 0; i < 27; i++ {
		floatRegisters = append(floatRegisters, fmt.Sprintf("$f%d", i+1))
	}
}

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
// The number of spills for each function is returned.
func AllocateRegisters(main ir.Node, functions []*ir.Function, types map[string]typing.Type) map[string]int {
	spills := map[string]int{}

	allocate := func(function *ir.Function) {
		intGraph := map[string]stringset.Set{}
		floatGraph := map[string]stringset.Set{}

		addEdges := func(variables stringset.Set) {
			for _, i := range variables.Slice() {
				if _, ok := types[i].(*typing.FloatType); ok {
					if _, exists := floatGraph[i]; !exists {
						floatGraph[i] = stringset.New()
					}
				} else {
					if _, exists := intGraph[i]; !exists {
						intGraph[i] = stringset.New()
					}
				}
			}
			for _, i := range variables.Slice() {
				for _, j := range variables.Slice() {
					if i != j {
						if _, ok := types[i].(*typing.FloatType); ok {
							if _, ok := types[j].(*typing.FloatType); ok {
								floatGraph[i].Add(j)
							}
						}

						if _, ok := types[i].(*typing.FloatType); !ok {
							if _, ok := types[j].(*typing.FloatType); !ok {
								intGraph[i].Add(j)
							}
						}
					}
				}
			}
		}

		// liveVariables returns live variables at a node, creating the interference graphs at the same time.
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
			case *ir.Assignment:
				n := node.(*ir.Assignment)
				v := stringset.New()
				v.Join(liveVariables(n.Next, variablesToKeep))
				v.Remove(n.Name)
				copied := v.Copy()
				copied.Join(variablesToKeep)
				v.Join(liveVariables(n.Value, copied))
				restore := v.Join(variablesToKeep)
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

		// getNodes returns a list of nodes in a graph.
		// Nodes are sorted so that the one with the highest degree comes first.
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

		// removeNode removes a node from a graph.
		removeNode := func(node string, graph map[string]stringset.Set) {
			for _, adjacent := range graph[node].Slice() {
				graph[adjacent].Remove(node)
			}
			delete(graph, node)
		}

		// variables names to register names
		mapping := map[string]string{}

		for _, i := range getNodes(intGraph) {
			if colorMap, ok := colorGraph(intGraph, len(intRegisters)); ok {
				for variable, color := range colorMap {
					mapping[variable] = intRegisters[color]
				}
				break
			}
			removeNode(i, intGraph)
			spills[function.Name]++
		}

		for _, i := range getNodes(floatGraph) {
			if colorMap, ok := colorGraph(floatGraph, len(floatRegisters)); ok {
				for variable, color := range colorMap {
					mapping[variable] = floatRegisters[color]
				}
				break
			}
			removeNode(i, floatGraph)
			spills[function.Name]++
		}

		function.Body.UpdateNames(mapping)

		for i, arg := range function.Args {
			if updated, exists := mapping[arg]; exists {
				function.Args[i] = updated
			}
		}
	}

	for _, function := range append(functions, &ir.Function{
		Name: "main",
		Args: nil,
		Body: main,
	}) {
		allocate(function)
	}

	return spills
}
