package typing

import (
	"fmt"
	"log"

	"github.com/kkty/mincaml-go/mir"
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

type constraint [2]Type

// GetTypes creates a mapping from names to types.
// The program should be alpha transformed before this operation.
func GetTypes(root mir.Node) map[string]Type {
	nameToType := map[string]Type{}

	constraints := []constraint{}

	nextTypeVarId := 0
	newTypeVar := func() TypeVar {
		defer func() { nextTypeVarId++ }()
		return TypeVar(fmt.Sprintf("_typing_%d", nextTypeVarId))
	}

	// Gets the type of a node, while gathering constraints.
	var getType func(node mir.Node) Type
	getType = func(node mir.Node) Type {
		switch node.(type) {
		case mir.Variable:
			return nameToType[node.(mir.Variable).Name]
		case mir.Unit:
			return UnitType
		case mir.Int:
			return IntType
		case mir.Bool:
			return BoolType
		case mir.Float:
			return FloatType
		case mir.Add:
			n := node.(mir.Add)
			constraints = append(constraints, constraint{nameToType[n.Left], IntType})
			constraints = append(constraints, constraint{nameToType[n.Right], IntType})
			return IntType
		case mir.Sub:
			n := node.(mir.Sub)
			constraints = append(constraints, constraint{nameToType[n.Left], IntType})
			constraints = append(constraints, constraint{nameToType[n.Right], IntType})
			return IntType
		case mir.FloatAdd:
			n := node.(mir.FloatAdd)
			constraints = append(constraints, constraint{nameToType[n.Left], FloatType})
			constraints = append(constraints, constraint{nameToType[n.Right], FloatType})
			return FloatType
		case mir.FloatSub:
			n := node.(mir.FloatSub)
			constraints = append(constraints, constraint{nameToType[n.Left], FloatType})
			constraints = append(constraints, constraint{nameToType[n.Right], FloatType})
			return FloatType
		case mir.FloatDiv:
			n := node.(mir.FloatDiv)
			constraints = append(constraints, constraint{nameToType[n.Left], FloatType})
			constraints = append(constraints, constraint{nameToType[n.Right], FloatType})
			return FloatType
		case mir.FloatMul:
			n := node.(mir.FloatMul)
			constraints = append(constraints, constraint{nameToType[n.Left], FloatType})
			constraints = append(constraints, constraint{nameToType[n.Right], FloatType})
			return FloatType
		case mir.IfEqual:
			n := node.(mir.IfEqual)
			constraints = append(constraints, constraint{nameToType[n.Left], nameToType[n.Right]})
			t1 := getType(n.True)
			t2 := getType(n.False)
			constraints = append(constraints, constraint{t1, t2})
			return t2
		case mir.IfLessThan:
			n := node.(mir.IfLessThan)
			constraints = append(constraints, constraint{nameToType[n.Left], nameToType[n.Right]})
			t1 := getType(n.True)
			t2 := getType(n.False)
			constraints = append(constraints, constraint{t1, t2})
			return t2
		case mir.ValueBinding:
			n := node.(mir.ValueBinding)
			t := getType(n.Value)
			nameToType[n.Name] = t
			t = getType(n.Next)
			return t
		case mir.FunctionBinding:
			n := node.(mir.FunctionBinding)

			argTypes := []Type{}
			for _, arg := range n.Args {
				t := newTypeVar()
				argTypes = append(argTypes, t)
				nameToType[arg] = t
			}

			returnType := newTypeVar()

			nameToType[n.Name] = FunctionType{argTypes, returnType}

			constraints = append(constraints, constraint{getType(n.Body), returnType})

			return getType(n.Next)
		case mir.Application:
			n := node.(mir.Application)
			argTypes := []Type{}
			for _, arg := range n.Args {
				argTypes = append(argTypes, nameToType[arg])
			}
			t := newTypeVar()
			constraints = append(constraints, constraint{nameToType[n.Function], FunctionType{argTypes, t}})
			return t
		case mir.Tuple:
			n := node.(mir.Tuple)

			elements := []Type{}
			for _, element := range n.Elements {
				elements = append(elements, nameToType[element])
			}

			return TupleType{elements}
		case mir.TupleBinding:
			n := node.(mir.TupleBinding)

			ts := []Type{}
			for _, name := range n.Names {
				t := newTypeVar()
				ts = append(ts, t)
				nameToType[name] = t
			}

			constraints = append(constraints, constraint{nameToType[n.Tuple], TupleType{ts}})

			return getType(n.Next)
		case mir.ArrayCreate:
			n := node.(mir.ArrayCreate)
			constraints = append(constraints, constraint{nameToType[n.Size], IntType})
			return ArrayType{nameToType[n.Value]}
		case mir.ArrayGet:
			n := node.(mir.ArrayGet)
			constraints = append(constraints, constraint{nameToType[n.Index], IntType})
			t := newTypeVar()
			constraints = append(constraints, constraint{nameToType[n.Array], ArrayType{t}})
			return t
		case mir.ArrayPut:
			n := node.(mir.ArrayPut)
			constraints = append(constraints, constraint{nameToType[n.Index], IntType})
			constraints = append(constraints, constraint{nameToType[n.Array], ArrayType{nameToType[n.Value]}})
			return UnitType
		case mir.ReadInt:
			return IntType
		case mir.ReadFloat:
			return FloatType
		case mir.PrintInt:
			n := node.(mir.PrintInt)
			constraints = append(constraints, constraint{IntType, nameToType[n.Arg]})
			return UnitType
		case mir.PrintChar:
			n := node.(mir.PrintChar)
			constraints = append(constraints, constraint{IntType, nameToType[n.Arg]})
			return UnitType
		case mir.IntToFloat:
			n := node.(mir.IntToFloat)
			constraints = append(constraints, constraint{IntType, nameToType[n.Arg]})
			return FloatType
		case mir.FloatToInt:
			n := node.(mir.FloatToInt)
			constraints = append(constraints, constraint{FloatType, nameToType[n.Arg]})
			return IntType
		case mir.Sqrt:
			n := node.(mir.Sqrt)
			constraints = append(constraints, constraint{FloatType, nameToType[n.Arg]})
			return FloatType
		case mir.Neg:
			n := node.(mir.Neg)
			return nameToType[n.Arg]
		}

		log.Fatal("invalid mir node")
		return nil
	}

	rootType := getType(root)

	mapping := unify(constraints)

	for k := range nameToType {
		nameToType[k] = nameToType[k].Replace(mapping, true)
	}

	rootType = rootType.Replace(mapping, true)

	if rootType != UnitType {
		log.Fatal("the program should be of unit type")
	}

	return nameToType
}

// Solves constraints and returns a mapping from TypeVar to Type.
func unify(constraints []constraint) map[TypeVar]Type {
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

			constraints = append(constraints, constraint{left.Return, right.Return})

			if len(left.Args) != len(right.Args) {
				log.Fatal("wrong number of arguments")
			}

			for i := 0; i < len(left.Args); i++ {
				constraints = append(constraints, constraint{left.Args[i], right.Args[i]})
			}

			continue
		}

		if left, ok := c[0].(TupleType); ok {
			right := c[1].(TupleType)

			if len(left.Elements) != len(right.Elements) {
				log.Fatal("wrong number of elements")
			}

			for i := 0; i < len(left.Elements); i++ {
				constraints = append(constraints, constraint{left.Elements[i], right.Elements[i]})
			}

			continue
		}

		if left, ok := c[0].(ArrayType); ok {
			right := c[1].(ArrayType)
			constraints = append(constraints, constraint{left.Inner, right.Inner})
			continue
		}

		if c[0] == c[1] {
			continue
		}

		log.Fatal("type mismatch")

	}

	return mapping
}
