package ast

import (
	"github.com/kkty/compiler/typing"
)

// GetTypes constructs the mapping from variable/function names to their types.
// This should be called after AlphaTransform().
func GetTypes(root Node) map[string]typing.Type {
	nameToType := map[string]typing.Type{}
	constraints := []typing.Constraint{}

	// get the type of a node while collecting constraints.
	var getType func(node Node) typing.Type
	getType = func(node Node) typing.Type {
		switch n := node.(type) {
		case *Variable:
			return nameToType[n.Name]
		case *Unit:
			return &typing.UnitType{}
		case *Int:
			return &typing.IntType{}
		case *Bool:
			return &typing.BoolType{}
		case *Float:
			return &typing.FloatType{}
		case *Add:
			constraints = append(constraints,
				typing.Constraint{getType(n.Left), &typing.IntType{}},
				typing.Constraint{getType(n.Right), &typing.IntType{}})
			return &typing.IntType{}
		case *Sub:
			constraints = append(constraints,
				typing.Constraint{getType(n.Left), &typing.IntType{}},
				typing.Constraint{getType(n.Right), &typing.IntType{}})
			return &typing.IntType{}
		case *FloatAdd:
			constraints = append(constraints,
				typing.Constraint{getType(n.Left), &typing.FloatType{}},
				typing.Constraint{getType(n.Right), &typing.FloatType{}})
			return &typing.FloatType{}
		case *FloatSub:
			constraints = append(constraints,
				typing.Constraint{getType(n.Left), &typing.FloatType{}},
				typing.Constraint{getType(n.Right), &typing.FloatType{}})
			return &typing.FloatType{}
		case *FloatDiv:
			constraints = append(constraints,
				typing.Constraint{getType(n.Left), &typing.FloatType{}},
				typing.Constraint{getType(n.Right), &typing.FloatType{}})
			return &typing.FloatType{}
		case *FloatMul:
			constraints = append(constraints,
				typing.Constraint{getType(n.Left), &typing.FloatType{}},
				typing.Constraint{getType(n.Right), &typing.FloatType{}})
			return &typing.FloatType{}
		case *Equal:
			constraints = append(constraints,
				typing.Constraint{getType(n.Left), getType(n.Right)})
			return &typing.BoolType{}
		case *LessThan:
			constraints = append(constraints,
				typing.Constraint{getType(n.Left), getType(n.Right)})
			return &typing.BoolType{}
		case *Neg:
			return getType(n.Inner)
		case *FloatNeg:
			constraints = append(constraints,
				typing.Constraint{getType(n.Inner), &typing.FloatType{}})
			return &typing.FloatType{}
		case *Not:
			constraints = append(constraints,
				typing.Constraint{getType(n.Inner), &typing.BoolType{}})
			return &typing.BoolType{}
		case *If:
			t1, t2 := getType(n.True), getType(n.False)
			constraints = append(constraints,
				typing.Constraint{getType(n.Condition), &typing.BoolType{}},
				typing.Constraint{t1, t2})
			return t1
		case *Assignment:
			nameToType[n.Name] = getType(n.Body)
			return getType(n.Next)
		case *FunctionAssignment:
			argTypes := []typing.Type{}
			for _, arg := range n.Args {
				t := typing.NewTypeVar()
				argTypes = append(argTypes, t)
				nameToType[arg] = t
			}
			returnType := typing.NewTypeVar()
			nameToType[n.Name] = &typing.FunctionType{argTypes, returnType}
			constraints = append(constraints,
				typing.Constraint{getType(n.Body), returnType})
			return getType(n.Next)
		case *Application:
			argTypes := []typing.Type{}
			for _, arg := range n.Args {
				argTypes = append(argTypes, getType(arg))
			}
			t := typing.NewTypeVar()
			constraints = append(constraints,
				typing.Constraint{nameToType[n.Function], &typing.FunctionType{argTypes, t}})
			return t
		case *Tuple:
			elements := []typing.Type{}
			for _, element := range n.Elements {
				elements = append(elements, getType(element))
			}
			return &typing.TupleType{elements}
		case *TupleAssignment:
			ts := []typing.Type{}
			for _, name := range n.Names {
				t := typing.NewTypeVar()
				ts = append(ts, t)
				nameToType[name] = t
			}
			constraints = append(constraints,
				typing.Constraint{getType(n.Tuple), &typing.TupleType{ts}})
			return getType(n.Next)
		case *ArrayCreate:
			constraints = append(constraints,
				typing.Constraint{getType(n.Size), &typing.IntType{}})
			return &typing.ArrayType{getType(n.Value)}
		case *ArrayGet:
			t := typing.NewTypeVar()
			constraints = append(constraints,
				typing.Constraint{getType(n.Index), &typing.IntType{}},
				typing.Constraint{getType(n.Array), &typing.ArrayType{t}})
			return t
		case *ArrayPut:
			constraints = append(constraints,
				typing.Constraint{getType(n.Index), &typing.IntType{}},
				typing.Constraint{getType(n.Array), &typing.ArrayType{getType(n.Value)}})
			return &typing.UnitType{}
		case *ReadInt:
			return &typing.IntType{}
		case *ReadFloat:
			return &typing.FloatType{}
		case *WriteByte:
			constraints = append(constraints,
				typing.Constraint{getType(n.Inner), &typing.IntType{}})
			return &typing.UnitType{}
		case *IntToFloat:
			constraints = append(constraints,
				typing.Constraint{getType(n.Inner), &typing.IntType{}})
			return &typing.FloatType{}
		case *FloatToInt:
			constraints = append(constraints,
				typing.Constraint{getType(n.Inner), &typing.FloatType{}})
			return &typing.IntType{}
		case *Sqrt:
			constraints = append(constraints,
				typing.Constraint{getType(n.Inner), &typing.FloatType{}})
			return &typing.FloatType{}
		}

		panic("invalid node type")
	}

	getType(root)

	mapping := typing.Unify(constraints)

	for name, t := range nameToType {
		nameToType[name] = t.Replace(mapping, true)
	}

	if _, ok := root.GetType(nameToType).(*typing.UnitType); !ok {
		panic("the program should be of unit type")
	}

	return nameToType
}
