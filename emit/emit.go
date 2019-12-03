package emit

import (
	"fmt"
	"io"
	"log"
	"math/rand"

	"github.com/kkty/compiler/ir"
	"github.com/kkty/compiler/typing"
	"github.com/thoas/go-funk"
)

const (
	argumentsToPassWithRegisters = 20
)

type Register interface {
	String() string
	register()
}

type IntRegister int
type FloatRegister int
type IntTemporaryRegister int
type FloatTemporaryRegister struct{}
type IntZeroRegister struct{}
type IntOneRegister struct{}
type FloatZeroRegister struct{}
type HeapPointer struct{}
type StackPointer struct{}

func (r IntRegister) register()            {}
func (r FloatRegister) register()          {}
func (r IntTemporaryRegister) register()   {}
func (r FloatTemporaryRegister) register() {}
func (r IntZeroRegister) register()        {}
func (r IntOneRegister) register()         {}
func (r FloatZeroRegister) register()      {}
func (r HeapPointer) regiser()             {}
func (r StackPointer) register()           {}

func (r IntRegister) String() string {
	return fmt.Sprintf("$i%d", r)
}

func (r FloatRegister) String() string {
	return fmt.Sprintf("$f%d", r)
}

func (r IntTemporaryRegister) String() string {
	return fmt.Sprintf("$tmp%d", r)
}

func (r FloatTemporaryRegister) String() string {
	return "$ftmp"
}

func (r IntZeroRegister) String() string {
	return "$zero"
}

func (r IntOneRegister) String() string {
	return "$one"
}

func (r FloatZeroRegister) String() string {
	return "$fzero"
}

func (r HeapPointer) String() string {
	return "$hp"
}

func (r StackPointer) String() string {
	return "$sp"
}

var intRegisters []IntRegister
var floatRegisters []FloatRegister
var stackPointer = StackPointer{}
var heapPointer = HeapPointer{}
var intZeroRegister = IntZeroRegister{}
var intOneRegister = IntOneRegister{}
var floatZeroRegister = FloatZeroRegister{}
var intTemporaryRegisters []IntTemporaryRegister
var floatTemporaryRegister = FloatTemporaryRegister{}

func init() {
	for i := 0; i < 2; i++ {
		intTemporaryRegisters = append(intTemporaryRegisters, IntTemporaryRegister(i))
	}

	for i := 0; i < 24; i++ {
		intRegisters = append(intRegisters, IntRegister(i))
	}

	for i := 0; i < 30; i++ {
		floatRegisters = append(floatRegisters, FloatRegister(i))
	}
}

type registerAndVariable struct {
	register Register
	variable string
}

type registerMapping []registerAndVariable

func newRegisterMapping() registerMapping {
	return registerMapping{}
}

// Finds the register allocated to the given variable.
// If no registers are allocated to the variable, the second
// return value will be false.
func (m registerMapping) findRegisterByVariable(name string) (Register, bool) {
	for _, registerAndVariable := range m {
		if registerAndVariable.variable == name {
			return registerAndVariable.register, true
		}
	}

	return IntRegister(0), false
}

// Finds the variable to which the given register is allocated to.
// If the register is not allocated to any variables, the second
// return value will be zero.
func (m registerMapping) findVariableByRegister(register Register) (string, bool) {
	for _, registerAndVariable := range m {
		if registerAndVariable.register == register {
			return registerAndVariable.variable, true
		}
	}

	return "", false
}

// Fetches one int register for use.
// If there is a variable that is allocated to registers and that is not in variablesToKeep,
// the matching register will be returned.
// If there is no such variables, the variable which appears last in "variablesToKeep" will be returned.
// If the second return value is true, the content of the register should be saved to the stack before use.
func (m registerMapping) getIntRegister(variablesToKeep map[string]struct{}) (IntRegister, bool) {
	// Finds a register which is not allocated to any registers.
	for i := len(intRegisters) - 1; i >= 0; i-- {
		intRegister := intRegisters[i]

		used := false

		for _, registerAndVariable := range m {
			if intRegister == registerAndVariable.register {
				used = true
			}
		}

		if !used {
			return intRegister, false
		}
	}

	for _, registerAndVariable := range m {
		if register, ok := registerAndVariable.register.(IntRegister); ok {
			if _, exists := variablesToKeep[registerAndVariable.variable]; !exists {
				return register, false
			}
		}
	}

	return intRegisters[rand.Intn(len(intRegisters))], true
}

