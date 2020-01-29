package emit

import (
	"fmt"
	"github.com/kkty/compiler/stringset"
	"io"
	"log"
	"math"
	"strings"

	"github.com/kkty/compiler/ir"
	"github.com/kkty/compiler/typing"
	"github.com/thoas/go-funk"
)

const (
	heapPointer          = "$hp"
	stackPointer         = "$sp"
	returnAddressPointer = "$ra"
	intZeroRegister      = "$izero"
	floatZeroRegister    = "$fzero"
	intReturnRegister    = "$iret"
	floatReturnRegister  = "$fret"
)

var (
	intTemporaryRegisters   = []string{"$itmp1", "$itmp2", "$itmp3"}
	floatTemporaryRegisters = []string{"$ftmp1", "$ftmp2"}
	intArgRegisters         = []string{"$iarg1", "$iarg2", "$iarg3"}
	floatArgRegisters       = []string{"$farg1", "$farg2"}
)

// Emit emits assembly code from IR.
func Emit(functions []*ir.Function, main ir.Node, types map[string]typing.Type, w io.Writer) {
	nextLabelId := 0
	getLabel := func() string {
		defer func() { nextLabelId++ }()
		return fmt.Sprintf("L%d", nextLabelId)
	}

	isIntRegister := func(s string) bool {
		return strings.HasPrefix(s, "$i")
	}

	isFloatRegister := func(s string) bool {
		return strings.HasPrefix(s, "$f")
	}

	isRegister := func(s string) bool {
		return isIntRegister(s) || isFloatRegister(s)
	}

	// load variables to registers if necessary
	// intArgRegisters/floatArgRegisters are used
	loadVariables := func(variables, storedVariables []string) []string {
		registers := []string{}
		nextIntArgRegister, nextFloatArgRegister := 0, 0
		for _, variable := range variables {
			if isRegister(variable) {
				registers = append(registers, variable)
			} else {
				idx := funk.IndexOf(storedVariables, variable)
				if idx == -1 {
					log.Panicf("variable not found on stack: %s", variable)
				}
				if _, ok := types[variable].(*typing.FloatType); ok {
					register := floatArgRegisters[nextFloatArgRegister]
					nextFloatArgRegister++
					fmt.Fprintf(w, "LWC1 %s, %d(%s)\n",
						register, idx*4, stackPointer)
					registers = append(registers, register)
				} else {
					register := intArgRegisters[nextIntArgRegister]
					nextIntArgRegister++
					fmt.Fprintf(w, "LW %s, %d(%s)\n",
						register, idx*4, stackPointer)
					registers = append(registers, register)
				}
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
			switch n.(type) {
			case *ir.IfEqual:
				n := n.(*ir.IfEqual)
				queue = append(queue, n.True, n.False)
			case *ir.IfEqualZero:
				n := n.(*ir.IfEqualZero)
				queue = append(queue, n.True, n.False)
			case *ir.IfEqualTrue:
				n := n.(*ir.IfEqualTrue)
				queue = append(queue, n.True, n.False)
			case *ir.IfLessThan:
				n := n.(*ir.IfLessThan)
				queue = append(queue, n.True, n.False)
			case *ir.IfLessThanZero:
				n := n.(*ir.IfLessThanZero)
				queue = append(queue, n.True, n.False)
			case *ir.Assignment:
				n := n.(*ir.Assignment)
				addVariables([]string{n.Name})
				queue = append(queue, n.Value, n.Next)
			case *ir.Application:
				n := n.(*ir.Application)
				functionToDependencies[function.Name].Add(n.Function)
			}
		}
	}

	// update functionToRegisters so that function applications are considered
	for {
		updated := false
		for _, function := range functions {
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

	var emit func(string, bool, ir.Node, []string)
	emit = func(
		destination string,
		tail bool,
		node ir.Node,
		variablesOnStack []string,
	) {
		findPosition := func(variable string) int {
			for i, v := range variablesOnStack {
				if v == variable {
					return i
				}
			}
			panic("variable not found on stack")
		}

		switch node.(type) {
		case *ir.Variable:
			n := node.(*ir.Variable)

			if isIntRegister(n.Name) {
				if isRegister(destination) {
					fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, n.Name, intZeroRegister)
				} else {
					fmt.Fprintf(w, "SW %s, %d(%s)\n", n.Name, findPosition(destination)*4, stackPointer)
				}
			} else if isFloatRegister(n.Name) {
				if isRegister(destination) {
					fmt.Fprintf(w, "ADDS %s, %s, %s\n", destination, n.Name, floatZeroRegister)
				} else {
					fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", n.Name, findPosition(destination)*4, stackPointer)
				}
			} else {
				if isRegister(destination) {
					fmt.Fprintf(w, "LW %s, %d(%s)\n", destination, findPosition(n.Name)*4, stackPointer)
				} else {
					if _, ok := types[n.Name].(*typing.FloatType); ok {
						fmt.Fprintf(w, "LWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(n.Name)*4, stackPointer)
						fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
					} else {
						fmt.Fprintf(w, "LW %s, %d(%s)\n", intTemporaryRegisters[0], findPosition(n.Name)*4, stackPointer)
						fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
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
			n := node.(*ir.Int)
			if isRegister(destination) {
				fmt.Fprintf(w, "ADDI %s, %s, %d\n", destination, intZeroRegister, n.Value)
			} else {
				fmt.Fprintf(w, "ADDI %s, %s, %d\n", intTemporaryRegisters[0], intZeroRegister, n.Value)
				fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}
			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Bool:
			n := node.(*ir.Bool)
			if n.Value {
				emit(destination, tail, &ir.Int{Value: 1}, variablesOnStack)
			} else {
				emit(destination, tail, &ir.Int{Value: 0}, variablesOnStack)
			}
		case *ir.Float:
			n := node.(*ir.Float)

			u := math.Float32bits(n.Value)
			fmt.Fprintf(w, "ORI %s, %s, %d\n", intTemporaryRegisters[0], intZeroRegister, u%(1<<16))
			fmt.Fprintf(w, "LUI %s, %s, %d\n", intTemporaryRegisters[0], intTemporaryRegisters[0], u>>16)

			if isRegister(destination) {
				fmt.Fprintf(w, "SW %s, 0(%s)\n", intTemporaryRegisters[0], heapPointer)
				fmt.Fprintf(w, "LWC1 %s, 0(%s)\n", destination, heapPointer)
			} else {
				fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Add:
			n := node.(*ir.Add)
			registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
			if isRegister(destination) {
				fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, registers[0], registers[1])
			} else {
				fmt.Fprintf(w, "ADD %s, %s, %s\n", intTemporaryRegisters[0], registers[0], registers[1])
				fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}
			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.AddImmediate:
			n := node.(*ir.AddImmediate)
			registers := loadVariables([]string{n.Left}, variablesOnStack)
			if isRegister(destination) {
				fmt.Fprintf(w, "ADDI %s, %s, %d\n", destination, registers[0], n.Right)
			} else {
				fmt.Fprintf(w, "ADDI %s, %s, %d\n", intTemporaryRegisters[0], registers[0], n.Right)
				fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}
			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Sub:
			n := node.(*ir.Sub)
			registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
			if isRegister(destination) {
				fmt.Fprintf(w, "SUB %s, %s, %s\n", destination, registers[0], registers[1])
			} else {
				fmt.Fprintf(w, "SUB %s, %s, %s\n", intTemporaryRegisters[0], registers[0], registers[1])
				fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}
			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.SubFromZero:
			n := node.(*ir.SubFromZero)
			registers := loadVariables([]string{n.Inner}, variablesOnStack)
			if isRegister(destination) {
				fmt.Fprintf(w, "SUB %s, %s, %s\n", destination, intZeroRegister, registers[0])
			} else {
				fmt.Fprintf(w, "SUB %s, %s, %s\n", intTemporaryRegisters[0], intZeroRegister, registers[0])
				fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}
			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.FloatAdd:
			n := node.(*ir.FloatAdd)
			registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
			if isRegister(destination) {
				fmt.Fprintf(w, "ADDS %s, %s, %s\n", destination, registers[0], registers[1])
			} else {
				fmt.Fprintf(w, "ADDS %s, %s, %s\n", floatTemporaryRegisters[0], registers[0], registers[1])
				fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}
			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.FloatSub:
			n := node.(*ir.FloatSub)
			registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
			if isRegister(destination) {
				fmt.Fprintf(w, "SUBS %s, %s, %s\n", destination, registers[0], registers[1])
			} else {
				fmt.Fprintf(w, "SUBS %s, %s, %s\n", floatTemporaryRegisters[0], registers[0], registers[1])
				fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}
			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.FloatSubFromZero:
			n := node.(*ir.FloatSubFromZero)
			registers := loadVariables([]string{n.Inner}, variablesOnStack)
			if isRegister(destination) {
				fmt.Fprintf(w, "SUBS %s, %s, %s\n", destination, floatZeroRegister, registers[0])
			} else {
				fmt.Fprintf(w, "SUBS %s, %s, %s\n", floatTemporaryRegisters[0], floatZeroRegister, registers[0])
				fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}
			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.FloatDiv:
			n := node.(*ir.FloatDiv)
			registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
			if isRegister(destination) {
				fmt.Fprintf(w, "DIVS %s, %s, %s\n", destination, registers[0], registers[1])
			} else {
				fmt.Fprintf(w, "DIVS %s, %s, %s\n", floatTemporaryRegisters[0], registers[0], registers[1])
				fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}
			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.FloatMul:
			n := node.(*ir.FloatMul)
			registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
			if isRegister(destination) {
				fmt.Fprintf(w, "MULS %s, %s, %s\n", destination, registers[0], registers[1])
			} else {
				fmt.Fprintf(w, "MULS %s, %s, %s\n", floatTemporaryRegisters[0], registers[0], registers[1])
				fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}
			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Not:
			n := node.(*ir.Not)
			registers := loadVariables([]string{n.Inner}, variablesOnStack)
			fmt.Fprintf(w, "ADDI %s, %s, 1\n", intTemporaryRegisters[0], intZeroRegister)
			if isRegister(destination) {
				fmt.Fprintf(w, "SUB %s, %s, %s\n", destination, intTemporaryRegisters[0], registers[0])
			} else {
				fmt.Fprintf(w, "SUB %s, %s, %s\n", intTemporaryRegisters[1], intTemporaryRegisters[0], registers[0])
				fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[1], findPosition(destination)*4, stackPointer)
			}
			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Equal:
			n := node.(*ir.Equal)
			registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)

			if isRegister(destination) {
				if isIntRegister(registers[0]) {
					fmt.Fprintf(w, "ADDI %s, %s, 1\n", destination, intZeroRegister)
					fmt.Fprintf(w, "BEQ %s, %s, 1\n", registers[0], registers[1])
					fmt.Fprintf(w, "ADDI %s, %s, 0\n", destination, intZeroRegister)
				} else {
					fmt.Fprintf(w, "ADDI %s, %s, 1\n", destination, intZeroRegister)
					fmt.Fprintf(w, "SUBS %s, %s, %s\n", floatTemporaryRegisters[0], registers[0], registers[1])
					fmt.Fprintf(w, "BZS %s, 1\n", floatTemporaryRegisters[0])
					fmt.Fprintf(w, "ADDI %s, %s, 0\n", destination, intZeroRegister)
				}
			} else {
				if isIntRegister(registers[0]) {
					fmt.Fprintf(w, "ADDI %s, %s, 1\n", intTemporaryRegisters[0], intZeroRegister)
					fmt.Fprintf(w, "BEQ %s, %s, 1\n", registers[0], registers[1])
					fmt.Fprintf(w, "ADDI %s, %s, 0\n", intTemporaryRegisters[0], intZeroRegister)
				} else {
					fmt.Fprintf(w, "ADDI %s, %s, 1\n", intTemporaryRegisters[0], intZeroRegister)
					fmt.Fprintf(w, "SUBS %s, %s, %s\n", floatTemporaryRegisters[0], registers[0], registers[1])
					fmt.Fprintf(w, "BZS %s, 1\n", floatTemporaryRegisters[0])
				}
				fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.LessThan:
			n := node.(*ir.LessThan)
			registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)

			if isRegister(destination) {
				if isIntRegister(registers[0]) {
					fmt.Fprintf(w, "SLT %s, %s, %s\n", destination, registers[0], registers[1])
				} else {
					fmt.Fprintf(w, "ADDI %s, %s, 0\n", destination, intZeroRegister)
					fmt.Fprintf(w, "BLS %s, %s, 1\n", registers[0], registers[1])
					fmt.Fprintf(w, "ADDI %s, %s, -1\n", destination, destination)
					fmt.Fprintf(w, "ADDI %s, %s, 1\n", destination, destination)
				}
			} else {
				if isIntRegister(registers[0]) {
					fmt.Fprintf(w, "SLT %s, %s, %s\n", intTemporaryRegisters[0], registers[0], registers[1])
				} else {
					fmt.Fprintf(w, "ADDI %s, %s, 0\n", intTemporaryRegisters[0], intZeroRegister)
					fmt.Fprintf(w, "BLS %s, %s, 1\n", registers[0], registers[1])
					fmt.Fprintf(w, "ADDI %s, %s, -1\n", intTemporaryRegisters[0], intTemporaryRegisters[0])
					fmt.Fprintf(w, "ADDI %s, %s, 1\n", intTemporaryRegisters[0], intTemporaryRegisters[0])
				}
				fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.IfEqual:
			n := node.(*ir.IfEqual)
			elseLabel := getLabel()
			continueLabel := getLabel()
			registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)
			if isIntRegister(registers[0]) {
				fmt.Fprintf(w, "BEQ %s, %s, 1\n", registers[0], registers[1])
			} else {
				fmt.Fprintf(w, "SUBS %s, %s, %s\n",
					floatTemporaryRegisters[0], registers[0], registers[1])
				fmt.Fprintf(w, "BZS %s, 1\n", floatTemporaryRegisters[0])
			}
			fmt.Fprintf(w, "J %s\n", elseLabel)
			emit(destination, tail, n.True, variablesOnStack)
			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}
			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")
			emit(destination, tail, n.False, variablesOnStack)
			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.IfEqualZero:
			n := node.(*ir.IfEqualZero)

			elseLabel := getLabel()
			continueLabel := getLabel()
			registers := loadVariables([]string{n.Inner}, variablesOnStack)

			if isIntRegister(registers[0]) {
				fmt.Fprintf(w, "BEQ %s, %s, 1\n", registers[0], intZeroRegister)
			} else {
				fmt.Fprintf(w, "SUBS %s, %s, %s\n",
					floatTemporaryRegisters[0], registers[0], registers[1])
				fmt.Fprintf(w, "BZS %s, 1\n", floatTemporaryRegisters[0])
			}

			fmt.Fprintf(w, "J %s\n", elseLabel)
			emit(destination, tail, n.True, variablesOnStack)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")
			emit(destination, tail, n.False, variablesOnStack)

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.IfEqualTrue:
			n := node.(*ir.IfEqualTrue)

			elseLabel := getLabel()
			continueLabel := getLabel()
			registers := loadVariables([]string{n.Inner}, variablesOnStack)

			fmt.Fprintf(w, "ADDI %s, %s, 1\n", intTemporaryRegisters[0], intZeroRegister)
			fmt.Fprintf(w, "BEQ %s, %s, 1\n", registers[0], intTemporaryRegisters[0])

			fmt.Fprintf(w, "J %s\n", elseLabel)
			emit(destination, tail, n.True, variablesOnStack)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")
			emit(destination, tail, n.False, variablesOnStack)

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.IfLessThan:
			n := node.(*ir.IfLessThan)

			elseLabel := getLabel()
			continueLabel := getLabel()
			registers := loadVariables([]string{n.Left, n.Right}, variablesOnStack)

			if isIntRegister(registers[0]) {
				fmt.Fprintf(w, "BL %s, %s, 1\n", registers[0], registers[1])
			} else {
				fmt.Fprintf(w, "BLS %s, %s, 1\n", registers[0], registers[1])
			}

			fmt.Fprintf(w, "J %s\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.True, variablesOnStack)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.False, variablesOnStack)

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.IfLessThanZero:
			n := node.(*ir.IfLessThanZero)

			elseLabel := getLabel()
			continueLabel := getLabel()
			registers := loadVariables([]string{n.Inner}, variablesOnStack)

			if isIntRegister(registers[0]) {
				fmt.Fprintf(w, "BL %s, %s, 1\n", registers[0], intZeroRegister)
			} else {
				fmt.Fprintf(w, "BLS %s, %s, 1\n", registers[0], floatZeroRegister)
			}

			fmt.Fprintf(w, "J %s\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.True, variablesOnStack)

			if !tail {
				fmt.Fprintf(w, "J %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "NOP\n")

			emit(destination, tail, n.False, variablesOnStack)

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "NOP\n")
			}
		case *ir.Assignment:
			n := node.(*ir.Assignment)
			emit(n.Name, false, n.Value, variablesOnStack)
			emit(destination, tail, n.Next, variablesOnStack)
		case *ir.Application:
			n := node.(*ir.Application)
			f := findFunction(n.Function)

			var registersToSave []string
			if tail {
				for _, arg := range f.Args {
					if isRegister(arg) {
						registersToSave = append(registersToSave, arg)
					}
				}
			} else {
				registersToSave = functionToRegisters[n.Function].Slice()
			}

			for i, register := range registersToSave {
				if isIntRegister(register) {
					fmt.Fprintf(w, "SW %s, %d(%s)\n",
						register, (len(variablesOnStack)+i)*4, stackPointer)
				} else {
					fmt.Fprintf(w, "SWC1 %s, %d(%s)\n",
						register, (len(variablesOnStack)+i)*4, stackPointer)
				}
			}

			// pass arguments by stack
			for i, arg := range f.Args {
				if !isRegister(arg) {
					registers := loadVariables([]string{n.Args[i]}, variablesOnStack)
					idx := funk.IndexOfString(functionToSpills[f.Name], arg)
					if idx == -1 {
						panic("variable not found")
					}
					if isIntRegister(registers[0]) {
						fmt.Fprintf(w, "SW %s, %d(%s)\n",
							registers[0], (len(variablesOnStack)+len(registersToSave)+1+idx)*4, stackPointer)
					} else {
						fmt.Fprintf(w, "SWC1 %s, %d(%s)\n",
							registers[0], (len(variablesOnStack)+len(registersToSave)+1+idx)*4, stackPointer)
					}
				}
			}

			// pass arguments by registers
			for i, arg := range f.Args {
				if isRegister(arg) {
					if isRegister(n.Args[i]) {
						idx := funk.IndexOfString(registersToSave, n.Args[i])
						if idx == -1 {
							if isIntRegister(arg) {
								fmt.Fprintf(w, "ADD %s, %s, %s\n",
									arg, n.Args[i], intZeroRegister)
							} else {
								fmt.Fprintf(w, "ADDS %s, %s, %s\n",
									arg, n.Args[i], floatZeroRegister)
							}
						} else {
							if isIntRegister(arg) {
								fmt.Fprintf(w, "LW %s, %d(%s)\n", arg, (len(variablesOnStack)+idx)*4, stackPointer)
							} else if isFloatRegister(arg) {
								fmt.Fprintf(w, "LWC1 %s, %d(%s)\n", arg, (len(variablesOnStack)+idx)*4, stackPointer)
							}
						}
					} else {
						idx := funk.IndexOfString(variablesOnStack, n.Args[i])
						if idx == -1 {
							log.Panicf("variable not found: %s", n.Args[i])
						}
						if isIntRegister(arg) {
							fmt.Fprintf(w, "LW %s, %d(%s)\n", arg, idx*4, stackPointer)
						} else {
							fmt.Fprintf(w, "LWC1 %s, %d(%s)\n", arg, idx*4, stackPointer)
						}
					}
				}
			}

			fmt.Fprintf(w, "SW %s, %d(%s)\n",
				returnAddressPointer, (len(variablesOnStack)+len(registersToSave))*4, stackPointer)

			fmt.Fprintf(w, "ADDI %s, %s, %d\n",
				stackPointer, stackPointer, (len(variablesOnStack)+len(registersToSave)+1)*4)

			fmt.Fprintf(w, "JAL %s\n", n.Function)

			fmt.Fprintf(w, "ADDI %s, %s, %d\n",
				stackPointer, stackPointer, -(len(variablesOnStack)+len(registersToSave)+1)*4)

			fmt.Fprintf(w, "LW %s, %d(%s)\n",
				returnAddressPointer, (len(variablesOnStack)+len(registersToSave))*4, stackPointer)

			// restore registers
			if !tail {
				for i, register := range registersToSave {
					if isIntRegister(register) {
						fmt.Fprintf(w, "LW %s, %d(%s)\n",
							register, (len(variablesOnStack)+i)*4, stackPointer)
					} else {
						fmt.Fprintf(w, "LWC1 %s, %d(%s)\n",
							register, (len(variablesOnStack)+i)*4, stackPointer)
					}
				}
			}

			if destination != "" {
				if _, ok := types[f.Name].(*typing.FunctionType).Return.(*typing.FloatType); ok {
					if isRegister(destination) {
						fmt.Fprintf(w, "ADDS %s, %s, %s\n", destination, floatReturnRegister, floatZeroRegister)
					} else {
						fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatReturnRegister, findPosition(destination)*4, stackPointer)
					}
				} else {
					if isRegister(destination) {
						fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, intReturnRegister, intZeroRegister)
					} else {
						fmt.Fprintf(w, "SW %s, %d(%s)\n", intReturnRegister, findPosition(destination)*4, stackPointer)
					}
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Tuple:
			n := node.(*ir.Tuple)

			for i, element := range n.Elements {
				registers := loadVariables([]string{element}, variablesOnStack)
				if isIntRegister(registers[0]) {
					fmt.Fprintf(w, "SW %s, %d(%s)\n", registers[0], i*4, heapPointer)
				} else {
					fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", registers[0], i*4, heapPointer)
				}
			}

			if isRegister(destination) {
				fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, heapPointer, intZeroRegister)
			} else {
				fmt.Fprintf(w, "SW %s, %d(%s)\n", heapPointer, findPosition(destination)*4, stackPointer)
			}

			fmt.Fprintf(w, "ADDI %s, %s, %d\n", heapPointer, heapPointer, len(n.Elements)*4)

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.TupleGet:
			n := node.(*ir.TupleGet)

			registers := loadVariables([]string{n.Tuple}, variablesOnStack)

			if isIntRegister(destination) {
				fmt.Fprintf(w, "LW %s, %d(%s)\n", destination, n.Index*4, registers[0])
			} else if isFloatRegister(destination) {
				fmt.Fprintf(w, "LWC1 %s, %d(%s)\n", destination, n.Index*4, registers[0])
			} else if _, ok := types[destination].(*typing.FloatType); ok {
				fmt.Fprintf(w, "LWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], n.Index*4, registers[0])
				fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			} else {
				fmt.Fprintf(w, "LW %s, %d(%s)\n", intTemporaryRegisters[0], n.Index*4, registers[0])
				fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayCreate:
			n := node.(*ir.ArrayCreate)

			registers := loadVariables([]string{n.Length, n.Value}, variablesOnStack)

			fmt.Fprintf(w, "ADD %s, %s, %s\n",
				intTemporaryRegisters[0], registers[0], intZeroRegister)

			if isIntRegister(registers[1]) {
				fmt.Fprintf(w, "ADD %s, %s, %s\n",
					intTemporaryRegisters[1], registers[1], intZeroRegister)
			} else {
				fmt.Fprintf(w, "ADDS %s, %s, %s\n",
					floatTemporaryRegisters[0], registers[1], floatZeroRegister)
			}

			if isRegister(destination) {
				fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, heapPointer, intZeroRegister)
			} else {
				fmt.Fprintf(w, "SW %s, %d(%s)\n", heapPointer, findPosition(destination)*4, stackPointer)
			}

			loopLabel := getLabel()

			fmt.Fprintf(w, "%s:\n", loopLabel)

			fmt.Fprintf(w, "BEQ %s, %s, 4\n",
				intTemporaryRegisters[0], intZeroRegister)

			if isIntRegister(registers[1]) {
				fmt.Fprintf(w, "SW %s, 0(%s)\n",
					intTemporaryRegisters[1], heapPointer)
			} else {
				fmt.Fprintf(w, "SWC1 %s, 0(%s)\n",
					floatTemporaryRegisters[0], heapPointer)
			}

			fmt.Fprintf(w, "ADDI %s, %s, 4\n", heapPointer, heapPointer)

			fmt.Fprintf(w, "ADDI %s, %s, -1\n",
				intTemporaryRegisters[0], intTemporaryRegisters[0])

			fmt.Fprintf(w, "J %s\n", loopLabel)

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayCreateImmediate:
			n := node.(*ir.ArrayCreateImmediate)

			registers := loadVariables([]string{n.Value}, variablesOnStack)

			if isIntRegister(registers[0]) {
				fmt.Fprintf(w, "ADD %s, %s, %s\n",
					intTemporaryRegisters[0], registers[0], intZeroRegister)
			} else {
				fmt.Fprintf(w, "ADDS %s, %s, %s\n",
					floatTemporaryRegisters[0], registers[0], floatZeroRegister)
			}

			if isRegister(destination) {
				fmt.Fprintf(w, "ADD %s, %s, %s\n", destination, heapPointer, intZeroRegister)
			} else {
				fmt.Fprintf(w, "SW %s, %d(%s)\n", heapPointer, findPosition(destination)*4, stackPointer)
			}

			for i := 0; i < int(n.Length); i++ {
				if isIntRegister(registers[0]) {
					fmt.Fprintf(w, "SW %s, %d(%s)\n",
						intTemporaryRegisters[0], i*4, heapPointer)
				} else {
					fmt.Fprintf(w, "SWC1 %s, %d(%s)\n",
						floatTemporaryRegisters[0], i*4, heapPointer)
				}
			}

			fmt.Fprintf(w, "ADDI %s, %s, %d\n", heapPointer, heapPointer, n.Length*4)

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayGet:
			n := node.(*ir.ArrayGet)

			registers := loadVariables([]string{n.Array, n.Index}, variablesOnStack)

			fmt.Fprintf(w, "SLL %s, %s, %d\n", intTemporaryRegisters[0], registers[1], 2)
			fmt.Fprintf(w, "ADD %s, %s, %s\n", intTemporaryRegisters[0], intTemporaryRegisters[0], registers[0])

			if isRegister(destination) {
				if isIntRegister(destination) {
					fmt.Fprintf(w, "LW %s, 0(%s)\n", destination, intTemporaryRegisters[0])
				} else {
					fmt.Fprintf(w, "LWC1 %s, 0(%s)\n", destination, intTemporaryRegisters[0])
				}
			} else {
				if _, ok := types[destination].(*typing.FloatType); !ok {
					fmt.Fprintf(w, "LW %s, 0(%s)\n", intTemporaryRegisters[1], intTemporaryRegisters[0])
					fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[1], findPosition(destination)*4, stackPointer)
				} else {
					fmt.Fprintf(w, "LWC1 %s, 0(%s)\n", floatTemporaryRegisters[0], intTemporaryRegisters[0])
					fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayGetImmediate:
			n := node.(*ir.ArrayGetImmediate)

			registers := loadVariables([]string{n.Array}, variablesOnStack)

			if isRegister(destination) {
				if isIntRegister(destination) {
					fmt.Fprintf(w, "LW %s, %d(%s)\n", destination, n.Index*4, registers[0])
				} else {
					fmt.Fprintf(w, "LWC1 %s, %d(%s)\n", destination, n.Index*4, registers[0])
				}
			} else {
				if _, ok := types[destination].(*typing.FloatType); !ok {
					fmt.Fprintf(w, "LW %s, %d(%s)\n", intTemporaryRegisters[0], n.Index*4, registers[0])
					fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
				} else {
					fmt.Fprintf(w, "LWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], n.Index*4, registers[0])
					fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
				}
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayPut:
			n := node.(*ir.ArrayPut)

			registers := loadVariables([]string{n.Array, n.Index, n.Value}, variablesOnStack)

			fmt.Fprintf(w, "SLL %s, %s, %d\n",
				intTemporaryRegisters[0], registers[1], 2)
			fmt.Fprintf(w, "ADD %s, %s, %s\n",
				intTemporaryRegisters[0], intTemporaryRegisters[0], registers[0])

			if isIntRegister(registers[2]) {
				fmt.Fprintf(w, "SW %s, 0(%s)\n",
					registers[2], intTemporaryRegisters[0])
			} else {
				fmt.Fprintf(w, "SWC1 %s, 0(%s)\n",
					registers[2], intTemporaryRegisters[0])
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ArrayPutImmediate:
			n := node.(*ir.ArrayPutImmediate)

			registers := loadVariables([]string{n.Array, n.Value}, variablesOnStack)

			if isIntRegister(registers[1]) {
				fmt.Fprintf(w, "SW %s, %d(%s)\n",
					registers[1], n.Index*4, registers[0])
			} else {
				fmt.Fprintf(w, "SWC1 %s, %d(%s)\n",
					registers[1], n.Index*4, registers[0])
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ReadInt:
			if isRegister(destination) {
				fmt.Fprintf(w, "IN %s\n", destination)
			} else {
				fmt.Fprintf(w, "IN %s\n", intTemporaryRegisters[0])
				fmt.Fprintf(w, "SW %s, %d(%s)\n", intTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.ReadFloat:
			if isRegister(destination) {
				fmt.Fprintf(w, "INF %s\n", destination)
			} else {
				fmt.Fprintf(w, "INF %s\n", floatTemporaryRegisters[0])
				fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.WriteByte:
			n := node.(*ir.WriteByte)
			registers := loadVariables([]string{n.Arg}, variablesOnStack)

			fmt.Fprintf(w, "OUT %s\n", registers[0])

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.IntToFloat:
			n := node.(*ir.IntToFloat)
			registers := loadVariables([]string{n.Arg}, variablesOnStack)

			fmt.Fprintf(w, "SW %s, 0(%s)\n", registers[0], heapPointer)
			fmt.Fprintf(w, "LWC1 %s, 0(%s)\n", floatTemporaryRegisters[0], heapPointer)

			if isRegister(destination) {
				fmt.Fprintf(w, "ITOF %s, %s\n", destination, floatTemporaryRegisters[0])
			} else {
				fmt.Fprintf(w, "ITOF %s, %s\n", floatTemporaryRegisters[0], floatTemporaryRegisters[0])
				fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.FloatToInt:
			n := node.(*ir.FloatToInt)
			registers := loadVariables([]string{n.Arg}, variablesOnStack)

			fmt.Fprintf(w, "FTOI %s, %s\n", floatTemporaryRegisters[0], registers[0])

			if isRegister(destination) {
				fmt.Fprintf(w, "SWC1 %s, 0(%s)\n", floatTemporaryRegisters[0], heapPointer)
				fmt.Fprintf(w, "LW %s, 0(%s)\n", destination, heapPointer)
			} else {
				fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		case *ir.Sqrt:
			n := node.(*ir.Sqrt)
			registers := loadVariables([]string{n.Arg}, variablesOnStack)

			if isRegister(destination) {
				fmt.Fprintf(w, "SQRT %s, %s\n", destination, registers[0])
			} else {
				fmt.Fprintf(w, "SQRT %s, %s\n", floatTemporaryRegisters[0], registers[0])
				fmt.Fprintf(w, "SWC1 %s, %d(%s)\n", floatTemporaryRegisters[0], findPosition(destination)*4, stackPointer)
			}

			if tail {
				fmt.Fprintf(w, "JR %s\n", returnAddressPointer)
			}
		default:
			log.Panic("invalid node")
		}
	}

	// 900000
	fmt.Fprintf(w, "LUI %s, %s, 13\n", stackPointer, intZeroRegister)
	fmt.Fprintf(w, "ORI %s, %s, 48032\n", stackPointer, stackPointer)

	// 1000000
	fmt.Fprintf(w, "LUI %s, %s, 15\n", heapPointer, intZeroRegister)
	fmt.Fprintf(w, "ORI %s, %s, 16960\n", heapPointer, heapPointer)

	fmt.Fprintf(w, "JAL main\n")
	fmt.Fprintf(w, "EXIT\n")

	for _, function := range append(functions, &ir.Function{
		Name: "main",
		Args: nil,
		Body: main,
	}) {
		fmt.Fprintf(w, "%s:\n", function.Name)
		if function.Name == "main" {
			emit(intReturnRegister, true, function.Body, functionToSpills[function.Name])
		} else if _, ok := types[function.Name].(*typing.FunctionType).Return.(*typing.FloatType); ok {
			emit(floatReturnRegister, true, function.Body, functionToSpills[function.Name])
		} else {
			emit(intReturnRegister, true, function.Body, functionToSpills[function.Name])
		}
	}
}
