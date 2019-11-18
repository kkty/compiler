package ast

type Node interface {
	astNode()
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
type LessThanOrEqual struct{ Left, Right Node }
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
type PrintChar struct{ Inner Node }
type IntToFloat struct{ Inner Node }
type FloatToInt struct{ Inner Node }
type Sqrt struct{ Inner Node }

func (n Variable) astNode()        {}
func (n Unit) astNode()            {}
func (n Int) astNode()             {}
func (n Bool) astNode()            {}
func (n Float) astNode()           {}
func (n Add) astNode()             {}
func (n Sub) astNode()             {}
func (n FloatAdd) astNode()        {}
func (n FloatSub) astNode()        {}
func (n FloatDiv) astNode()        {}
func (n FloatMul) astNode()        {}
func (n Equal) astNode()           {}
func (n LessThanOrEqual) astNode() {}
func (n Neg) astNode()             {}
func (n FloatNeg) astNode()        {}
func (n Not) astNode()             {}
func (n If) astNode()              {}
func (n ValueBinding) astNode()    {}
func (n FunctionBinding) astNode() {}
func (n Application) astNode()     {}
func (n Tuple) astNode()           {}
func (n TupleBinding) astNode()    {}
func (n ArrayCreate) astNode()     {}
func (n ArrayGet) astNode()        {}
func (n ArrayPut) astNode()        {}
func (n ReadInt) astNode()         {}
func (n ReadFloat) astNode()       {}
func (n PrintInt) astNode()        {}
func (n PrintChar) astNode()       {}
func (n IntToFloat) astNode()      {}
func (n FloatToInt) astNode()      {}
func (n Sqrt) astNode()            {}
