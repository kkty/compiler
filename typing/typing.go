package typing

import (
	"fmt"
	"log"

	"github.com/kkty/mincaml-go/mir"
)

type Type interface {
	Concrete(mapping map[TypeVar]Type) Type
	isType()
}

type atomic int

const (
	UnitType atomic = iota
	IntType
	FloatType
	BoolType
)

type TupleType struct{ Elements []Type }
type ArrayType struct{ Inner Type }

type FunctionType struct {
	Args   []Type
	Return Type
}

type TypeVar string

func (t atomic) Concrete(mapping map[TypeVar]Type) Type { return t }

func (t TupleType) Concrete(mapping map[TypeVar]Type) Type {
	elements := []Type{}
	for _, element := range t.Elements {
		elements = append(elements, element.Concrete(mapping))
	}
	return TupleType{elements}
}

func (t ArrayType) Concrete(mapping map[TypeVar]Type) Type {
	return ArrayType{t.Inner.Concrete(mapping)}
}

func (t FunctionType) Concrete(mapping map[TypeVar]Type) Type {
	args := []Type{}
	for _, arg := range t.Args {
		args = append(args, arg.Concrete(mapping))
	}
	return FunctionType{args, t.Return.Concrete(mapping)}
}

func (t TypeVar) Concrete(mapping map[TypeVar]Type) Type {
	return mapping[t].Concrete(mapping)
}

func (t atomic) isType()       {}
func (t TupleType) isType()    {}
func (t ArrayType) isType()    {}
func (t FunctionType) isType() {}
func (t TypeVar) isType()      {}

var nextTypeVarId = 0

func newTypeVar() TypeVar {
	defer func() { nextTypeVarId++ }()
	return TypeVar(fmt.Sprintf("_t_%d", nextTypeVarId))
}

func GetTypes(node mir.Node) map[string]Type {
	typeEnv := map[string]Type{}
	_, constraints := getTypeAndConstraints(node, typeEnv)

	mapping := unify(constraints)

	for k := range typeEnv {
		typeEnv[k] = typeEnv[k].Concrete(mapping)
	}

	return typeEnv
}

type constraint [2]Type

