package typing

import (
	"fmt"
	"log"
)

type Type interface {
	// Replace replaces TypeVars with Types.
	Replace(mapping map[string]Type, recursive bool) Type
}

type UnitType struct{}
type IntType struct{}
type FloatType struct{}
type BoolType struct{}

// TypeVar is for values of unknown type.
type TypeVar struct { Name string }

// TupleType is for tuples.
type TupleType struct{ Elements []Type }

// ArrayType is for arrays.
type ArrayType struct{ Inner Type }

// FunctionType is for functions.
type FunctionType struct {
	Args   []Type
	Return Type
}

func (t *UnitType) Replace(mapping map[string]Type, recursive bool) Type { return t }
func (t *IntType) Replace(mapping map[string]Type, recursive bool) Type { return t }
func (t *FloatType) Replace(mapping map[string]Type, recursive bool) Type { return t }
func (t *BoolType) Replace(mapping map[string]Type, recursive bool) Type { return t }

func (t *TupleType) Replace(mapping map[string]Type, recursive bool) Type {
	for i, element := range t.Elements {
		t.Elements[i] = element.Replace(mapping, recursive)
	}
	return t
}

func (t *ArrayType) Replace(mapping map[string]Type, recursive bool) Type {
	t.Inner = t.Inner.Replace(mapping, recursive)
	return t
}

func (t *FunctionType) Replace(mapping map[string]Type, recursive bool) Type {
	for i, arg := range t.Args {
		t.Args[i] = arg.Replace(mapping, recursive)
	}
	t.Return = t.Return.Replace(mapping, recursive)
	return t
}

func (t *TypeVar) Replace(mapping map[string]Type, recursive bool) Type {
	if v, ok := mapping[t.Name]; ok {
		if recursive {
			return v.Replace(mapping, recursive)
		}

		return v
	}

	return t
}

var nextTypeVarId int

func NewTypeVar() *TypeVar {
	defer func() { nextTypeVarId++ }()
	return &TypeVar{fmt.Sprintf("_typing_%d", nextTypeVarId)}
}

type Constraint [2]Type

// Unify solves constraints and returns a mapping from type variable names to Type.
func Unify(constraints []Constraint) map[string]Type {
	mapping := map[string]Type{}

	// Replaces TypeVar with another Type.
	updateConstraints := func(from string, to Type) {
		mapping := map[string]Type{from: to}
		for i := 0; i < len(constraints); i++ {
			constraints[i][0] = constraints[i][0].Replace(mapping, false)
			constraints[i][1] = constraints[i][1].Replace(mapping, false)
		}
	}

	for len(constraints) > 0 {
		c := constraints[0]
		constraints = constraints[1:]

		if left, ok := c[0].(*TypeVar); ok {
			right := c[1]
			mapping[left.Name] = right
			updateConstraints(left.Name, right)
			continue
		}

		if right, ok := c[1].(*TypeVar); ok {
			left := c[0]
			mapping[right.Name] = left
			updateConstraints(right.Name, left)
			continue
		}

		if left, ok := c[0].(*FunctionType); ok {
			right := c[1].(*FunctionType)

			constraints = append(constraints, Constraint{left.Return, right.Return})

			if len(left.Args) != len(right.Args) {
				log.Fatal("wrong number of arguments")
			}

			for i := 0; i < len(left.Args); i++ {
				constraints = append(constraints, Constraint{left.Args[i], right.Args[i]})
			}

			continue
		}

		if left, ok := c[0].(*TupleType); ok {
			right := c[1].(*TupleType)

			if len(left.Elements) != len(right.Elements) {
				log.Fatal("wrong number of elements")
			}

			for i := 0; i < len(left.Elements); i++ {
				constraints = append(constraints, Constraint{left.Elements[i], right.Elements[i]})
			}

			continue
		}

		if left, ok := c[0].(*ArrayType); ok {
			right := c[1].(*ArrayType)
			constraints = append(constraints, Constraint{left.Inner, right.Inner})
			continue
		}

		if c[0] == c[1] {
			continue
		}

		log.Fatal("type mismatch")

	}

	return mapping
}
