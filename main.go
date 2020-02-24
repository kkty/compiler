package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/kkty/compiler/ast"
	"github.com/kkty/compiler/emit"
	"github.com/kkty/compiler/ir"
	"github.com/kkty/compiler/parser"
)

func main() {
	interpret := flag.Bool("i", false, "interprets program instead of generating assembly")
	debug := flag.Bool("debug", false, "enables debugging output")
	graph := flag.Bool("graph", false, "outputs graph in dot format")
	inline := flag.Int("inline", 0, "number of inline expansions")
	iter := flag.Int("iter", 0, "number of iterations for optimization")

	flag.Parse()

	b, err := ioutil.ReadFile(flag.Arg(0))

	if err != nil {
		log.Fatal(err)
	}

	root := parser.Parse(string(b))
	ast.AlphaTransform(root)
	types := ast.GetTypes(root)

	main, functions, _ := ir.Generate(root, types)

	main, functions = ir.Inline(main, functions, *inline, types, *debug)

	for i := 0; i < *iter; i++ {
		if *debug {
			fmt.Fprintf(os.Stderr, "optimizing (i=%d)\n", i)
		}

		main = ir.RemoveRedundantAssignments(main, functions)
		main = ir.RemoveRedundantVariables(main, functions)
		main = ir.Immediate(main, functions)
		main = ir.Reorder(main, functions)

		if *debug {
			cnt := 0
			for _, function := range functions {
				cnt += function.Body.Size()
			}
			fmt.Fprintf(os.Stderr, "program size = %d\n", cnt)
		}
	}

	spills := emit.AllocateRegisters(main, functions, types)
	if *debug {
		for function, count := range spills {
			fmt.Fprintf(os.Stderr, "spilled %d variables in %s\n", count, function)
		}
	}

	if *graph {
		ir.GenerateGraph(main, functions)
	} else if *interpret {
		evaluated, called := ir.Execute(functions, main, os.Stdout, os.Stdin)
		if *debug {
			print := func(m map[string]int) {
				keys := []string{}
				for key := range m {
					keys = append(keys, key)
				}
				sort.Slice(keys, func(i, j int) bool {
					return m[keys[i]] > m[keys[j]]
				})
				for _, key := range keys {
					fmt.Fprintf(os.Stderr, "%s: %d\n", key, m[key])
				}

			}

			print(evaluated)
			print(called)
		}
	} else {
		emit.Emit(functions, main, types, os.Stdout)
	}
}
