package ast

import (
	"github.com/kkty/compiler/typing"
)

type Node interface {
	Children() []Node
	GetType(map[string]typing.Type) typing.Type
}

type Variable struct{ Name string }
type Unit struct{}
type Int struct{ Value int32 }
type Bool struct{ Value bool }
type Float struct{ Value float32 }
type Add struct{ Left, Right Node }
type Sub struct{ Left, Right Node }
type FloatAdd struct{ Left, Right Node }
type FloatSub struct{ Left, Right Node }
type FloatDiv struct{ Left, Right Node }
type FloatMul struct{ Left, Right Node }
type Equal struct{ Left, Right Node }
type LessThan struct{ Left, Right Node }
type Neg struct{ Inner Node }
type FloatNeg struct{ Inner Node }
type Not struct{ Inner Node }
type If struct{ Condition, True, False Node }

type Assignment struct {
	Name       string
	Body, Next Node
}

type FunctionBinding struct {
	Name       string
	Args       []string
	Body, Next Node
}

type Application struct {
	Function string
	Args     []Node
}

type Tuple struct {
	Elements []Node
}

type TupleBinding struct {
	Names       []string
	Tuple, Next Node
}

type ArrayCreate struct{ Size, Value Node }
type ArrayGet struct{ Array, Index Node }
type ArrayPut struct{ Array, Index, Value Node }
type ReadInt struct{}
type ReadFloat struct{}
type WriteByte struct{ Inner Node }
type IntToFloat struct{ Inner Node }
type FloatToInt struct{ Inner Node }
type Sqrt struct{ Inner Node }

func (n *Variable) GetType(nameToType map[string]typing.Type) typing.Type { return nameToType[n.Name] }
func (n *Unit) GetType(nameToType map[string]typing.Type) typing.Type     { return typing.UnitType }
func (n *Int) GetType(nameToType map[string]typing.Type) typing.Type      { return typing.IntType }
func (n *Bool) GetType(nameToType map[string]typing.Type) typing.Type     { return typing.BoolType }
func (n *Float) GetType(nameToType map[string]typing.Type) typing.Type    { return typing.FloatType }
func (n *Add) GetType(nameToType map[string]typing.Type) typing.Type      { return typing.IntType }
func (n *Sub) GetType(nameToType map[string]typing.Type) typing.Type      { return typing.IntType }
func (n *FloatAdd) GetType(nameToType map[string]typing.Type) typing.Type { return typing.FloatType }
func (n *FloatSub) GetType(nameToType map[string]typing.Type) typing.Type { return typing.FloatType }
func (n *FloatDiv) GetType(nameToType map[string]typing.Type) typing.Type { return typing.FloatType }
func (n *FloatMul) GetType(nameToType map[string]typing.Type) typing.Type { return typing.FloatType }
func (n *Equal) GetType(nameToType map[string]typing.Type) typing.Type    { return typing.BoolType }
func (n *LessThan) GetType(nameToType map[string]typing.Type) typing.Type { return typing.BoolType }

func (n *Neg) GetType(nameToType map[string]typing.Type) typing.Type {
	return n.Inner.GetType(nameToType)
}

func (n *FloatNeg) GetType(nameToType map[string]typing.Type) typing.Type { return typing.FloatType }
func (n *Not) GetType(nameToType map[string]typing.Type) typing.Type      { return typing.BoolType }
func (n *If) GetType(nameToType map[string]typing.Type) typing.Type       { return n.True.GetType(nameToType) }

func (n *Assignment) GetType(nameToType map[string]typing.Type) typing.Type {
	return n.Next.GetType(nameToType)
}

func (n *FunctionBinding) GetType(nameToType map[string]typing.Type) typing.Type {
	return n.Next.GetType(nameToType)
}

func (n *Application) GetType(nameToType map[string]typing.Type) typing.Type {
	return nameToType[n.Function].(typing.FunctionType).Return
}

