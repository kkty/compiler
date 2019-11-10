package ir

type Node interface {
	UpdateNames(mapping map[string]string) Node
	irNode()
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

type IfLessThanOrEqual struct {
	Left, Right string
	True, False Node
}

type ValueBinding struct {
	Name        string
	Value, Next Node
}

type Function struct {
	Name string
	Args []string
	Body Node
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

func (n Variable) irNode()          {}
func (n Unit) irNode()              {}
func (n Int) irNode()               {}
func (n Bool) irNode()              {}
func (n Float) irNode()             {}
func (n Add) irNode()               {}
func (n Sub) irNode()               {}
func (n FloatAdd) irNode()          {}
func (n FloatSub) irNode()          {}
func (n FloatDiv) irNode()          {}
func (n FloatMul) irNode()          {}
func (n IfEqual) irNode()           {}
func (n IfLessThanOrEqual) irNode() {}
func (n ValueBinding) irNode()      {}
func (n Application) irNode()       {}
func (n Tuple) irNode()             {}
func (n TupleBinding) irNode()      {}
func (n ArrayCreate) irNode()       {}
func (n ArrayGet) irNode()          {}
func (n ArrayPut) irNode()          {}

func replaceIfFound(k string, m map[string]string) string {
	if v, ok := m[k]; ok {
		return v
	}

	return k
}

func (n Variable) UpdateNames(mapping map[string]string) Node {
	return Variable{replaceIfFound(n.Name, mapping)}
}

func (n Unit) UpdateNames(mapping map[string]string) Node  { return n }
func (n Int) UpdateNames(mapping map[string]string) Node   { return n }
func (n Bool) UpdateNames(mapping map[string]string) Node  { return n }
func (n Float) UpdateNames(mapping map[string]string) Node { return n }

func (n Add) UpdateNames(mapping map[string]string) Node {
	return Add{replaceIfFound(n.Left, mapping), replaceIfFound(n.Right, mapping)}
}
func (n Sub) UpdateNames(mapping map[string]string) Node {
	return Add{replaceIfFound(n.Left, mapping), replaceIfFound(n.Right, mapping)}
}

func (n FloatAdd) UpdateNames(mapping map[string]string) Node {
	return Add{replaceIfFound(n.Left, mapping), replaceIfFound(n.Right, mapping)}
}
func (n FloatSub) UpdateNames(mapping map[string]string) Node {
	return Add{replaceIfFound(n.Left, mapping), replaceIfFound(n.Right, mapping)}
}
func (n FloatDiv) UpdateNames(mapping map[string]string) Node {
	return Add{replaceIfFound(n.Left, mapping), replaceIfFound(n.Right, mapping)}
}

func (n FloatMul) UpdateNames(mapping map[string]string) Node {
	return Add{replaceIfFound(n.Left, mapping), replaceIfFound(n.Right, mapping)}
}

func (n IfEqual) UpdateNames(mapping map[string]string) Node {
	return IfEqual{replaceIfFound(n.Left, mapping), replaceIfFound(n.Right, mapping), n.True.UpdateNames(mapping), n.False.UpdateNames(mapping)}
}

func (n IfLessThanOrEqual) UpdateNames(mapping map[string]string) Node {
	return IfLessThanOrEqual{replaceIfFound(n.Left, mapping), replaceIfFound(n.Right, mapping), n.True.UpdateNames(mapping), n.False.UpdateNames(mapping)}
}

func (n ValueBinding) UpdateNames(mapping map[string]string) Node {
	return ValueBinding{n.Name, n.Value.UpdateNames(mapping), n.Next.UpdateNames(mapping)}
}

func (n Application) UpdateNames(mapping map[string]string) Node {
	args := []string{}
	for _, arg := range n.Args {
		args = append(args, replaceIfFound(arg, mapping))
	}
	return Application{n.Function, args}
}

func (n Tuple) UpdateNames(mapping map[string]string) Node {
	elements := []string{}
	for _, element := range n.Elements {
		elements = append(elements, replaceIfFound(element, mapping))
	}
	return Tuple{elements}
}

func (n TupleBinding) UpdateNames(mapping map[string]string) Node {
	return TupleBinding{n.Names, replaceIfFound(n.Tuple, mapping), n.Next.UpdateNames(mapping)}
}

func (n ArrayCreate) UpdateNames(mapping map[string]string) Node {
	return ArrayCreate{replaceIfFound(n.Size, mapping), replaceIfFound(n.Value, mapping)}
}

func (n ArrayGet) UpdateNames(mapping map[string]string) Node {
	return ArrayGet{replaceIfFound(n.Array, mapping), replaceIfFound(n.Index, mapping)}
}

func (n ArrayPut) UpdateNames(mapping map[string]string) Node {
	return ArrayPut{replaceIfFound(n.Array, mapping), replaceIfFound(n.Index, mapping), replaceIfFound(n.Value, mapping)}
}
