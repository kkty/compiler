package typing

import (
	"fmt"
	"log"
)

type Type interface {
	// Replace replaces TypeVars with Types.
	Replace(mapping map[TypeVar]Type, recursive bool) Type
}

type atomic int

const (
	// UnitType is for unit values.
	UnitType atomic = iota
	// IntType is for int values.
	IntType
	// FloatType is for float values.
	FloatType
	// BoolType is for boolean values.
	BoolType
)

// TupleType is for tuples.
type TupleType struct{ Elements []Type }

// ArrayType is for arrays.
type ArrayType struct{ Inner Type }

// FunctionType is for functions.
type FunctionType struct {
	Args   []Type
	Return Type
}

// TypeVar is for values of unknown type.
type TypeVar string

func (t atomic) Replace(mapping map[TypeVar]Type, recursive bool) Type { return t }

func (t TupleType) Replace(mapping map[TypeVar]Type, recursive bool) Type {
	elements := []Type{}
	for _, element := range t.Elements {
		elements = append(elements, element.Replace(mapping, recursive))
	}
	return TupleType{elements}
}

func (t ArrayType) Replace(mapping map[TypeVar]Type, recursive bool) Type {
	return ArrayType{t.Inner.Replace(mapping, recursive)}
}

func (t FunctionType) Replace(mapping map[TypeVar]Type, recursive bool) Type {
	args := []Type{}
	for _, arg := range t.Args {
		args = append(args, arg.Replace(mapping, recursive))
	}
	return FunctionType{args, t.Return.Replace(mapping, recursive)}
}

func (t TypeVar) Replace(mapping map[TypeVar]Type, recursive bool) Type {
	if v, ok := mapping[t]; ok {
		if recursive {
			return v.Replace(mapping, recursive)
		}

		return v
	}

	return t
}

var nextTypeVarId int

func NewTypeVar() TypeVar {
	defer func() { nextTypeVarId++ }()
	return TypeVar(fmt.Sprintf("_typing_%d", nextTypeVarId))
}

type Constraint [2]Type

// Solves constraints and returns a mapping from TypeVar to Type.
func Unify(constraints []Constraint) map[TypeVar]Type {
	mapping := map[TypeVar]Type{}

	// Replaces the specified TypeVar with another Type.
	updateConstraints := func(from TypeVar, to Type) {
		mapping := map[TypeVar]Type{from: to}
		for i := 0; i < len(constraints); i++ {
			constraints[i][0] = constraints[i][0].Replace(mapping, false)
			constraints[i][1] = constraints[i][1].Replace(mapping, false)
		}
	}

	for len(constraints) > 0 {
		c := constraints[0]
		constraints = constraints[1:]

		if left, ok := c[0].(TypeVar); ok {
			right := c[1]
			mapping[left] = right
			updateConstraints(left, right)
			continue
		}

		if right, ok := c[1].(TypeVar); ok {
			left := c[0]
			mapping[right] = left
			updateConstraints(right, left)
			continue
		}

		if left, ok := c[0].(FunctionType); ok {
			right := c[1].(FunctionType)

			constraints = append(constraints, Constraint{left.Return, right.Return})

			if len(left.Args) != len(right.Args) {
				log.Fatal("wrong number of arguments")
			}

			for i := 0; i < len(left.Args); i++ {
				constraints = append(constraints, Constraint{left.Args[i], right.Args[i]})
			}

			continue
		}

		if left, ok := c[0].(TupleType); ok {
			right := c[1].(TupleType)

			if len(left.Elements) != len(right.Elements) {
				log.Fatal("wrong number of elements")
			}

			for i := 0; i < len(left.Elements); i++ {
				constraints = append(constraints, Constraint{left.Elements[i], right.Elements[i]})
			}

			continue
		}

		if left, ok := c[0].(ArrayType); ok {
			right := c[1].(ArrayType)
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
