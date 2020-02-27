package emit

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"

	"github.com/kkty/compiler/ir"
	"github.com/kkty/compiler/stringmap"
	"github.com/kkty/compiler/stringset"
	"github.com/kkty/compiler/typing"
	"github.com/thoas/go-funk"
)

// NumRegisters is the number of available general registers.
const NumRegisters = 24

var (
	registers []string
)

func init() {
	// general registers are named "$r0", "$r1", ...
	for i := 0; i < NumRegisters; i++ {
		registers = append(registers, fmt.Sprintf("$r%d", i))
	}
}

// colorGraph colors a graph with k colors (0, 1, ... k - 1) using Welsh–Powell algorithm.
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
// Variable names in ir.Node are replaced with register names like $r0.
// Variables that are never referenced are renamed to "".
// If a variable could not be assigned to any registers, its name will be kept unchanged
// and should be saved on the stack.
// References to global variables are kept intact.
// The number of spills for each function is returned.
func AllocateRegisters(main ir.Node, functions []*ir.Function, globals map[string]ir.Node, types map[string]typing.Type) map[string]int {
	globalNames := stringset.New()
	for n := range globals {
		globalNames.Add(n)
	}

	spills := map[string]int{}

	allocate := func(function *ir.Function) {
		graph := map[string]stringset.Set{}

		addEdges := func(variables stringset.Set) {
			for _, i := range variables.Slice() {
				if globalNames.Has(i) {
					continue
				}
				if _, exists := graph[i]; !exists {
					graph[i] = stringset.New()
				}
				for _, j := range variables.Slice() {
					if globalNames.Has(j) {
						continue
					}
					if i != j {
						graph[i].Add(j)
					}
				}
			}
		}

		// liveVariables returns live variables at a node.
		// At the same time, the interference graphs are constructed and the variables that are never referenced
		// are renamed to "".
		var liveVariables func(ir.Node, stringset.Set) stringset.Set
		liveVariables = func(node ir.Node, variablesToKeep stringset.Set) stringset.Set {
			switch n := node.(type) {
			case *ir.IfEqual:
				v := stringset.NewFromSlice([]string{n.Left, n.Right})
				v.Join(liveVariables(n.True, variablesToKeep))
				v.Join(liveVariables(n.False, variablesToKeep))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			case *ir.IfEqualZero:
				v := stringset.NewFromSlice([]string{n.Inner})
				v.Join(liveVariables(n.True, variablesToKeep))
				v.Join(liveVariables(n.False, variablesToKeep))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			case *ir.IfEqualTrue:
				v := stringset.NewFromSlice([]string{n.Inner})
				v.Join(liveVariables(n.True, variablesToKeep))
				v.Join(liveVariables(n.False, variablesToKeep))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			case *ir.IfLessThan:
				v := stringset.NewFromSlice([]string{n.Left, n.Right})
				v.Join(liveVariables(n.True, variablesToKeep))
				v.Join(liveVariables(n.False, variablesToKeep))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			case *ir.IfLessThanFloat:
				v := stringset.NewFromSlice([]string{n.Left, n.Right})
				v.Join(liveVariables(n.True, variablesToKeep))
				v.Join(liveVariables(n.False, variablesToKeep))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			case *ir.IfLessThanZero:
				v := stringset.NewFromSlice([]string{n.Inner})
				v.Join(liveVariables(n.True, variablesToKeep))
				v.Join(liveVariables(n.False, variablesToKeep))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			case *ir.IfLessThanZeroFloat:
				v := stringset.NewFromSlice([]string{n.Inner})
				v.Join(liveVariables(n.True, variablesToKeep))
				v.Join(liveVariables(n.False, variablesToKeep))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			case *ir.Assignment:
				if !n.Next.FreeVariables(stringset.New()).Has(n.Name) {
					n.Name = ""
				}
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
				v := node.FreeVariables(stringset.New())
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

		// variable names to register names
		mapping := map[string]string{}

		for _, i := range getNodes(graph) {
			if colorMap, ok := colorGraph(graph, len(registers)); ok {
				for variable, color := range colorMap {
					mapping[variable] = registers[color]
				}
				break
			}
			removeNode(i, graph)
			spills[function.Name]++
		}

		{
			freeVariables := function.Body.FreeVariables(stringset.New())
			for i, arg := range function.Args {
				if !freeVariables.Has(arg) {
					function.Args[i] = ""
				} else if updated, exists := mapping[arg]; exists {
					function.Args[i] = updated
				}
			}
		}

		function.Body.UpdateNames(mapping)
	}

	for _, function := range append(functions, &ir.Function{
		Name: "main",
		Args: nil,
		Body: main,
	}) {
		allocate(function)
	}

	for _, node := range globals {
		allocate(&ir.Function{Body: node})
	}

	findFunction := func(name string) *ir.Function {
		for _, function := range functions {
			if function.Name == name {
				return function
			}
		}
		return nil
	}

	// estimated cost of register moves in function applications
	cost := func() int {
		count := 0
		for _, function := range append(functions, &ir.Function{
			Name: "main",
			Args: nil,
			Body: main,
		}) {
			for _, application := range function.Body.Applications() {
				f := findFunction(application.Function)
				for i, arg := range f.Args {
					if strings.HasPrefix(arg, "$") && strings.HasPrefix(application.Args[i], "$") && arg != application.Args[i] {
						count++
					}
				}
			}
		}
		return count
	}

	// randomly shuffle registers to minimize cost()
	if len(functions) > 0 {
		sinceLastImprovement := 0
		for sinceLastImprovement < 1000 {
			function := functions[rand.Int()%len(functions)]
			if len(function.Args) == 0 {
				sinceLastImprovement++
				continue
			}
			r1, r2 := registers[rand.Int()%len(registers)], registers[rand.Int()%len(registers)]
			if r1 == r2 {
				sinceLastImprovement++
				continue
			}
			swap := func() {
				tmp := "_swap_tmp"
				function.Body.UpdateNames(stringmap.Map{r1: tmp})
				if funk.ContainsString(function.Args, r1) {
					function.Args[funk.IndexOf(function.Args, r1)] = tmp
				}
				function.Body.UpdateNames(stringmap.Map{r2: r1})
				if funk.ContainsString(function.Args, r2) {
					function.Args[funk.IndexOf(function.Args, r2)] = r1
				}
				function.Body.UpdateNames(stringmap.Map{tmp: r2})
				if funk.ContainsString(function.Args, tmp) {
					function.Args[funk.IndexOf(function.Args, tmp)] = r2
				}
			}

			prevCost := cost()
			swap()
			newCost := cost()
			if newCost > prevCost {
				// rollback
				swap()
			}
			if newCost < prevCost {
				sinceLastImprovement = 0
				fmt.Fprintln(os.Stderr, cost())
			} else {
				sinceLastImprovement++
			}
		}
	}

	return spills
}