func (n *Tuple) GetType(nameToType map[string]typing.Type) typing.Type {
	elementTypes := []typing.Type{}
	for _, element := range n.Elements {
		elementTypes = append(elementTypes, element.GetType(nameToType))
	}
	return typing.TupleType{elementTypes}
}

func (n *TupleBinding) GetType(nameToType map[string]typing.Type) typing.Type {
	return n.Next.GetType(nameToType)
}

func (n *ArrayCreate) GetType(nameToType map[string]typing.Type) typing.Type {
	return typing.ArrayType{Inner: n.Value.GetType(nameToType)}
}

func (n *ArrayGet) GetType(nameToType map[string]typing.Type) typing.Type {
	return n.Array.GetType(nameToType).(typing.ArrayType).Inner
}

func (n *ArrayPut) GetType(nameToType map[string]typing.Type) typing.Type   { return typing.UnitType }
func (n *ReadInt) GetType(nameToType map[string]typing.Type) typing.Type    { return typing.IntType }
func (n *ReadFloat) GetType(nameToType map[string]typing.Type) typing.Type  { return typing.FloatType }
func (n *WriteByte) GetType(nameToType map[string]typing.Type) typing.Type  { return typing.UnitType }
func (n *IntToFloat) GetType(nameToType map[string]typing.Type) typing.Type { return typing.FloatType }
func (n *FloatToInt) GetType(nameToType map[string]typing.Type) typing.Type { return typing.IntType }
func (n *Sqrt) GetType(nameToType map[string]typing.Type) typing.Type       { return typing.FloatType }

func (n *Variable) Children() []Node        { return []Node{} }
func (n *Unit) Children() []Node            { return []Node{} }
func (n *Int) Children() []Node             { return []Node{} }
func (n *Bool) Children() []Node            { return []Node{} }
func (n *Float) Children() []Node           { return []Node{} }
func (n *Add) Children() []Node             { return []Node{n.Left, n.Right} }
func (n *Sub) Children() []Node             { return []Node{n.Left, n.Right} }
func (n *FloatAdd) Children() []Node        { return []Node{n.Left, n.Right} }
func (n *FloatSub) Children() []Node        { return []Node{n.Left, n.Right} }
func (n *FloatDiv) Children() []Node        { return []Node{n.Left, n.Right} }
func (n *FloatMul) Children() []Node        { return []Node{n.Left, n.Right} }
func (n *Equal) Children() []Node           { return []Node{n.Left, n.Right} }
func (n *LessThan) Children() []Node        { return []Node{n.Left, n.Right} }
func (n *Neg) Children() []Node             { return []Node{n.Inner} }
func (n *FloatNeg) Children() []Node        { return []Node{n.Inner} }
func (n *Not) Children() []Node             { return []Node{n.Inner} }
func (n *If) Children() []Node              { return []Node{n.Condition, n.True, n.False} }
func (n *Assignment) Children() []Node    { return []Node{n.Body, n.Next} }
func (n *FunctionBinding) Children() []Node { return []Node{n.Body, n.Next} }
func (n *Application) Children() []Node     { return n.Args }
func (n *Tuple) Children() []Node           { return n.Elements }
func (n *TupleBinding) Children() []Node    { return []Node{n.Tuple, n.Next} }
func (n *ArrayCreate) Children() []Node     { return []Node{n.Size, n.Value} }
func (n *ArrayGet) Children() []Node        { return []Node{n.Array, n.Index} }
func (n *ArrayPut) Children() []Node        { return []Node{n.Array, n.Index, n.Value} }
func (n *ReadInt) Children() []Node         { return []Node{} }
func (n *ReadFloat) Children() []Node       { return []Node{} }
func (n *WriteByte) Children() []Node       { return []Node{n.Inner} }
func (n *IntToFloat) Children() []Node      { return []Node{n.Inner} }
func (n *FloatToInt) Children() []Node      { return []Node{n.Inner} }
func (n *Sqrt) Children() []Node            { return []Node{n.Inner} }