// Fetches one float register for use.
// Similar to getIntRegister.
func (m registerMapping) getFloatRegister(variablesToKeep map[string]struct{}) (FloatRegister, bool) {
	for i := len(floatRegisters) - 1; i >= 0; i-- {
		floatRegister := floatRegisters[i]

		used := false

		for _, registerAndVariable := range m {
			if floatRegister == registerAndVariable.register {
				used = true
			}
		}

		if !used {
			return floatRegister, false
		}
	}

	for _, registerAndVariable := range m {
		if register, ok := registerAndVariable.register.(FloatRegister); ok {
			if _, exists := variablesToKeep[registerAndVariable.variable]; !exists {
				return register, false
			}
		}
	}

	return floatRegisters[rand.Intn(len(floatRegisters))], true
}

// Allocates a register to a variable.
// This operation is not in-place, as in "append" functions.
func (m registerMapping) add(variable string, register Register) registerMapping {
	for i, registerAndVariable := range m {
		if registerAndVariable.register == register {
			m := m.clone()
			m[i].variable = variable
			return m
		}
	}

	return append(m, registerAndVariable{register, variable})
}

func (m registerMapping) union(mm registerMapping) registerMapping {
	ret := newRegisterMapping()

	for _, registerAndVariable := range m {
		if funk.Contains(mm, registerAndVariable) {
			ret = append(ret, registerAndVariable)
		}
	}

	return ret
}

func (m registerMapping) remove(register Register) registerMapping {
	updated := newRegisterMapping()
	for _, registerAndVariable := range m {
		if registerAndVariable.register != register {
			updated = append(updated, registerAndVariable)
		}
	}
	return updated
}

// Clones a register mapping.
func (m registerMapping) clone() registerMapping {
	cloned := newRegisterMapping()
	for _, i := range m {
		cloned = append(cloned, i)
	}
	return cloned
}

