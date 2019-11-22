package mir

type Node interface {
	mirNode()
}

type Variable struct{ Name string }
type Unit struct{}
type Int struct{ Value int32 }
type Bool struct{ Value bool }
type Float struct{ Value float32 }

type Add struct{ Left, Right string }
type Sub struct{ Left, Right string }
type FloatAdd struct{ Left, Right string }
type FloatSub struct{ Left, Right string }
type FloatDiv struct{ Left, Right string }
type FloatMul struct{ Left, Right string }

type IfEqual struct {
	Left, Right string
	True, False Node
}

type IfLessThan struct {
	Left, Right string
	True, False Node
}

type ValueBinding struct {
	Name        string
	Value, Next Node
}

type FunctionBinding struct {
	Name       string
	Args       []string
	Body, Next Node
}

type Application struct {
	Function string
	Args     []string
}

type Tuple struct{ Elements []string }

type TupleBinding struct {
	Names []string
	Tuple string
	Next  Node
}

type ArrayCreate struct{ Size, Value string }
type ArrayGet struct{ Array, Index string }
type ArrayPut struct{ Array, Index, Value string }

type ReadInt struct{}
type ReadFloat struct{}
type PrintInt struct{ Arg string }
type PrintChar struct{ Arg string }
type IntToFloat struct{ Arg string }
type FloatToInt struct{ Arg string }
type Sqrt struct{ Arg string }
type Neg struct{ Arg string }

func (n *Variable) mirNode()        {}
func (n *Unit) mirNode()            {}
func (n *Int) mirNode()             {}
func (n *Bool) mirNode()            {}
func (n *Float) mirNode()           {}
func (n *Add) mirNode()             {}
func (n *Sub) mirNode()             {}
func (n *FloatAdd) mirNode()        {}
func (n *FloatSub) mirNode()        {}
func (n *FloatDiv) mirNode()        {}
func (n *FloatMul) mirNode()        {}
func (n *IfEqual) mirNode()         {}
func (n *IfLessThan) mirNode()      {}
func (n *ValueBinding) mirNode()    {}
func (n *FunctionBinding) mirNode() {}
func (n *Application) mirNode()     {}
func (n *Tuple) mirNode()           {}
func (n *TupleBinding) mirNode()    {}
func (n *ArrayCreate) mirNode()     {}
func (n *ArrayGet) mirNode()        {}
func (n *ArrayPut) mirNode()        {}
func (n *ReadInt) mirNode()         {}
func (n *ReadFloat) mirNode()       {}
func (n *PrintInt) mirNode()        {}
func (n *PrintChar) mirNode()       {}
func (n *IntToFloat) mirNode()      {}
func (n *FloatToInt) mirNode()      {}
func (n *Sqrt) mirNode()            {}
func (n *Neg) mirNode()             {}
