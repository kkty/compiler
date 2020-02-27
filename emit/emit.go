package emit

import (
	"fmt"
	"io"
	"log"
	"math"
	"strings"

	"github.com/kkty/compiler/ir"
	"github.com/kkty/compiler/stringset"
	"github.com/kkty/compiler/typing"
	"github.com/thoas/go-funk"
)

const (
	heapPointer              = "$hp"
	stackPointer             = "$sp"
	returnAddressPointer     = "$ra"
	zeroRegister             = "$zero"
	returnRegister           = "$r54"
	initialStackPointerValue = 210000
	initialHeapPointerValue  = 240000
)

var (
	argRegisters       = []string{"$r55", "$r56", "$r57"}
	temporaryRegisters = []string{"$r58", "$r59"}
)

// Emit emits assembly code from IR.
func Emit(functions []*ir.Function, main ir.Node, globals map[string]ir.Node, types map[string]typing.Type, w io.Writer) {
	nextLabelId := 0
	getLabel := func() string {
		defer func() { nextLabelId++ }()
		return fmt.Sprintf("L%d", nextLabelId)
	}

	isRegister := func(s string) bool {
		return strings.HasPrefix(s, "$")
	}

	// load variables to registers if necessary
	// intArgRegisters/floatArgRegisters are used
	loadVariables := func(variables, storedVariables []string) []string {
		registers := []string{}
		nextArgRegister := 0
		for _, variable := range variables {
			if isRegister(variable) {
				registers = append(registers, variable)
			} else {
				idx := funk.IndexOfString(storedVariables, variable)
				if idx == -1 {
					log.Panicf("variable not found on stack: %s", variable)
				}
				register := argRegisters[nextArgRegister]
				nextArgRegister++
				fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", register, idx, zeroRegister, stackPointer)
				registers = append(registers, register)
			}
		}
		return registers
	}

	// find function by name
	findFunction := func(name string) *ir.Function {
		for _, function := range functions {
			if function.Name == name {
				return function
			}
		}
		return nil
	}

	// registers used in a function
	functionToRegisters := map[string]stringset.Set{}

	// spilled variables in a function
	functionToSpills := map[string][]string{}

	// functions that are called in a function
	functionToDependencies := map[string]stringset.Set{}

	for _, function := range append(functions, &ir.Function{
		Name: "main",
		Args: nil,
		Body: main,
	}) {
		functionToRegisters[function.Name] = stringset.New()
		functionToDependencies[function.Name] = stringset.New()

		// add items to functionToRegisters/functionToSpills
		addVariables := func(variables []string) {
			for _, variable := range variables {
				if variable == "" {
					continue
				}
				if isRegister(variable) {
					functionToRegisters[function.Name].Add(variable)
				} else {
					if !funk.ContainsString(functionToSpills[function.Name], variable) {
						functionToSpills[function.Name] = append(functionToSpills[function.Name], variable)
					}
				}
			}
		}

		addVariables(function.Args)

		// use bfs
		queue := []ir.Node{function.Body}
		for len(queue) > 0 {
			n := queue[0]
			queue = queue[1:]
			switch n := n.(type) {
			case *ir.IfEqual:
				queue = append(queue, n.True, n.False)
			case *ir.IfEqualZero:
				queue = append(queue, n.True, n.False)
			case *ir.IfEqualTrue:
				queue = append(queue, n.True, n.False)
			case *ir.IfLessThan:
				queue = append(queue, n.True, n.False)
			case *ir.IfLessThanFloat:
				queue = append(queue, n.True, n.False)
			case *ir.IfLessThanZero:
				queue = append(queue, n.True, n.False)
			case *ir.IfLessThanZeroFloat:
				queue = append(queue, n.True, n.False)
			case *ir.Assignment:
				addVariables([]string{n.Name})
				queue = append(queue, n.Value, n.Next)
			case *ir.Application:
				functionToDependencies[function.Name].Add(n.Function)
			}
		}
	}

	// update functionToRegisters so that function applications are considered
	// i.e. if function A uses register X and function B calls function A,
	// function B uses register X.
	for {
		updated := false
		for _, function := range append(functions, &ir.Function{
			Name: "main",
		}) {
			before := len(functionToRegisters[function.Name].Slice())
			for _, dependency := range functionToDependencies[function.Name].Slice() {
				functionToRegisters[function.Name].Join(functionToRegisters[dependency])
			}
			after := len(functionToRegisters[function.Name].Slice())
			if before != after {
				updated = true
			}
		}
		if !updated {
			break
		}
	}

	// floatValues[i] will later be saved to memory[i] and loaded when necessary
	floatValues := []float32{}
	for _, function := range append(functions, &ir.Function{
		Name: "main",
		Body: main,
	}) {
		for _, floatValue := range function.Body.FloatValues() {
			if !funk.ContainsFloat32(floatValues, floatValue) {
				floatValues = append(floatValues, floatValue)
			}
		}
	}

	// global variable v will later be saved to memory[globalToPosition[v]] or globalToRegister[v]
	globalToPosition := map[string]int{}
	globalToRegister := map[string]string{}
	for name := range globals {
		if len(globalToRegister) < 30 {
			globalToRegister[name] = fmt.Sprintf("$r%d", len(globalToRegister)+len(registers))
		} else {
			globalToPosition[name] = len(globalToPosition) + len(floatValues)
		}
	}

	var emit func(string, bool, ir.Node, []string, stringset.Set)
	emit = func(
		destination string,
		tail bool,
		node ir.Node,
		variablesOnStack []string,
		registersToUse stringset.Set,
	) {
		findPosition := func(variable string) int {
			for i, v := range variablesOnStack {
				if v == variable {
					return i
				}
			}
			panic(fmt.Sprintf("variable not found on stack: %s", variable))
		}

		switch n := node.(type) {
		case *ir.Variable:
			if destination != "" {
				if register, ok := globalToRegister[n.Name]; ok {
					if isRegister(destination) {
						fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, register, zeroRegister)
					} else {
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", register, findPosition(destination), zeroRegister, stackPointer)
					}
				} else if position, ok := globalToPosition[n.Name]; ok {
					if isRegister(destination) {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", destination, position, zeroRegister, zeroRegister)
					} else {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], position, zeroRegister, zeroRegister)
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
					}
				} else if isRegister(n.Name) {
					if isRegister(destination) {
						fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, n.Name, zeroRegister)
					} else {
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", n.Name, findPosition(destination), zeroRegister, stackPointer)
					}
				} else {
					if isRegister(destination) {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", destination, findPosition(n.Name), zeroRegister, stackPointer)
					} else {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(n.Name), zeroRegister, stackPointer)
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
					}
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Unit:
			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Int:
			if destination != "" {
				if isRegister(destination) {
					fmt.Fprintf(w, "ADDI %s, %s, %d\n", destination, zeroRegister, n.Value)
				} else {
					fmt.Fprintf(w, "ADDI %s, %s, %d\n", temporaryRegisters[0], zeroRegister, n.Value)
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Bool:
			if n.Value {
				emit(destination, tail, &ir.Int{Value: 1}, variablesOnStack, registersToUse)
			} else {
				emit(destination, tail, &ir.Int{Value: 0}, variablesOnStack, registersToUse)
			}
		case *ir.Float:
			if destination != "" {
				if isRegister(destination) {
					if n.Value == 0 {
						fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, zeroRegister, zeroRegister)
					} else {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", destination, funk.IndexOf(floatValues, n.Value), zeroRegister, zeroRegister)
					}
				} else {
					if n.Value == 0 {
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", zeroRegister, findPosition(destination), zeroRegister, stackPointer)
					} else {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], funk.IndexOf(floatValues, n.Value), zeroRegister, zeroRegister)
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
					}
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Add:
			if destination != "" {
				registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
				if isRegister(destination) {
					fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, registers[0], registers[1])
				} else {
					fmt.Fprintf(w, "ADD %s, %s, %s\n", temporaryRegisters[0], registers[0], registers[1])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.AddImmediate:
			if destination != "" {
				registers := loadVariables([]string{n.Left}, variablesOnStack)
				if isRegister(destination) {
					fmt.Fprintf(w, "ADDI %s, %s, %d\n", destination, registers[0], n.Right)
				} else {
					fmt.Fprintf(w, "ADDI %s, %s, %d\n", temporaryRegisters[0], registers[0], n.Right)
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Sub:
			if destination != "" {
				registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
				if isRegister(destination) {
					fmt.Fprintf(w, "SUB %s, %s, %s\n", destination, registers[0], registers[1])
				} else {
					fmt.Fprintf(w, "SUB %s, %s, %s\n", temporaryRegisters[0], registers[0], registers[1])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.SubFromZero:
			if destination != "" {
				registers := loadVariables([]string{n.Inner}, variablesOnStack)
				if isRegister(destination) {
					fmt.Fprintf(w, "SUB %s, %s, %s\n", destination, zeroRegister, registers[0])
				} else {
					fmt.Fprintf(w, "SUB %s, %s, %s\n", temporaryRegisters[0], zeroRegister, registers[0])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.FloatAdd:
			if destination != "" {
				registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
				if isRegister(destination) {
					fmt.Fprintf(w, "ADDS %s, %s, %s\n", destination, registers[0], registers[1])
				} else {
					fmt.Fprintf(w, "ADDS %s, %s, %s\n", temporaryRegisters[0], registers[0], registers[1])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.FloatSub:
			if destination != "" {
				registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
				if isRegister(destination) {
					fmt.Fprintf(w, "SUBS %s, %s, %s\n", destination, registers[0], registers[1])
				} else {
					fmt.Fprintf(w, "SUBS %s, %s, %s\n", temporaryRegisters[0], registers[0], registers[1])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.FloatSubFromZero:
			if destination != "" {
				registers := loadVariables([]string{n.Inner}, variablesOnStack)
				if isRegister(destination) {
					fmt.Fprintf(w, "SUBS %s, %s, %s\n", destination, zeroRegister, registers[0])
				} else {
					fmt.Fprintf(w, "SUBS %s, %s, %s\n", temporaryRegisters[0], zeroRegister, registers[0])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.FloatDiv:
			if destination != "" {
				registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
				if isRegister(destination) {
					fmt.Fprintf(w, "DIVS %s, %s, %s\n", destination, registers[0], registers[1])
				} else {
					fmt.Fprintf(w, "DIVS %s, %s, %s\n", temporaryRegisters[0], registers[0], registers[1])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.FloatMul:
			if destination != "" {
				registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
				if isRegister(destination) {
					fmt.Fprintf(w, "MULS %s, %s, %s\n", destination, registers[0], registers[1])
				} else {
					fmt.Fprintf(w, "MULS %s, %s, %s\n", temporaryRegisters[0], registers[0], registers[1])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Not:
			if destination != "" {
				registers := loadVariables([]string{n.Inner}, variablesOnStack)
				fmt.Fprintf(w, "ADDI %s, %s, 1\n", temporaryRegisters[0], zeroRegister)
				if isRegister(destination) {
					fmt.Fprintf(w, "SUB %s, %s, %s\n", destination, temporaryRegisters[0], registers[0])
				} else {
					fmt.Fprintf(w, "SUB %s, %s, %s\n", temporaryRegisters[1], temporaryRegisters[0], registers[0])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[1], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Equal:
			if destination != "" {
				registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)

				if isRegister(destination) {
					fmt.Fprintf(w, "SEQ %s, %s, %s\n", destination, registers[0], registers[1])
				} else {
					fmt.Fprintf(w, "SEQ %s, %s, %s\n", temporaryRegisters[0], registers[0], registers[1])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.EqualZero:
			if destination != "" {
				registers := loadVariables([]string{n.Inner}, variablesOnStack)

				if isRegister(destination) {
					fmt.Fprintf(w, "SEQ %s, %s, %s\n", destination, registers[0], zeroRegister)
				} else {
					fmt.Fprintf(w, "SEQ %s, %s, %s\n", temporaryRegisters[0], registers[0], zeroRegister)
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.LessThan:
			if destination != "" {
				registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)

				if isRegister(destination) {
					fmt.Fprintf(w, "SLT %s, %s, %s\n", destination, registers[0], registers[1])
				} else {
					fmt.Fprintf(w, "SLT %s, %s, %s\n", temporaryRegisters[0], registers[0], registers[1])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.LessThanFloat:
			if destination != "" {
				registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)

				if isRegister(destination) {
					fmt.Fprintf(w, "SLTS %s, %s, %s\n", destination, registers[0], registers[1])
				} else {
					fmt.Fprintf(w, "SLTS %s, %s, %s\n", temporaryRegisters[0], registers[0], registers[1])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.LessThanZero:
			if destination != "" {
				registers := loadVariables([]string{n.Inner}, variablesOnStack)

				if isRegister(destination) {
					fmt.Fprintf(w, "SLT %s, %s, %s\n", destination, registers[0], zeroRegister)
				} else {
					fmt.Fprintf(w, "SLT %s, %s, %s\n", temporaryRegisters[0], registers[0], zeroRegister)
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.LessThanZeroFloat:
			if destination != "" {
				registers := loadVariables([]string{n.Inner}, variablesOnStack)

				if isRegister(destination) {
					fmt.Fprintf(w, "SLTS %s, %s, %s\n", destination, registers[0], zeroRegister)
				} else {
					fmt.Fprintf(w, "SLTS %s, %s, %s\n", temporaryRegisters[0], registers[0], zeroRegister)
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.GreaterThanZero:
			if destination != "" {
				registers := loadVariables([]string{n.Inner}, variablesOnStack)

				if isRegister(destination) {
					fmt.Fprintf(w, "SLT %s, %s, %s\n", destination, zeroRegister, registers[0])
				} else {
					fmt.Fprintf(w, "SLT %s, %s, %s\n", temporaryRegisters[0], zeroRegister, registers[0])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.GreaterThanZeroFloat:
			if destination != "" {
				registers := loadVariables([]string{n.Inner}, variablesOnStack)

				if isRegister(destination) {
					fmt.Fprintf(w, "SLTS %s, %s, %s\n", destination, zeroRegister, registers[0])
				} else {
					fmt.Fprintf(w, "SLTS %s, %s, %s\n", temporaryRegisters[0], zeroRegister, registers[0])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.IfEqual:
			elseLabel := getLabel()
			continueLabel := getLabel()
			registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
			fmt.Fprintf(w, "BEQ %s, %s, 1\n", registers[0], registers[1])
			fmt.Fprintf(w, "J %s\n", elseLabel)
			emit(destination, tail, n.True, variablesOnStack, registersToUse)
			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}
			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")
			emit(destination, tail, n.False, variablesOnStack, registersToUse)
			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.IfEqualZero:
			elseLabel := getLabel()
			continueLabel := getLabel()
			registers := loadVariables([]string{n.Inner}, variablesOnStack)

			fmt.Fprintf(w, "BEQ %s, %s, 1\n", registers[0], zeroRegister)

			fmt.Fprintf(w, "J %s\n", elseLabel)
			emit(destination, tail, n.True, variablesOnStack, registersToUse)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")
			emit(destination, tail, n.False, variablesOnStack, registersToUse)

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.IfEqualTrue:
			elseLabel := getLabel()
			continueLabel := getLabel()
			registers := loadVariables([]string{n.Inner}, variablesOnStack)

			fmt.Fprintf(w, "BLT %s, %s, 1\n", zeroRegister, registers[0])

			fmt.Fprintf(w, "J %s\n", elseLabel)
			emit(destination, tail, n.True, variablesOnStack, registersToUse)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")
			emit(destination, tail, n.False, variablesOnStack, registersToUse)

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.IfLessThan:
			elseLabel := getLabel()
			continueLabel := getLabel()
			registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)

			fmt.Fprintf(w, "BLT %s, %s, 1\n", registers[0], registers[1])

			fmt.Fprintf(w, "J %s\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.True, variablesOnStack, registersToUse)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.False, variablesOnStack, registersToUse)

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.IfLessThanFloat:
			elseLabel := getLabel()
			continueLabel := getLabel()
			registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)

			fmt.Fprintf(w, "BLTS %s, %s, 1\n", registers[0], registers[1])

			fmt.Fprintf(w, "J %s\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.True, variablesOnStack, registersToUse)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.False, variablesOnStack, registersToUse)

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.IfLessThanZero:
			elseLabel := getLabel()
			continueLabel := getLabel()
			registers := loadVariables([]string{n.Inner}, variablesOnStack)

			fmt.Fprintf(w, "BLT %s, %s, 1\n", registers[0], zeroRegister)

			fmt.Fprintf(w, "J %s\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.True, variablesOnStack, registersToUse)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.False, variablesOnStack, registersToUse)

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.IfLessThanZeroFloat:
			elseLabel := getLabel()
			continueLabel := getLabel()
			registers := loadVariables([]string{n.Inner}, variablesOnStack)

			fmt.Fprintf(w, "BLTS %s, %s, 1\n", registers[0], zeroRegister)

			fmt.Fprintf(w, "J %s\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.True, variablesOnStack, registersToUse)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.False, variablesOnStack, registersToUse)

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.Assignment:
			registers := stringset.New()
			for v := range n.Next.FreeVariables(stringset.NewFromSlice(func() []string {
				if isRegister(n.Name) {
					return []string{n.Name}
				}
				return nil
			}())) {
				if isRegister(v) {
					registers.Add(v)
				}
			}
			restore := registersToUse.Join(registers)
			emit(n.Name, false, n.Value, variablesOnStack, registersToUse)
			restore(registersToUse)
			emit(destination, tail, n.Next, variablesOnStack, registersToUse)
		case *ir.Application:
			f := findFunction(n.Function)

			// restore values on the registers that will be used by the callee
			var registersToSave []string
			if !tail {
				for _, register := range functionToRegisters[n.Function].Slice() {
					if registersToUse.Has(register) {
						registersToSave = append(registersToSave, register)
					}
				}
				for i, register := range registersToSave {
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n",
						register, (len(variablesOnStack) + i), zeroRegister, stackPointer)
				}
			}

			// move values among registers and stack

			registerToRegister := map[string]map[string]struct{}{}
			registerToMemory := map[string]map[int]struct{}{}
			memoryToRegister := map[int]map[string]struct{}{}
			memoryToMemory := map[int]map[int]struct{}{}
			globalMemoryToRegister := map[int]map[string]struct{}{}
			globalMemoryToMemory := map[int]map[int]struct{}{}
			globalRegisterToRegister := map[string]map[string]struct{}{}
			globalRegisterToMemory := map[string]map[int]struct{}{}

			findPositionInF := func(variable string) int {
				idx := funk.IndexOfString(functionToSpills[f.Name], variable)
				if idx == -1 {
					log.Panicf("variable not found: %s", variable)
				}
				return idx
			}

			for i, arg := range f.Args {
				if arg == "" {
					continue
				}
				if from, ok := globalToRegister[n.Args[i]]; ok {
					if isRegister(arg) {
						to := arg
						if from != to {
							if _, exists := globalRegisterToRegister[from]; !exists {
								globalRegisterToRegister[from] = map[string]struct{}{}
							}
							globalRegisterToRegister[from][to] = struct{}{}
						}
					} else {
						var to int
						if tail {
							to = findPositionInF(arg)
						} else {
							to = len(variablesOnStack) + len(registersToSave) + 1 + findPositionInF(arg)
						}
						if _, exists := globalRegisterToMemory[from]; !exists {
							globalRegisterToMemory[from] = map[int]struct{}{}
						}
						globalRegisterToMemory[from][to] = struct{}{}
					}
				} else if from, ok := globalToPosition[n.Args[i]]; ok {
					if isRegister(arg) {
						to := arg
						if _, exists := globalMemoryToRegister[from]; !exists {
							globalMemoryToRegister[from] = map[string]struct{}{}
						}
						globalMemoryToRegister[from][to] = struct{}{}
					} else {
						var to int
						if tail {
							to = findPositionInF(arg)
						} else {
							to = len(variablesOnStack) + len(registersToSave) + 1 + findPositionInF(arg)
						}
						if from != to {
							if _, exists := globalMemoryToMemory[from]; !exists {
								globalMemoryToMemory[from] = map[int]struct{}{}
							}
							globalMemoryToMemory[from][to] = struct{}{}
						}
					}
				} else if isRegister(n.Args[i]) {
					from := n.Args[i]
					if isRegister(arg) {
						to := arg
						if from != to {
							if _, exists := registerToRegister[from]; !exists {
								registerToRegister[from] = map[string]struct{}{}
							}
							registerToRegister[from][to] = struct{}{}
						}
					} else {
						var to int
						if tail {
							to = findPositionInF(arg)
						} else {
							to = len(variablesOnStack) + len(registersToSave) + 1 + findPositionInF(arg)
						}
						if _, exists := registerToMemory[from]; !exists {
							registerToMemory[from] = map[int]struct{}{}
						}
						registerToMemory[from][to] = struct{}{}
					}
				} else {
					from := findPosition(n.Args[i])
					if isRegister(arg) {
						to := arg
						if _, exists := memoryToRegister[from]; !exists {
							memoryToRegister[from] = map[string]struct{}{}
						}
						memoryToRegister[from][to] = struct{}{}
					} else {
						var to int
						if tail {
							to = findPositionInF(arg)
						} else {
							to = len(variablesOnStack) + len(registersToSave) + 1 + findPositionInF(arg)
						}
						if from != to {
							if _, exists := memoryToMemory[from]; !exists {
								memoryToMemory[from] = map[int]struct{}{}
							}
							memoryToMemory[from][to] = struct{}{}
						}
					}
				}
			}

			// this is used to break cycles
			after := []func(){}

			for len(registerToRegister)+len(registerToMemory)+len(memoryToRegister)+len(memoryToMemory)+len(globalMemoryToRegister)+len(globalMemoryToMemory)+len(globalRegisterToRegister)+len(globalRegisterToMemory) > 0 {
				updated := func() bool {
					for from, tos := range registerToRegister {
						for to := range tos {
							if _, exists := registerToRegister[to]; !exists {
								if _, exists := registerToMemory[to]; !exists {
									fmt.Fprintf(w, "ADD %s, %s, %s\n", to, from, zeroRegister)
									delete(registerToRegister[from], to)
									if len(registerToRegister[from]) == 0 {
										delete(registerToRegister, from)
									}
									return true
								}
							}
						}
					}
					for from, tos := range registerToMemory {
						for to := range tos {
							if _, exists := memoryToRegister[to]; !exists {
								if _, exists := memoryToMemory[to]; !exists {
									fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", from, to, zeroRegister, stackPointer)
									delete(registerToMemory[from], to)
									if len(registerToMemory[from]) == 0 {
										delete(registerToMemory, from)
									}
									return true
								}
							}
						}
					}
					for from, tos := range memoryToRegister {
						for to := range tos {
							if _, exists := registerToRegister[to]; !exists {
								if _, exists := registerToMemory[to]; !exists {
									fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", to, from, zeroRegister, stackPointer)
									delete(memoryToRegister[from], to)
									if len(memoryToRegister[from]) == 0 {
										delete(memoryToRegister, from)
									}
									return true
								}
							}
						}
					}
					for from, tos := range memoryToMemory {
						for to := range tos {
							if _, exists := memoryToRegister[to]; !exists {
								if _, exists := memoryToMemory[to]; !exists {
									fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], from, zeroRegister, stackPointer)
									fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], to, zeroRegister, stackPointer)
									delete(memoryToMemory[from], to)
									if len(memoryToMemory[from]) == 0 {
										delete(memoryToMemory, from)
									}
									return true
								}
							}
						}
					}
					for from, tos := range globalMemoryToRegister {
						for to := range tos {
							if _, exists := registerToRegister[to]; !exists {
								if _, exists := registerToMemory[to]; !exists {
									fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", to, from, zeroRegister, zeroRegister)
									delete(globalMemoryToRegister[from], to)
									if len(globalMemoryToRegister[from]) == 0 {
										delete(globalMemoryToRegister, from)
									}
									return true
								}
							}
						}
					}
					for from, tos := range globalMemoryToMemory {
						for to := range tos {
							if _, exists := memoryToRegister[to]; !exists {
								if _, exists := memoryToMemory[to]; !exists {
									fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], from, zeroRegister, zeroRegister)
									fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], to, zeroRegister, stackPointer)
									delete(globalMemoryToMemory[from], to)
									if len(globalMemoryToMemory[from]) == 0 {
										delete(globalMemoryToMemory, from)
									}
									return true
								}
							}
						}
					}
					for from, tos := range globalRegisterToRegister {
						for to := range tos {
							if _, exists := registerToRegister[to]; !exists {
								if _, exists := registerToMemory[to]; !exists {
									fmt.Fprintf(w, "ADD %s, %s, %s\n", to, from, zeroRegister)
									delete(globalRegisterToRegister[from], to)
									if len(globalRegisterToRegister[from]) == 0 {
										delete(globalRegisterToRegister, from)
									}
									return true
								}
							}
						}
					}
					for from, tos := range globalRegisterToMemory {
						for to := range tos {
							if _, exists := memoryToRegister[to]; !exists {
								if _, exists := memoryToMemory[to]; !exists {
									fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", from, to, zeroRegister, stackPointer)
									delete(globalRegisterToMemory[from], to)
									if len(globalRegisterToMemory[from]) == 0 {
										delete(globalRegisterToMemory, from)
									}
									return true
								}
							}
						}
					}
					return false
				}()

				if !updated {
					// break a cycle by using heap
					func() {
						for from, tos := range registerToRegister {
							idx := len(after)
							fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", from, idx, zeroRegister, heapPointer)
							for to := range tos {
								after = append(after, func() {
									fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", to, idx, zeroRegister, heapPointer)
								})
							}
							delete(registerToRegister, from)
							return
						}
						for from, tos := range registerToMemory {
							idx := len(after)
							fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", from, idx, zeroRegister, heapPointer)
							for to := range tos {
								after = append(after, func() {
									fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], idx, zeroRegister, heapPointer)
									fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], to, zeroRegister, stackPointer)
								})
							}
							delete(registerToMemory, from)
							return
						}
						for from, tos := range memoryToMemory {
							idx := len(after)
							fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], from, zeroRegister, stackPointer)
							fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], idx, zeroRegister, heapPointer)
							for to := range tos {
								after = append(after, func() {
									fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], idx, zeroRegister, heapPointer)
									fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], to, zeroRegister, stackPointer)
								})
							}
							delete(memoryToMemory, from)
							return
						}
					}()
				}
			}

			for _, fn := range after {
				fn()
			}

			if tail {
				fmt.Fprintf(w, "J %s\n", n.Function)
			} else {
				fmt.Fprintf(w, "SW %s, %d(%s, %s)\n",
					returnAddressPointer, (len(variablesOnStack) + len(registersToSave)), zeroRegister, stackPointer)

				fmt.Fprintf(w, "ADDI %s, %s, %d\n",
					stackPointer, stackPointer, (len(variablesOnStack) + len(registersToSave) + 1))

				fmt.Fprintf(w, "JAL %s\n", n.Function)

				fmt.Fprintf(w, "ADDI %s, %s, %d\n",
					stackPointer, stackPointer, -(len(variablesOnStack) + len(registersToSave) + 1))

				fmt.Fprintf(w, "LW %s, %d(%s, %s)\n",
					returnAddressPointer, (len(variablesOnStack) + len(registersToSave)), zeroRegister, stackPointer)

				// restore registers
				for i, register := range registersToSave {
					fmt.Fprintf(w, "LW %s, %d(%s, %s)\n",
						register, (len(variablesOnStack) + i), zeroRegister, stackPointer)
				}

				if destination != "" {
					if isRegister(destination) {
						fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, returnRegister, zeroRegister)
					} else {
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", returnRegister, findPosition(destination), zeroRegister, stackPointer)
					}
				}
			}
		case *ir.Tuple:
			if destination != "" {
				for i, element := range n.Elements {
					registers := loadVariables([]string{element}, variablesOnStack)
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", registers[0], i, zeroRegister, heapPointer)
				}

				if isRegister(destination) {
					fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, heapPointer, zeroRegister)
				} else {
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", heapPointer, findPosition(destination), zeroRegister, stackPointer)
				}

				fmt.Fprintf(w, "ADDI %s, %s, %d\n", heapPointer, heapPointer, len(n.Elements))
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.TupleGet:
			if destination != "" {
				if register, ok := globalToRegister[n.Tuple]; ok {
					if isRegister(destination) {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", destination, n.Index, zeroRegister, register)
					} else {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], n.Index, zeroRegister, register)
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
					}
				} else if position, ok := globalToPosition[n.Tuple]; ok {
					fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], position, zeroRegister, zeroRegister)

					if isRegister(destination) {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", destination, n.Index, zeroRegister, temporaryRegisters[0])
					} else {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[1], n.Index, zeroRegister, temporaryRegisters[0])
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[1], findPosition(destination), zeroRegister, stackPointer)
					}
				} else {
					registers := loadVariables([]string{n.Tuple}, variablesOnStack)

					if isRegister(destination) {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", destination, n.Index, zeroRegister, registers[0])
					} else {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], n.Index, zeroRegister, registers[0])
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
					}
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayCreate:
			if destination != "" {
				registers := loadVariables([]string{n.Length, n.Value}, variablesOnStack)

				fmt.Fprintf(w, "ADD %s, %s, %s\n",
					temporaryRegisters[0], registers[0], zeroRegister)

				fmt.Fprintf(w, "ADD %s, %s, %s\n",
					temporaryRegisters[1], registers[1], zeroRegister)

				if isRegister(destination) {
					fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, heapPointer, zeroRegister)
				} else {
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", heapPointer, findPosition(destination), zeroRegister, stackPointer)
				}

				loopLabel := getLabel()

				fmt.Fprintf(w, "%s:\n", loopLabel)

				fmt.Fprintf(w, "BEQ %s, %s, 4\n",
					temporaryRegisters[0], zeroRegister)

				fmt.Fprintf(w, "SW %s, 0(%s, %s)\n",
					temporaryRegisters[1], zeroRegister, heapPointer)

				fmt.Fprintf(w, "ADDI %s, %s, 1\n", heapPointer, heapPointer)

				fmt.Fprintf(w, "ADDI %s, %s, -1\n",
					temporaryRegisters[0], temporaryRegisters[0])

				fmt.Fprintf(w, "J %s\n", loopLabel)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayCreateImmediate:
			if destination != "" {
				registers := loadVariables([]string{n.Value}, variablesOnStack)

				fmt.Fprintf(w, "ADD %s, %s, %s\n",
					temporaryRegisters[0], registers[0], zeroRegister)

				if isRegister(destination) {
					fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, heapPointer, zeroRegister)
				} else {
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", heapPointer, findPosition(destination), zeroRegister, stackPointer)
				}

				for i := 0; i < int(n.Length); i++ {
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n",
						temporaryRegisters[0], i, zeroRegister, heapPointer)
				}

				fmt.Fprintf(w, "ADDI %s, %s, %d\n", heapPointer, heapPointer, n.Length)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayGet:
			if destination != "" {
				if register, ok := globalToRegister[n.Array]; ok {
					registers := loadVariables([]string{n.Index}, variablesOnStack)
					if isRegister(destination) {
						fmt.Fprintf(w, "LW %s, 0(%s, %s)\n", destination, registers[0], register)
					} else {
						fmt.Fprintf(w, "LW %s, 0(%s, %s)\n", temporaryRegisters[0], registers[0], register)
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
					}
				} else if position, ok := globalToPosition[n.Array]; ok {
					registers := loadVariables([]string{n.Index}, variablesOnStack)
					fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], position, zeroRegister, zeroRegister)

					if isRegister(destination) {
						fmt.Fprintf(w, "LW %s, 0(%s, %s)\n", destination, temporaryRegisters[0], registers[0])
					} else {
						fmt.Fprintf(w, "LW %s, 0(%s, %s)\n", temporaryRegisters[1], temporaryRegisters[0], registers[0])
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[1], findPosition(destination), zeroRegister, stackPointer)
					}
				} else {
					registers := loadVariables([]string{n.Array, n.Index}, variablesOnStack)

					if isRegister(destination) {
						fmt.Fprintf(w, "LW %s, 0(%s, %s)\n", destination, registers[0], registers[1])
					} else {
						fmt.Fprintf(w, "LW %s, 0(%s, %s)\n", temporaryRegisters[0], registers[0], registers[1])
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
					}
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayGetImmediate:
			if destination != "" {
				if register, ok := globalToRegister[n.Array]; ok {
					if isRegister(destination) {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", destination, n.Index, zeroRegister, register)
					} else {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], n.Index, zeroRegister, register)
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
					}
				} else if position, ok := globalToPosition[n.Array]; ok {
					fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], position, zeroRegister, zeroRegister)

					if isRegister(destination) {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", destination, n.Index, zeroRegister, temporaryRegisters[0])
					} else {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[1], n.Index, zeroRegister, temporaryRegisters[0])
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[1], findPosition(destination), zeroRegister, stackPointer)
					}
				} else {
					registers := loadVariables([]string{n.Array}, variablesOnStack)

					if isRegister(destination) {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", destination, n.Index, zeroRegister, registers[0])
					} else {
						fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], n.Index, zeroRegister, registers[0])
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
					}
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayPut:
			if register, ok := globalToRegister[n.Array]; ok {
				registers := loadVariables([]string{n.Index, n.Value}, variablesOnStack)

				fmt.Fprintf(w, "SW %s, 0(%s, %s)\n", registers[1], register, registers[0])

				if tail {
					fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
				}
			} else if position, ok := globalToPosition[n.Array]; ok {
				registers := loadVariables([]string{n.Index, n.Value}, variablesOnStack)

				fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], position, zeroRegister, zeroRegister)
				fmt.Fprintf(w, "SW %s, 0(%s, %s)\n", registers[1], temporaryRegisters[0], registers[0])

				if tail {
					fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
				}
			} else {
				registers := loadVariables([]string{n.Array, n.Index, n.Value}, variablesOnStack)

				fmt.Fprintf(w, "SW %s, 0(%s, %s)\n", registers[2], registers[0], registers[1])

				if tail {
					fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
				}
			}
		case *ir.ArrayPutImmediate:
			if register, ok := globalToRegister[n.Array]; ok {
				registers := loadVariables([]string{n.Value}, variablesOnStack)

				fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", registers[0], n.Index, zeroRegister, register)

				if tail {
					fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
				}

			} else if position, ok := globalToPosition[n.Array]; ok {
				fmt.Fprintf(w, "LW %s, %d(%s, %s)\n", temporaryRegisters[0], position, zeroRegister, zeroRegister)

				registers := loadVariables([]string{n.Value}, variablesOnStack)

				fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", registers[0], n.Index, zeroRegister, temporaryRegisters[0])

				if tail {
					fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
				}
			} else {
				registers := loadVariables([]string{n.Array, n.Value}, variablesOnStack)

				fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", registers[1], n.Index, zeroRegister, registers[0])

				if tail {
					fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
				}
			}
		case *ir.ReadInt:
			if destination == "" {
				fmt.Fprintf(w, "IN %s\n", temporaryRegisters[0])
			} else if isRegister(destination) {
				fmt.Fprintf(w, "IN %s\n", destination)
			} else {
				fmt.Fprintf(w, "IN %s\n", temporaryRegisters[0])
				fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ReadFloat:
			if destination == "" {
				fmt.Fprintf(w, "INF %s\n", temporaryRegisters[0])
			} else if isRegister(destination) {
				fmt.Fprintf(w, "INF %s\n", destination)
			} else {
				fmt.Fprintf(w, "INF %s\n", temporaryRegisters[0])
				fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.WriteByte:
			registers := loadVariables([]string{n.Arg}, variablesOnStack)

			fmt.Fprintf(w, "OUT %s\n", registers[0])

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.IntToFloat:
			if destination != "" {
				registers := loadVariables([]string{n.Arg}, variablesOnStack)

				if isRegister(destination) {
					fmt.Fprintf(w, "ITOF %s, %s\n", destination, registers[0])
				} else {
					fmt.Fprintf(w, "ITOF %s, %s\n", temporaryRegisters[0], registers[0])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.FloatToInt:
			if destination != "" {
				registers := loadVariables([]string{n.Arg}, variablesOnStack)

				if isRegister(destination) {
					fmt.Fprintf(w, "FTOI %s, %s\n", destination, registers[0])
				} else {
					fmt.Fprintf(w, "FTOI %s, %s\n", temporaryRegisters[0], registers[0])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Sqrt:
			if destination != "" {
				registers := loadVariables([]string{n.Arg}, variablesOnStack)

				if isRegister(destination) {
					fmt.Fprintf(w, "SQRT %s, %s\n", destination, registers[0])
				} else {
					fmt.Fprintf(w, "SQRT %s, %s\n", temporaryRegisters[0], registers[0])
					fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], findPosition(destination), zeroRegister, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		default:
			log.Panic("invalid node")
		}
	}

	// set stack pointer to 210000
	fmt.Fprintf(w, "LUI %s, %s, %d\n", stackPointer, zeroRegister, initialStackPointerValue>>16)
	fmt.Fprintf(w, "ORI %s, %s, %d\n", stackPointer, stackPointer, initialStackPointerValue%(1<<16))

	// set heap pointer to 240000
	fmt.Fprintf(w, "LUI %s, %s, %d\n", heapPointer, zeroRegister, initialHeapPointerValue>>16)
	fmt.Fprintf(w, "ORI %s, %s, %d\n", heapPointer, heapPointer, initialHeapPointerValue%(1<<16))

	// save float values to memory
	for i, value := range floatValues {
		u := math.Float32bits(value)
		fmt.Fprintf(w, "ORI %s, %s, %d\n", temporaryRegisters[0], zeroRegister, u%(1<<16))
		fmt.Fprintf(w, "LUI %s, %s, %d\n", temporaryRegisters[0], temporaryRegisters[0], u>>16)
		fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", temporaryRegisters[0], i, zeroRegister, zeroRegister)
	}

	// calculate global variables and save them to memory
	{
		// As global variables may use another global variable in their definitions,
		// we have to be careful about their order here.
		defined := stringset.New()
		for len(defined) < len(globals) {
			for name, node := range globals {
				if !defined.Has(name) && len(node.FreeVariables(defined)) == 0 {
					emit(returnRegister, false, node, nil, stringset.New())
					if register, ok := globalToRegister[name]; ok {
						fmt.Fprintf(w, "ADD %s, %s, %s\n", register, returnRegister, zeroRegister)
					} else {
						fmt.Fprintf(w, "SW %s, %d(%s, %s)\n", returnRegister, globalToPosition[name], zeroRegister, zeroRegister)
					}
					defined.Add(name)
				}
			}
		}
	}

	fmt.Fprintf(w, "JAL main\n")
	fmt.Fprintf(w, "EXIT\n")

	for _, function := range append(functions, &ir.Function{
		Name: "main",
		Args: nil,
		Body: main,
	}) {
		fmt.Fprintf(w, "%s:\n", function.Name)
		emit(returnRegister, true, function.Body, functionToSpills[function.Name], stringset.New())
	}
}
