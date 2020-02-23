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
	heapPointer          = "$hp"
	stackPointer         = "$sp"
	returnAddressPointer = "$ra"
	zeroRegister         = "$zero"
	returnRegister       = "$r50"
)

var (
	argRegisters       = []string{"$r51", "$r52", "$r53"}
	temporaryRegisters = []string{"$r54", "$r55", "$r56"}
)

// Emit emits assembly code from IR.
func Emit(functions []*ir.Function, main ir.Node, types map[string]typing.Type, w io.Writer) {
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
				fmt.Fprintf(w, "LW %s, %d(%s)\n", register, idx, stackPointer)
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

	var emit func(string, bool, ir.Node, []string, stringset.Set)
	emit = func(
		destination string,
		tail bool,
		node ir.Node,
		variablesOnStack []string,
		registersInUse stringset.Set,
	) {
		findPosition := func(variable string) int {
			for i, v := range variablesOnStack {
				if v == variable {
					return i
				}
			}
			panic("variable not found on stack")
		}

		switch n := node.(type) {
		case *ir.Variable:
			if destination != "" {
				if isRegister(n.Name) {
					if isRegister(destination) {
						fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, n.Name, zeroRegister)
					} else {
						fmt.Fprintf(w, "SW %s, %d(%s)\n", n.Name, findPosition(destination), stackPointer)
					}
				} else {
					if isRegister(destination) {
						fmt.Fprintf(w, "LW %s, %d(%s)\n", destination, findPosition(n.Name), stackPointer)
					} else {
						fmt.Fprintf(w, "LW %s, %d(%s)\n", temporaryRegisters[0], findPosition(n.Name), stackPointer)
						fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Bool:
			if n.Value {
				emit(destination, tail, &ir.Int{Value: 1}, variablesOnStack, registersInUse)
			} else {
				emit(destination, tail, &ir.Int{Value: 0}, variablesOnStack, registersInUse)
			}
		case *ir.Float:
			if destination != "" {
				u := math.Float32bits(n.Value)

				if isRegister(destination) {
					fmt.Fprintf(w, "ORI %s, %s, %d\n", destination, zeroRegister, u%(1<<16))
					fmt.Fprintf(w, "LUI %s, %s, %d\n", destination, destination, u>>16)
				} else {
					fmt.Fprintf(w, "ORI %s, %s, %d\n", temporaryRegisters[0], zeroRegister, u%(1<<16))
					fmt.Fprintf(w, "LUI %s, %s, %d\n", temporaryRegisters[0], temporaryRegisters[0], u>>16)
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[1], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
			emit(destination, tail, n.True, variablesOnStack, registersInUse)
			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}
			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")
			emit(destination, tail, n.False, variablesOnStack, registersInUse)
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
			emit(destination, tail, n.True, variablesOnStack, registersInUse)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")
			emit(destination, tail, n.False, variablesOnStack, registersInUse)

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.IfEqualTrue:
			elseLabel := getLabel()
			continueLabel := getLabel()
			registers := loadVariables([]string{n.Inner}, variablesOnStack)

			fmt.Fprintf(w, "ADDI %s, %s, 1\n", temporaryRegisters[0], zeroRegister)
			fmt.Fprintf(w, "BEQ %s, %s, 1\n", registers[0], temporaryRegisters[0])

			fmt.Fprintf(w, "J %s\n", elseLabel)
			emit(destination, tail, n.True, variablesOnStack, registersInUse)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")
			emit(destination, tail, n.False, variablesOnStack, registersInUse)

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

			emit(destination, tail, n.True, variablesOnStack, registersInUse)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.False, variablesOnStack, registersInUse)

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

			emit(destination, tail, n.True, variablesOnStack, registersInUse)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.False, variablesOnStack, registersInUse)

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

			emit(destination, tail, n.True, variablesOnStack, registersInUse)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.False, variablesOnStack, registersInUse)

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

			emit(destination, tail, n.True, variablesOnStack, registersInUse)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.False, variablesOnStack, registersInUse)

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.Assignment:
			emit(n.Name, false, n.Value, variablesOnStack, registersInUse)
			if isRegister(n.Name) {
				restore := registersInUse.Join(stringset.NewFromSlice([]string{n.Name}))
				emit(destination, tail, n.Next, variablesOnStack, registersInUse)
				restore(registersInUse)
			} else {
				emit(destination, tail, n.Next, variablesOnStack, registersInUse)
			}
		case *ir.Application:
			f := findFunction(n.Function)

			var registersToSave []string
			if tail {
				for _, arg := range f.Args {
					if isRegister(arg) {
						registersToSave = append(registersToSave, arg)
					}
				}
			} else {
				for _, register := range functionToRegisters[n.Function].Slice() {
					if registersInUse.Has(register) {
						registersToSave = append(registersToSave, register)
					}
				}
			}

			for i, register := range registersToSave {
				fmt.Fprintf(w, "SW %s, %d(%s)\n",
					register, (len(variablesOnStack) + i), stackPointer)
			}

			// pass arguments by stack
			for i, arg := range f.Args {
				if !isRegister(arg) && arg != "" {
					registers := loadVariables([]string{n.Args[i]}, variablesOnStack)
					idx := funk.IndexOfString(functionToSpills[f.Name], arg)
					if idx == -1 {
						panic("variable not found")
					}
					fmt.Fprintf(w, "SW %s, %d(%s)\n",
						registers[0], (len(variablesOnStack) + len(registersToSave) + 1 + idx), stackPointer)
				}
			}

			// pass arguments by registers
			for i, arg := range f.Args {
				if isRegister(arg) {
					if isRegister(n.Args[i]) {
						idx := funk.IndexOfString(registersToSave, n.Args[i])
						if idx == -1 {
							fmt.Fprintf(w, "ADD %s, %s, %s\n",
								arg, n.Args[i], zeroRegister)
						} else {
							fmt.Fprintf(w, "LW %s, %d(%s)\n", arg, (len(variablesOnStack) + idx), stackPointer)
						}
					} else {
						idx := funk.IndexOfString(variablesOnStack, n.Args[i])
						if idx == -1 {
							log.Panicf("variable not found: %s", n.Args[i])
						}
						fmt.Fprintf(w, "LW %s, %d(%s)\n", arg, idx, stackPointer)
					}
				}
			}

			fmt.Fprintf(w, "SW %s, %d(%s)\n",
				returnAddressPointer, (len(variablesOnStack) + len(registersToSave)), stackPointer)

			fmt.Fprintf(w, "ADDI %s, %s, %d\n",
				stackPointer, stackPointer, (len(variablesOnStack) + len(registersToSave) + 1))

			fmt.Fprintf(w, "JAL %s\n", n.Function)

			fmt.Fprintf(w, "ADDI %s, %s, %d\n",
				stackPointer, stackPointer, -(len(variablesOnStack) + len(registersToSave) + 1))

			fmt.Fprintf(w, "LW %s, %d(%s)\n",
				returnAddressPointer, (len(variablesOnStack) + len(registersToSave)), stackPointer)

			// restore registers
			if !tail {
				for i, register := range registersToSave {
					fmt.Fprintf(w, "LW %s, %d(%s)\n",
						register, (len(variablesOnStack) + i), stackPointer)
				}
			}

			if destination != "" {
				if isRegister(destination) {
					fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, returnRegister, zeroRegister)
				} else {
					fmt.Fprintf(w, "SW %s, %d(%s)\n", returnRegister, findPosition(destination), stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Tuple:
			if destination != "" {
				for i, element := range n.Elements {
					registers := loadVariables([]string{element}, variablesOnStack)
					fmt.Fprintf(w, "SW %s, %d(%s)\n", registers[0], i, heapPointer)
				}

				if isRegister(destination) {
					fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, heapPointer, zeroRegister)
				} else {
					fmt.Fprintf(w, "SW %s, %d(%s)\n", heapPointer, findPosition(destination), stackPointer)
				}

				fmt.Fprintf(w, "ADDI %s, %s, %d\n", heapPointer, heapPointer, len(n.Elements))
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.TupleGet:
			if destination != "" {
				registers := loadVariables([]string{n.Tuple}, variablesOnStack)

				if destination != "" {
					if isRegister(destination) {
						fmt.Fprintf(w, "LW %s, %d(%s)\n", destination, n.Index, registers[0])
					} else {
						fmt.Fprintf(w, "LW %s, %d(%s)\n", temporaryRegisters[0], n.Index, registers[0])
						fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", heapPointer, findPosition(destination), stackPointer)
				}

				loopLabel := getLabel()

				fmt.Fprintf(w, "%s:\n", loopLabel)

				fmt.Fprintf(w, "BEQ %s, %s, 4\n",
					temporaryRegisters[0], zeroRegister)

				fmt.Fprintf(w, "SW %s, 0(%s)\n",
					temporaryRegisters[1], heapPointer)

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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", heapPointer, findPosition(destination), stackPointer)
				}

				for i := 0; i < int(n.Length); i++ {
					fmt.Fprintf(w, "SW %s, %d(%s)\n",
						temporaryRegisters[0], i, heapPointer)
				}

				fmt.Fprintf(w, "ADDI %s, %s, %d\n", heapPointer, heapPointer, n.Length)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayGet:
			if destination != "" {
				registers := loadVariables([]string{n.Array, n.Index}, variablesOnStack)

				fmt.Fprintf(w, "ADD %s, %s, %s\n", temporaryRegisters[0], registers[0], registers[1])

				if isRegister(destination) {
					fmt.Fprintf(w, "LW %s, 0(%s)\n", destination, temporaryRegisters[0])
				} else {
					fmt.Fprintf(w, "LW %s, 0(%s)\n", temporaryRegisters[1], temporaryRegisters[0])
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[1], findPosition(destination), stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayGetImmediate:
			if destination != "" {
				registers := loadVariables([]string{n.Array}, variablesOnStack)

				if isRegister(destination) {
					fmt.Fprintf(w, "LW %s, %d(%s)\n", destination, n.Index, registers[0])
				} else {
					fmt.Fprintf(w, "LW %s, %d(%s)\n", temporaryRegisters[0], n.Index, registers[0])
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayPut:
			registers := loadVariables([]string{n.Array, n.Index, n.Value}, variablesOnStack)

			fmt.Fprintf(w, "ADD %s, %s, %s\n",
				temporaryRegisters[0], registers[0], registers[1])

			fmt.Fprintf(w, "SW %s, 0(%s)\n",
				registers[2], temporaryRegisters[0])

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayPutImmediate:
			registers := loadVariables([]string{n.Array, n.Value}, variablesOnStack)

			fmt.Fprintf(w, "SW %s, %d(%s)\n",
				registers[1], n.Index, registers[0])

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ReadInt:
			if destination == "" {
				fmt.Fprintf(w, "IN %s\n", temporaryRegisters[0])
			} else if isRegister(destination) {
				fmt.Fprintf(w, "IN %s\n", destination)
			} else {
				fmt.Fprintf(w, "IN %s\n", temporaryRegisters[0])
				fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
				fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
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
					fmt.Fprintf(w, "SW %s, %d(%s)\n", temporaryRegisters[0], findPosition(destination), stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		default:
			log.Panic("invalid node")
		}
	}

	// 210000
	fmt.Fprintf(w, "LUI %s, %s, 3\n", stackPointer, zeroRegister)
	fmt.Fprintf(w, "ORI %s, %s, 13392\n", stackPointer, stackPointer)

	// 240000
	fmt.Fprintf(w, "LUI %s, %s, 3\n", heapPointer, zeroRegister)
	fmt.Fprintf(w, "ORI %s, %s, 43392\n", heapPointer, heapPointer)

	fmt.Fprintf(w, "JAL main\n")
	fmt.Fprintf(w, "EXIT\n")

	for _, function := range append(functions, &ir.Function{
		Name: "main",
		Args: nil,
		Body: main,
	}) {
		fmt.Fprintf(w, "%s:\n", function.Name)
		if function.Name == "main" {
			emit(returnRegister, true, function.Body, functionToSpills[function.Name], stringset.New())
		} else {
			registersInUse := stringset.New()
			for _, arg := range function.Args {
				if isRegister(arg) {
					registersInUse.Add(arg)
				}
			}
			emit(returnRegister, true, function.Body, functionToSpills[function.Name], registersInUse)
		}
	}
}
