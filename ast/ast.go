package ast

type Node interface {
	astNode()
	Children() []Node
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

type ValueBinding struct {
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

type Tuple struct{ Elements []Node }

type TupleBinding struct {
	Names       []string
	Tuple, Next Node
}

type ArrayCreate struct{ Size, Value Node }

type ArrayGet struct {
	Array, Index Node
}

type ArrayPut struct {
	Array, Index, Value Node
}

type ReadInt struct{}
type ReadFloat struct{}
type PrintInt struct{ Inner Node }
type WriteByte struct{ Inner Node }
type IntToFloat struct{ Inner Node }
type FloatToInt struct{ Inner Node }
type Sqrt struct{ Inner Node }

func (n *Variable) astNode()        {}
func (n *Unit) astNode()            {}
func (n *Int) astNode()             {}
func (n *Bool) astNode()            {}
func (n *Float) astNode()           {}
func (n *Add) astNode()             {}
func (n *Sub) astNode()             {}
func (n *FloatAdd) astNode()        {}
func (n *FloatSub) astNode()        {}
func (n *FloatDiv) astNode()        {}
func (n *FloatMul) astNode()        {}
func (n *Equal) astNode()           {}
func (n *LessThan) astNode()        {}
func (n *Neg) astNode()             {}
func (n *FloatNeg) astNode()        {}
func (n *Not) astNode()             {}
func (n *If) astNode()              {}
func (n *ValueBinding) astNode()    {}
func (n *FunctionBinding) astNode() {}
func (n *Application) astNode()     {}
func (n *Tuple) astNode()           {}
func (n *TupleBinding) astNode()    {}
func (n *ArrayCreate) astNode()     {}
func (n *ArrayGet) astNode()        {}
func (n *ArrayPut) astNode()        {}
func (n *ReadInt) astNode()         {}
func (n *ReadFloat) astNode()       {}
func (n *PrintInt) astNode()        {}
func (n *WriteByte) astNode()       {}
func (n *IntToFloat) astNode()      {}
func (n *FloatToInt) astNode()      {}
func (n *Sqrt) astNode()            {}

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
func (n *ValueBinding) Children() []Node    { return []Node{n.Body, n.Next} }
func (n *FunctionBinding) Children() []Node { return []Node{n.Body, n.Next} }
func (n *Application) Children() []Node     { return n.Args }
func (n *Tuple) Children() []Node           { return n.Elements }
func (n *TupleBinding) Children() []Node    { return []Node{n.Tuple, n.Next} }
func (n *ArrayCreate) Children() []Node     { return []Node{n.Size, n.Value} }
func (n *ArrayGet) Children() []Node        { return []Node{n.Array, n.Index} }
func (n *ArrayPut) Children() []Node        { return []Node{n.Array, n.Index, n.Value} }
func (n *ReadInt) Children() []Node         { return []Node{} }
func (n *ReadFloat) Children() []Node       { return []Node{} }
func (n *PrintInt) Children() []Node        { return []Node{n.Inner} }
func (n *WriteByte) Children() []Node       { return []Node{n.Inner} }
func (n *IntToFloat) Children() []Node      { return []Node{n.Inner} }
func (n *FloatToInt) Children() []Node      { return []Node{n.Inner} }
func (n *Sqrt) Children() []Node            { return []Node{n.Inner} }