func Emit(functions []*ir.Function, body ir.Node, types map[string]typing.Type, w io.Writer) {
	nextLabelId := 0
	getLabel := func() string {
		defer func() { nextLabelId++ }()
		return fmt.Sprintf("L%d", nextLabelId)
	}

	// Gathers float values in the program and prints the data section.
	floatValueToLabel := map[float32]string{}
	fmt.Fprintf(w, ".data\n")
	{
		floatValues := body.FloatValues()
		for _, function := range functions {
			floatValues = append(floatValues, function.Body.FloatValues()...)
		}

		floatValues = funk.UniqFloat32(floatValues)

		for _, floatValue := range floatValues {
			l := getLabel()
			floatValueToLabel[floatValue] = l
			fmt.Fprintf(w, "%s: .float %f\n", l, floatValue)
		}
	}

	fmt.Fprintf(w, ".text\n")

	// Loads variables to registers (if necessary) and returns the registers
	// allocated to the variables.
	// As the register mapping and the stack may change during the operation,
	// they are passed as return values.
	loadVariablesToRegisters := func(
		variables []string,
		registerMapping registerMapping,
		storedVariables []string,
		variablesToKeep map[string]struct{},
	) ([]Register, registerMapping, []string) {
		registers := []Register{}

		for _, variable := range variables {
			register, isInRegister := registerMapping.findRegisterByVariable(variable)

			if !isInRegister {
				// Loads the variable from the stack to a register.

				added := map[string]struct{}{}
				for _, v := range variables {
					if _, exists := variablesToKeep[v]; !exists {
						variablesToKeep[v] = struct{}{}
						added[v] = struct{}{}
					}
				}

				if types[variable] == typing.FloatType {
					var spill bool
					register, spill = registerMapping.getFloatRegister(variablesToKeep)

					if spill {
						v, _ := registerMapping.findVariableByRegister(register)

						if !funk.ContainsString(storedVariables, v) {
							fmt.Fprintf(w, "swc1 %s, %d(%s)\n",
								register, len(storedVariables)*4, stackPointer)
							storedVariables = append(storedVariables, v)
						}
					}

					idx := funk.IndexOfString(storedVariables, variable)

					if idx == -1 {
						log.Fatalf("variable not found on stack: %s", variable)
					}

					fmt.Fprintf(w, "lwc1 %s, %d(%s)\n",
						register, idx*4, stackPointer)
				} else {
					var spill bool
					register, spill = registerMapping.getIntRegister(variablesToKeep)

					if spill {
						// Stores the variable currently on the register to the stack.
						v, _ := registerMapping.findVariableByRegister(register)
						if !funk.ContainsString(storedVariables, v) {
							fmt.Fprintf(w, "sw %s, %d(%s)\n",
								register, len(storedVariables)*4, stackPointer)
							storedVariables = append(storedVariables, v)
						}
					}

					idx := funk.IndexOfString(storedVariables, variable)

					if idx == -1 {
						log.Fatalf("variable not found on stack: %s", variable)
					}

					fmt.Fprintf(w, "lw %s, %d(%s)\n",
						register.String(), idx*4, stackPointer)
				}

				for v := range added {
					delete(variablesToKeep, v)
				}

				registerMapping = registerMapping.add(variable, register)
			}

			registers = append(registers, register)
		}

		return registers, registerMapping, storedVariables
	}

	// Spills the variable on the given register if any.
	spillVariableOnRegister := func(
		register Register,
		registerMapping registerMapping,
		storedVariables []string,
		variablesToKeep map[string]struct{},
	) []string {
		v, exists := registerMapping.findVariableByRegister(register)

		if !exists {
			return storedVariables
		}

		// Does nothing if the variable is already on the stack.
		if funk.ContainsString(storedVariables, v) {
			return storedVariables
		}

		// Does nothing if the variable is not in "variablesToKeep".
		if _, exists := variablesToKeep[v]; !exists {
			return storedVariables
		}

		if types[v] == typing.FloatType {
			fmt.Fprintf(w, "swc1 %s, %d(%s)\n",
				register, len(storedVariables)*4, stackPointer)
		} else {
			fmt.Fprintf(w, "sw %s, %d(%s)\n",
				register, len(storedVariables)*4, stackPointer)
		}

		return append(storedVariables, v)
	}

	// Spills all the variables that will be used in the future
	// and that are allocated to registers.
	spillVariablesOnRegisters := func(
		registerMapping registerMapping,
		storedVariables []string,
		variablesToKeep map[string]struct{},
	) []string {
		for _, registerAndVariable := range registerMapping {
			register := registerAndVariable.register
			variable := registerAndVariable.variable

			if _, exists := variablesToKeep[variable]; exists {
				if !funk.ContainsString(storedVariables, variable) {
					if types[variable] == typing.FloatType {
						fmt.Fprintf(w, "swc1 %s, %d(%s)\n",
							register, len(storedVariables)*4, stackPointer)
					} else {
						fmt.Fprintf(w, "sw %s, %d(%s)\n",
							register, len(storedVariables)*4, stackPointer)
					}
					storedVariables = append(storedVariables, variable)
				}
			}
		}

		return storedVariables
	}

	var emit func(
		destination Register,
		tail bool,
		node ir.Node,
		registerMapping registerMapping,
		storedVariables []string,
		variablesToKeep map[string]struct{},
	) (registerMapping, []string)

	emit = func(
		destination Register,
		tail bool,
		node ir.Node,
		registerMapping registerMapping,
		storedVariables []string,
		variablesToKeep map[string]struct{},
	) (registerMapping, []string) {
		switch node.(type) {
		case *ir.Variable:
			n := node.(*ir.Variable)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Name},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			if types[n.Name] == typing.FloatType {
				fmt.Fprintf(w, "add.s %s, %s, %s\n",
					destination, registers[0], floatZeroRegister)
			} else {
				fmt.Fprintf(w, "add %s, %s, %s\n",
					destination, registers[0], intZeroRegister)
			}

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.Unit:
			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.Int:
			n := node.(*ir.Int)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "addi %s, %s, %d\n",
				destination, intZeroRegister, n.Value)

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.Bool:
			n := node.(*ir.Bool)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			if n.Value {
				fmt.Fprintf(w, "addi %s, %s, %d\n",
					destination, intZeroRegister.String(), 1)
			} else {
				fmt.Fprintf(w, "addi %s, %s, %d\n",
					destination, intZeroRegister.String(), 0)
			}

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.Float:
			n := node.(*ir.Float)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "lwc1 %s, %s(%s)\n",
				destination, floatValueToLabel[n.Value], intZeroRegister)

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.Add:
			n := node.(*ir.Add)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Left, n.Right},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "add %s, %s, %s\n", destination, registers[0], registers[1])

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.AddImmediate:
			n := node.(*ir.AddImmediate)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Left},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "addi %s, %s, %d\n", destination, registers[0], n.Right)

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.Sub:
			n := node.(*ir.Sub)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Left, n.Right},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "sub %s, %s, %s\n", destination, registers[0], registers[1])

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.SubFromZero:
			n := node.(*ir.SubFromZero)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Inner},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "sub %s, %s, %s\n", destination, intZeroRegister, registers[0])

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.FloatAdd:
			n := node.(*ir.FloatAdd)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Left, n.Right},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "add.s %s, %s, %s\n", destination, registers[0], registers[1])

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.FloatSub:
			n := node.(*ir.FloatSub)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Left, n.Right},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "sub.s %s, %s, %s\n", destination, registers[0], registers[1])

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.FloatSubFromZero:
			n := node.(*ir.FloatSubFromZero)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Inner},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "sub.s %s, %s, %s\n", destination, floatZeroRegister, registers[0])

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.FloatDiv:
			n := node.(*ir.FloatDiv)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Left, n.Right},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "div.s %s, %s, %s\n", destination, registers[0], registers[1])

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.FloatMul:
			n := node.(*ir.FloatMul)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Left, n.Right},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "mul.s %s, %s, %s\n", destination, registers[0], registers[1])

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.IfEqual:
			n := node.(*ir.IfEqual)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Left, n.Right},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariablesOnRegisters(
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			elseLabel := getLabel()
			continueLabel := getLabel()

			if types[n.Left] == typing.FloatType {
				fmt.Fprintf(w, "c.eq.s %s, %s\n", registers[0], registers[1])
				fmt.Fprintf(w, "bc1t 1\n")
			} else {
				fmt.Fprintf(w, "beq %s, %s, 1\n", registers[0], registers[1])
			}

			fmt.Fprintf(w, "j %s\n", elseLabel)

			// There is no need to keep variables alive inside because
			// all the registers that should be kept alive were stored to the stack.
			registerMapping1, storedVariables1 := emit(
				destination, tail, n.True, registerMapping, storedVariables, map[string]struct{}{})

			if !tail {
				fmt.Fprintf(w, "j %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "nop\n")
			registerMapping2, storedVariables2 := emit(
				destination, tail, n.False, registerMapping, storedVariables, map[string]struct{}{})

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "nop\n")
			}

			for i := len(storedVariables); i < len(storedVariables1) && i < len(storedVariables2); i++ {
				if storedVariables1[i] == storedVariables2[i] {
					storedVariables = append(storedVariables, storedVariables1[i])
				}
			}

			return registerMapping1.union(registerMapping2), storedVariables
		case *ir.IfEqualZero:
			n := node.(*ir.IfEqualZero)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Inner},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariablesOnRegisters(
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			elseLabel := getLabel()
			continueLabel := getLabel()

			if types[n.Inner] == typing.FloatType {
				fmt.Fprintf(w, "c.eq.s %s, %s\n", registers[0], floatZeroRegister)
				fmt.Fprintf(w, "bc1t 1\n")
			} else {
				fmt.Fprintf(w, "beq %s, %s, 1\n", registers[0], intZeroRegister)
			}

			fmt.Fprintf(w, "j %s\n", elseLabel)

			registerMapping1, storedVariables1 := emit(
				destination, tail, n.True, registerMapping, storedVariables, map[string]struct{}{})

			if !tail {
				fmt.Fprintf(w, "j %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "nop\n")
			registerMapping2, storedVariables2 := emit(
				destination, tail, n.False, registerMapping, storedVariables, map[string]struct{}{})

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "nop\n")
			}

			for i := len(storedVariables); i < len(storedVariables1) && i < len(storedVariables2); i++ {
				if storedVariables1[i] == storedVariables2[i] {
					storedVariables = append(storedVariables, storedVariables1[i])
				}
			}

			return registerMapping1.union(registerMapping2), storedVariables
		case *ir.IfLessThan:
			n := node.(*ir.IfLessThan)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Left, n.Right},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariablesOnRegisters(
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			elseLabel := getLabel()
			continueLabel := getLabel()

			if types[n.Left] == typing.FloatType {
				fmt.Fprintf(w, "c.eq.s %s, %s\n", registers[0], registers[1])
				fmt.Fprintf(w, "bc1t 2\n")
				fmt.Fprintf(w, "c.le.s %s, %s\n", registers[0], registers[1])
				fmt.Fprintf(w, "bc1t 1\n")
			} else {
				fmt.Fprintf(w, "slt %s, %s, %s\n", intTemporaryRegisters[0], registers[0], registers[1])
				fmt.Fprintf(w, "beq %s, %s, 1\n", intTemporaryRegisters[0], intOneRegister)
			}

			fmt.Fprintf(w, "j %s\n", elseLabel)
			fmt.Fprintf(w, "nop\n")
			registerMapping1, storedVariables1 := emit(
				destination, tail, n.True, registerMapping, storedVariables, map[string]struct{}{})

			if !tail {
				fmt.Fprintf(w, "j %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "nop\n")
			registerMapping2, storedVariables2 := emit(
				destination, tail, n.False, registerMapping, storedVariables, map[string]struct{}{})

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "nop\n")
			}

			for i := len(storedVariables); i < len(storedVariables1) && i < len(storedVariables2); i++ {
				if storedVariables1[i] == storedVariables2[i] {
					storedVariables = append(storedVariables, storedVariables1[i])
				}
			}

			return registerMapping1.union(registerMapping2), storedVariables
		case *ir.IfLessThanZero:
			n := node.(*ir.IfLessThanZero)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Inner},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariablesOnRegisters(
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			elseLabel := getLabel()
			continueLabel := getLabel()

			if types[n.Inner] == typing.FloatType {
				fmt.Fprintf(w, "c.eq.s %s, %s\n", registers[0], floatZeroRegister)
				fmt.Fprintf(w, "bc1t 2\n")
				fmt.Fprintf(w, "c.le.s %s, %s\n", registers[0], floatZeroRegister)
				fmt.Fprintf(w, "bc1t 1\n")
			} else {
				fmt.Fprintf(w, "slt %s, %s, %s\n",
					intTemporaryRegisters[0], registers[0], intZeroRegister)
				fmt.Fprintf(w, "beq %s, %s, 1\n", intTemporaryRegisters[0], intOneRegister)
			}

			fmt.Fprintf(w, "j %s\n", elseLabel)
			fmt.Fprintf(w, "nop\n")
			registerMapping1, storedVariables1 := emit(
				destination, tail, n.True, registerMapping, storedVariables, map[string]struct{}{})

			if !tail {
				fmt.Fprintf(w, "j %s\n", continueLabel)
			}

			fmt.Fprintf(w, "%s:\n", elseLabel)
			fmt.Fprintf(w, "nop\n")
			registerMapping2, storedVariables2 := emit(
				destination, tail, n.False, registerMapping, storedVariables, map[string]struct{}{})

			if !tail {
				fmt.Fprintf(w, "%s:\n", continueLabel)
				fmt.Fprintf(w, "nop\n")
			}

			for i := len(storedVariables); i < len(storedVariables1) && i < len(storedVariables2); i++ {
				if storedVariables1[i] == storedVariables2[i] {
					storedVariables = append(storedVariables, storedVariables1[i])
				}
			}

			return registerMapping1.union(registerMapping2), storedVariables
		case *ir.ValueBinding:
			n := node.(*ir.ValueBinding)

			register := func(name string, node ir.Node) Register {
				stack := []ir.Node{node}
				for len(stack) > 0 {
					node := stack[len(stack)-1]
					stack = stack[:len(stack)-1]

					switch node.(type) {
					case *ir.ValueBinding:
						n := node.(*ir.ValueBinding)
						stack = append(stack, n.Next, n.Value)
					case *ir.Application:
						n := node.(*ir.Application)
						nextIntRegister := IntRegister(0)
						nextFloatRegister := FloatRegister(0)
						for i, arg := range n.Args {
							if i >= argumentsToPassWithRegisters {
								if arg == name {
									return nil
								}

								continue
							}

							if types[arg] == typing.FloatType {
								if arg == name {
									return nextFloatRegister
								}

								nextFloatRegister++
							} else {
								if arg == name {
									return nextIntRegister
								}

								nextIntRegister++
							}
						}
					}
				}

				return nil
			}(n.Name, n.Next)

			register = nil

			added := map[string]struct{}{}

			for v := range n.Next.FreeVariables(map[string]struct{}{}) {
				if _, exists := variablesToKeep[v]; !exists {
					variablesToKeep[v] = struct{}{}
					added[v] = struct{}{}
				}
			}

			// Whether to spill the variable on the register (if any)
			// will be determined elsewhere (e.g. the case for ir.Add).

			if register == nil {
				if types[n.Name] == typing.FloatType {
					register, _ = registerMapping.getFloatRegister(variablesToKeep)
				} else {
					register, _ = registerMapping.getIntRegister(variablesToKeep)
				}
			}

			registerMapping, storedVariables = emit(
				register, false, n.Value, registerMapping, storedVariables, variablesToKeep)

			registerMapping = registerMapping.add(n.Name, register)

			for v := range added {
				delete(variablesToKeep, v)
			}

			return emit(destination, tail, n.Next, registerMapping, storedVariables, variablesToKeep)
		case *ir.Application:
			n := node.(*ir.Application)

			storedVariables = spillVariablesOnRegisters(
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			// Moves the values to their correct positions.
			// If an argument is allocated to a wrong register, it is be moved to
			// the correct one. If an argument is not allocated to a register,
			// it is loaded from the stack.

			nextIntRegister := IntRegister(0)
			nextFloatRegister := FloatRegister(0)
			argumentsToPassWithStack := []string{}

			for i, arg := range n.Args {
				if i >= argumentsToPassWithRegisters {
					argumentsToPassWithStack = append(argumentsToPassWithStack, arg)
					continue
				}

				variablesToKeep := map[string]struct{}{}
				for _, v := range n.Args[i+1:] {
					variablesToKeep[v] = struct{}{}
				}

				if types[arg] == typing.FloatType {
					correctRegister := nextFloatRegister
					nextFloatRegister++

					storedVariables = spillVariableOnRegister(
						correctRegister,
						registerMapping,
						storedVariables,
						variablesToKeep,
					)

					currentRegister, isInRegister := registerMapping.findRegisterByVariable(arg)

					if !isInRegister || correctRegister != currentRegister {
						if isInRegister {
							fmt.Fprintf(w, "add.s %s, %s, %s\n",
								correctRegister, currentRegister, floatZeroRegister)
						} else {
							idx := funk.IndexOfString(storedVariables, arg)

							if idx == -1 {
								log.Fatalf("variable not found on stack: %s", arg)
							}

							fmt.Fprintf(w, "lwc1 %s, %d(%s)\n",
								correctRegister, idx*4, stackPointer)
						}

						registerMapping = registerMapping.add(arg, correctRegister)
					}
				} else {
					correctRegister := nextIntRegister
					nextIntRegister++

					storedVariables = spillVariableOnRegister(
						correctRegister,
						registerMapping,
						storedVariables,
						variablesToKeep,
					)

					currentRegister, isInRegister := registerMapping.findRegisterByVariable(arg)

					if !isInRegister || correctRegister != currentRegister {
						if isInRegister {
							fmt.Fprintf(w, "add %s, %s, %s\n",
								correctRegister, currentRegister, intZeroRegister)
						} else {
							idx := funk.IndexOfString(storedVariables, arg)

							if idx == -1 {
								log.Fatalf("variable not found on stack: %s", arg)
							}

							fmt.Fprintf(w, "lw %s, %d(%s)\n",
								correctRegister, idx*4, stackPointer)
						}

						registerMapping = registerMapping.add(arg, correctRegister)
					}
				}
			}

			for i, arg := range argumentsToPassWithStack {
				register, isInRegister := registerMapping.findRegisterByVariable(arg)

				if isInRegister {
					if types[arg] == typing.FloatType {
						fmt.Fprintf(w, "add.s %s, %s, %s\n",
							floatTemporaryRegister, floatZeroRegister, register)
					} else {
						fmt.Fprintf(w, "add %s, %s, %s\n",
							intTemporaryRegisters[0], intZeroRegister, register)
					}
				} else {
					idx := funk.IndexOfString(storedVariables, arg)

					if idx == -1 {
						log.Fatalf("variable not found on stack: %s", arg)
					}

					if types[arg] == typing.FloatType {
						fmt.Fprintf(w, "lwc1 %s, %d(%s)\n",
							floatTemporaryRegister, idx*4, stackPointer)
					} else {
						fmt.Fprintf(w, "lw %s, %d(%s)\n",
							intTemporaryRegisters[0], idx*4, stackPointer)
					}
				}

				if types[arg] == typing.FloatType {
					fmt.Fprintf(w, "swc1 %s, %d(%s)\n",
						floatTemporaryRegister, (len(storedVariables)+1)*4+i*4, stackPointer)
				} else {
					fmt.Fprintf(w, "sw %s, %d(%s)\n",
						intTemporaryRegisters[0], (len(storedVariables)+1)*4+i*4, stackPointer)
				}
			}

			if tail {
				for i := range argumentsToPassWithStack {
					fmt.Fprintf(w, "lw %s, %d(%s)\n",
						intTemporaryRegisters[0], (len(storedVariables)+1)*4+i*4, stackPointer)
					fmt.Fprintf(w, "sw %s, %d(%s)\n",
						intTemporaryRegisters[0], i*4, stackPointer)
				}

				fmt.Fprintf(w, "j %s\n", n.Function)
			} else {
				fmt.Fprintf(w, "sw $ra, %d(%s)\n",
					len(storedVariables)*4, stackPointer)

				fmt.Fprintf(w, "addi %s, %s, %d\n",
					stackPointer, stackPointer, (len(storedVariables)+1)*4)

				fmt.Fprintf(w, "jal %s\n", n.Function)

				fmt.Fprintf(w, "addi %s, %s, %d\n",
					stackPointer, stackPointer, -(len(storedVariables)+1)*4)

				fmt.Fprintf(w, "lw $ra, %d(%s)\n",
					len(storedVariables)*4, stackPointer)

				if types[n.Function].(typing.FunctionType).Return == typing.FloatType {
					fmt.Fprintf(w, "add.s %s, %s, %s\n",
						destination, FloatRegister(0), floatZeroRegister)
				} else {
					fmt.Fprintf(w, "add %s, %s, %s\n",
						destination, IntRegister(0), intZeroRegister)
				}
			}

			return newRegisterMapping(), storedVariables
		case *ir.Tuple:
			n := node.(*ir.Tuple)

			for i, element := range n.Elements {
				var registers []Register

				added := map[string]struct{}{}
				for _, v := range n.Elements[i+1:] {
					if _, exists := variablesToKeep[v]; !exists {
						added[v] = struct{}{}
						variablesToKeep[v] = struct{}{}
					}
				}

				registers, registerMapping, storedVariables = loadVariablesToRegisters(
					[]string{element},
					registerMapping,
					storedVariables,
					variablesToKeep,
				)

				for v := range added {
					delete(variablesToKeep, v)
				}

				switch registers[0].(type) {
				case FloatRegister:
					fmt.Fprintf(w, "swc1 %s, %d(%s)\n", registers[0], i*4, heapPointer)
				default:
					fmt.Fprintf(w, "sw %s, %d(%s)\n", registers[0], i*4, heapPointer)
				}
			}

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "add %s, %s, %s\n", destination, heapPointer, intZeroRegister)
			fmt.Fprintf(w, "addi %s, %s, %d\n", heapPointer, heapPointer, len(n.Elements)*4)

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.TupleGet:
			n := node.(*ir.TupleGet)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Tuple},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			switch destination.(type) {
			case FloatRegister:
				fmt.Fprintf(w, "lwc1 %s, %d(%s)\n",
					destination, n.Index*4, registers[0])
			default:
				fmt.Fprintf(w, "lw %s, %d(%s)\n",
					destination, n.Index*4, registers[0])
			}

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.ArrayCreate:
			n := node.(*ir.ArrayCreate)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Length, n.Value},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "add %s, %s, %s\n",
				intTemporaryRegisters[0], registers[0], intZeroRegister)

			if types[n.Value] == typing.FloatType {
				fmt.Fprintf(w, "add.s %s, %s, %s\n",
					floatTemporaryRegister, registers[1], floatZeroRegister)
			} else {
				fmt.Fprintf(w, "add %s, %s, %s\n",
					intTemporaryRegisters[1], registers[1], intZeroRegister)
			}

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "add %s, %s, %s\n",
				destination, heapPointer, intZeroRegister)

			loopLabel := getLabel()

			fmt.Fprintf(w, "%s:\n", loopLabel)

			fmt.Fprintf(w, "beq %s, %s, 4\n",
				intTemporaryRegisters[0], intZeroRegister)

			switch registers[1].(type) {
			case FloatRegister:
				fmt.Fprintf(w, "swc1 %s, 0(%s)\n",
					floatTemporaryRegister, heapPointer)
			default:
				fmt.Fprintf(w, "sw %s, 0(%s)\n",
					intTemporaryRegisters[1], heapPointer)
			}

			fmt.Fprintf(w, "addi %s, %s, 4\n", heapPointer, heapPointer)

			fmt.Fprintf(w, "addi %s, %s -1\n",
				intTemporaryRegisters[0], intTemporaryRegisters[0])

			fmt.Fprintf(w, "j %s\n", loopLabel)

			fmt.Fprintf(w, "nop\n")

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.ArrayCreateImmediate:
			n := node.(*ir.ArrayCreateImmediate)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Value},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			if types[n.Value] == typing.FloatType {
				fmt.Fprintf(w, "add.s %s, %s, %s\n",
					floatTemporaryRegister, registers[0], floatZeroRegister)
			} else {
				fmt.Fprintf(w, "add %s, %s, %s\n",
					intTemporaryRegisters[0], registers[0], intZeroRegister)
			}

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "add %s, %s, %s\n",
				destination, heapPointer, intZeroRegister)

			for i := 0; i < int(n.Length); i++ {
				switch registers[0].(type) {
				case FloatRegister:
					fmt.Fprintf(w, "swc1 %s, %d(%s)\n",
						floatTemporaryRegister, i*4, heapPointer)
				default:
					fmt.Fprintf(w, "sw %s, %d(%s)\n",
						intTemporaryRegisters[0], i*4, heapPointer)
				}
			}

			fmt.Fprintf(w, "addi %s, %s, %d\n", heapPointer, heapPointer, n.Length*4)

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.ArrayGet:
			n := node.(*ir.ArrayGet)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Array, n.Index},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "sll %s, %s, %d\n",
				intTemporaryRegisters[0], registers[1], 2)
			fmt.Fprintf(w, "add %s, %s, %s\n",
				intTemporaryRegisters[0], intTemporaryRegisters[0], registers[0])

			switch destination.(type) {
			case FloatRegister:
				fmt.Fprintf(w, "lwc1 %s, 0(%s)\n",
					destination, intTemporaryRegisters[0])
			default:
				fmt.Fprintf(w, "lw %s, 0(%s)\n",
					destination, intTemporaryRegisters[0])
			}

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.ArrayGetImmediate:
			n := node.(*ir.ArrayGetImmediate)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Array},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			switch destination.(type) {
			case FloatRegister:
				fmt.Fprintf(w, "lwc1 %s, %d(%s)\n",
					destination, n.Index*4, registers[0])
			default:
				fmt.Fprintf(w, "lw %s, %d(%s)\n",
					destination, n.Index*4, registers[0])
			}

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.ArrayPut:
			n := node.(*ir.ArrayPut)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Array, n.Index, n.Value},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "sll %s, %s, %d\n",
				intTemporaryRegisters[0], registers[1], 2)
			fmt.Fprintf(w, "add %s, %s, %s\n",
				intTemporaryRegisters[0], intTemporaryRegisters[0], registers[0])

			switch registers[2].(type) {
			case FloatRegister:
				fmt.Fprintf(w, "swc1 %s, 0(%s)\n",
					registers[2], intTemporaryRegisters[0])
			default:
				fmt.Fprintf(w, "sw %s, 0(%s)\n",
					registers[2], intTemporaryRegisters[0])
			}

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.ArrayPutImmediate:
			n := node.(*ir.ArrayPutImmediate)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Array, n.Value},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			switch registers[1].(type) {
			case FloatRegister:
				fmt.Fprintf(w, "swc1 %s, %d(%s)\n",
					registers[1], n.Index*4, registers[0])
			default:
				fmt.Fprintf(w, "sw %s, %d(%s)\n",
					registers[1], n.Index*4, registers[0])
			}

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.ReadInt:
			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "read_i %s\n", destination.String())

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.ReadFloat:
			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "read_f %s\n", destination.String())

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.PrintInt:
			n := node.(*ir.PrintInt)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Arg},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "out_i %s\n", registers[0].String())

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.PrintChar:
			n := node.(*ir.PrintChar)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Arg},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "out_c %s\n", registers[0].String())

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.IntToFloat:
			n := node.(*ir.IntToFloat)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Arg},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "itof %s, %s\n", destination, registers[0])

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.FloatToInt:
			n := node.(*ir.FloatToInt)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Arg},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "ftoi %s, %s\n", destination, registers[0])

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		case *ir.Sqrt:
			n := node.(*ir.Sqrt)

			registers, registerMapping, storedVariables := loadVariablesToRegisters(
				[]string{n.Arg},
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			storedVariables = spillVariableOnRegister(
				destination,
				registerMapping,
				storedVariables,
				variablesToKeep,
			)

			fmt.Fprintf(w, "sqrt %s, %s\n", destination, registers[0])

			if tail {
				fmt.Fprintf(w, "jr $ra\n")
			}

			return registerMapping, storedVariables
		}

		log.Fatal("invalid ir node")

		return registerMapping, storedVariables
	}

	for _, function := range functions {
		fmt.Fprintf(w, "%s:\n", function.Name)

		storedVariables := []string{}
		registerMapping := newRegisterMapping()

		intRegister := IntRegister(0)
		floatRegister := FloatRegister(0)
		for i, arg := range function.Args {
			if i >= argumentsToPassWithRegisters {
				storedVariables = append(storedVariables, arg)
				continue
			}

			switch types[arg] {
			case typing.FloatType:
				registerMapping = registerMapping.add(arg, floatRegister)
				floatRegister++
			default:
				registerMapping = registerMapping.add(arg, intRegister)
				intRegister++
			}
		}

		var destination Register
		if types[function.Name].(typing.FunctionType).Return == typing.FloatType {
			destination = FloatRegister(0)
		} else {
			destination = IntRegister(0)
		}

		emit(destination, true, function.Body, registerMapping, storedVariables, map[string]struct{}{})
	}

	fmt.Fprintf(w, "start:\n")
	fmt.Fprintf(w, "addi $sp, $zero, 10000000\n")
	fmt.Fprintf(w, "addi $hp, $zero, 20000000\n")
	fmt.Fprintf(w, "addi %s, %s, %d\n", intOneRegister, intZeroRegister, 1)

	emit(
		IntRegister(0),
		false,
		body,
		newRegisterMapping(),
		[]string{},
		map[string]struct{}{},
	)
	fmt.Fprintf(w, "exit\n")
}