func getTypeAndConstraints(node mir.Node, typeEnv map[string]Type) (Type, []constraint) {
	constraints := []constraint{}

	switch node.(type) {
	case mir.Variable:
		return typeEnv[node.(mir.Variable).Name], constraints
	case mir.Unit:
		return UnitType, constraints
	case mir.Int:
		return IntType, constraints
	case mir.Bool:
		return BoolType, constraints
	case mir.Float:
		return FloatType, constraints
	case mir.Add:
		n := node.(mir.Add)
		constraints = append(constraints, constraint{typeEnv[n.Left], IntType})
		constraints = append(constraints, constraint{typeEnv[n.Right], IntType})
		return IntType, constraints
	case mir.Sub:
		n := node.(mir.Sub)
		constraints = append(constraints, constraint{typeEnv[n.Left], IntType})
		constraints = append(constraints, constraint{typeEnv[n.Right], IntType})
		return IntType, constraints
	case mir.FloatAdd:
		n := node.(mir.FloatAdd)
		constraints = append(constraints, constraint{typeEnv[n.Left], FloatType})
		constraints = append(constraints, constraint{typeEnv[n.Right], FloatType})
		return FloatType, constraints
	case mir.FloatSub:
		n := node.(mir.FloatSub)
		constraints = append(constraints, constraint{typeEnv[n.Left], FloatType})
		constraints = append(constraints, constraint{typeEnv[n.Right], FloatType})
		return FloatType, constraints
	case mir.FloatDiv:
		n := node.(mir.FloatDiv)
		constraints = append(constraints, constraint{typeEnv[n.Left], FloatType})
		constraints = append(constraints, constraint{typeEnv[n.Right], FloatType})
		return FloatType, constraints
	case mir.FloatMul:
		n := node.(mir.FloatMul)
		constraints = append(constraints, constraint{typeEnv[n.Left], FloatType})
		constraints = append(constraints, constraint{typeEnv[n.Right], FloatType})
		return FloatType, constraints
	case mir.IfEqual:
		n := node.(mir.IfEqual)
		constraints = append(constraints, constraint{typeEnv[n.Left], typeEnv[n.Right]})
		t1, c := getTypeAndConstraints(n.True, typeEnv)
		constraints = append(constraints, c...)
		t2, c := getTypeAndConstraints(n.False, typeEnv)
		constraints = append(constraints, c...)
		constraints = append(constraints, constraint{t1, t2})
		return t2, constraints
	case mir.IfLessThanOrEqual:
		n := node.(mir.IfLessThanOrEqual)
		constraints = append(constraints, constraint{typeEnv[n.Left], typeEnv[n.Right]})
		t1, c := getTypeAndConstraints(n.True, typeEnv)
		constraints = append(constraints, c...)
		t2, c := getTypeAndConstraints(n.False, typeEnv)
		constraints = append(constraints, c...)
		constraints = append(constraints, constraint{t1, t2})
		return t2, constraints
	case mir.ValueBinding:
		n := node.(mir.ValueBinding)
		t, c := getTypeAndConstraints(n.Value, typeEnv)
		constraints = append(constraints, c...)
		typeEnv[n.Name] = t
		t, c = getTypeAndConstraints(n.Next, typeEnv)
		constraints = append(constraints, c...)
		return t, constraints
	case mir.FunctionBinding:
		n := node.(mir.FunctionBinding)

		argTypes := []Type{}
		for _ = range n.Args {
			argTypes = append(argTypes, newTypeVar())
		}

		returnType := newTypeVar()

		typeEnv[n.Name] = FunctionType{argTypes, returnType}
		for i, arg := range n.Args {
			typeEnv[arg] = argTypes[i]
		}
		t, c := getTypeAndConstraints(n.Body, typeEnv)
		constraints = append(constraints, c...)
		constraints = append(constraints, constraint{t, returnType})

		typeEnv[n.Name] = FunctionType{argTypes, returnType}
		t, c = getTypeAndConstraints(n.Next, typeEnv)
		constraints = append(constraints, c...)

		return t, constraints
	case mir.Application:
		n := node.(mir.Application)
		functionType := typeEnv[n.Function]
		argTypes := []Type{}
		for _, arg := range n.Args {
			argTypes = append(argTypes, typeEnv[arg])
		}
		t := newTypeVar()
		constraints = append(constraints, constraint{functionType, FunctionType{argTypes, t}})
		return t, constraints
	case mir.Tuple:
		n := node.(mir.Tuple)

		elements := []Type{}
		for _, element := range n.Elements {
			elements = append(elements, typeEnv[element])
		}

		return TupleType{elements}, constraints
	case mir.TupleBinding:
		n := node.(mir.TupleBinding)

		ts := []Type{}
		for _ = range n.Names {
			ts = append(ts, newTypeVar())
		}

		constraints = append(constraints, constraint{typeEnv[n.Tuple], TupleType{ts}})

		for i, name := range n.Names {
			typeEnv[name] = ts[i]
		}

		t, c := getTypeAndConstraints(n.Next, typeEnv)
		constraints = append(constraints, c...)

		return t, constraints
	case mir.ArrayCreate:
		n := node.(mir.ArrayCreate)
		constraints = append(constraints, constraint{typeEnv[n.Size], IntType})
		return ArrayType{typeEnv[n.Value]}, constraints
	case mir.ArrayGet:
		n := node.(mir.ArrayGet)
		constraints = append(constraints, constraint{typeEnv[n.Index], IntType})
		t := newTypeVar()
		constraints = append(constraints, constraint{typeEnv[n.Array], ArrayType{t}})
		return t, constraints
	case mir.ArrayPut:
		n := node.(mir.ArrayPut)
		constraints = append(constraints, constraint{typeEnv[n.Index], IntType})
		constraints = append(constraints, constraint{typeEnv[n.Array], ArrayType{typeEnv[n.Value]}})
		return UnitType, constraints
	}

	return nil, nil
}

func unify(constraints []constraint) map[TypeVar]Type {
	mapping := map[TypeVar]Type{}

	updateConstraints := func(from TypeVar, to Type) {
		for i := 0; i < len(constraints); i++ {
			if _, ok := constraints[i][0].(TypeVar); ok && constraints[i][0] == from {
				constraints[i][0] = to
			}

			if _, ok := constraints[i][1].(TypeVar); ok && constraints[i][1] == from {
				constraints[i][1] = to
			}
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
