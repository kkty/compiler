package emit

import (
	"fmt"
	"github.com/kkty/compiler/ir"
	"github.com/kkty/compiler/stringset"
	"github.com/kkty/compiler/typing"
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

// AllocateRegisters does register allocation.
// Variable names in ir.Node are replaced with register names like $i1.
// Variables that are never referenced are renamed to "".
// If a variable could not be assigned to any registers, its name will be kept unchanged
// and should be saved on the stack.
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
				if !stringset.Set(n.Next.FreeVariables(stringset.New())).Has(n.Name) {
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
				v := stringset.Set(node.FreeVariables(stringset.New()))
				restore := v.Join(variablesToKeep)
				addEdges(v)
				restore(v)
				return v
			}
		}

		addEdges(liveVariables(function.Body, stringset.New()))

		mapping := map[string]string{}

		assign := func(variable string) {
			if variable == "" {
				return
			}

			var graph map[string]stringset.Set
			var registers []string

			if _, ok := types[variable].(*typing.FloatType); ok {
				graph, registers = floatGraph, floatRegisters
			} else {
				graph, registers = intGraph, intRegisters
			}

			unavailable := stringset.New()
			for _, adjacent := range graph[variable].Slice() {
				if register, exists := mapping[adjacent]; exists {
					unavailable.Add(register)
				}
			}

			mapped := false

			for _, register := range registers {
				if !unavailable.Has(register) {
					mapping[variable] = register
					mapped = true
					break
				}
			}

			if !mapped {
				spills[function.Name]++
			}
		}

		// visit all nodes using dfs.
		var visit func(ir.Node)
		visit = func(node ir.Node) {
			switch node.(type) {
			case *ir.IfEqual:
				n := node.(*ir.IfEqual)
				visit(n.True)
				visit(n.False)
			case *ir.IfEqualZero:
				n := node.(*ir.IfEqualZero)
				visit(n.True)
				visit(n.False)
			case *ir.IfEqualTrue:
				n := node.(*ir.IfEqualTrue)
				visit(n.True)
				visit(n.False)
			case *ir.IfLessThan:
				n := node.(*ir.IfLessThan)
				visit(n.True)
				visit(n.False)
			case *ir.IfLessThanZero:
				n := node.(*ir.IfLessThanZero)
				visit(n.True)
				visit(n.False)
			case *ir.Assignment:
				n := node.(*ir.Assignment)
				if n.Name != "" {
					assign(n.Name)
				}
				visit(n.Value)
				visit(n.Next)
			}
		}

		for _, arg := range function.Args {
			assign(arg)
		}

		visit(function.Body)

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
