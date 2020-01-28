package emit

import (
	"fmt"
	"github.com/kkty/compiler/stringset"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
)

func TestColorGraph(t *testing.T) {
	testColorGraph := func(t *testing.T) {
		graph := map[string]stringset.Set{}

		// creates a graph with 100 nodes

		for i := 0; i < 100; i++ {
			graph[strconv.Itoa(i)] = stringset.New()
		}

		for i := 0; i < 100; i++ {
			for j := i + 1; j < 100; j++ {
				if rand.Float64() < 0.8 {
					graph[strconv.Itoa(i)].Add(strconv.Itoa(j))
					graph[strconv.Itoa(j)].Add(strconv.Itoa(i))
				}
			}
		}

		{
			_, ok := colorGraph(graph, 100)
			assert.True(t, ok, "a graph with n nodes should be colorable with n colors")
		}

		for k := 0; k <= 100; k++ {
			if colorMap, ok := colorGraph(graph, k); ok {
				for i := 0; i < 100; i++ {
					for j := i + 1; j < 100; j++ {
						if graph[strconv.Itoa(i)].Has(strconv.Itoa(j)) {
							assert.NotEqual(t, colorMap[strconv.Itoa(i)], colorMap[strconv.Itoa(j)],
								"adjacent nodes should not have the same color")
						}
					}
				}
			}
		}
	}

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("case%d", i), testColorGraph)
	}
}
