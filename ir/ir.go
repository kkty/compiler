package ir

type Node interface {
	UpdateNames(mapping map[string]string) Node
	FreeVariables(bound map[string]struct{}) []string
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

type TupleGet struct {
	Tuple string
	Index int32
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
func (n ArrayCreate) irNode()       {}
func (n ArrayGet) irNode()          {}
func (n ArrayPut) irNode()          {}
func (n ReadInt) irNode()           {}
func (n ReadFloat) irNode()         {}
func (n PrintInt) irNode()          {}
func (n PrintChar) irNode()         {}
func (n IntToFloat) irNode()        {}
func (n FloatToInt) irNode()        {}
func (n Sqrt) irNode()              {}
func (n TupleGet) irNode()          {}

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
	return Sub{replaceIfFound(n.Left, mapping), replaceIfFound(n.Right, mapping)}
}

func (n FloatAdd) UpdateNames(mapping map[string]string) Node {
	return FloatAdd{replaceIfFound(n.Left, mapping), replaceIfFound(n.Right, mapping)}
}
func (n FloatSub) UpdateNames(mapping map[string]string) Node {
	return FloatSub{replaceIfFound(n.Left, mapping), replaceIfFound(n.Right, mapping)}
}
func (n FloatDiv) UpdateNames(mapping map[string]string) Node {
	return FloatDiv{replaceIfFound(n.Left, mapping), replaceIfFound(n.Right, mapping)}
}

func (n FloatMul) UpdateNames(mapping map[string]string) Node {
	return FloatMul{replaceIfFound(n.Left, mapping), replaceIfFound(n.Right, mapping)}
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

func (n ArrayCreate) UpdateNames(mapping map[string]string) Node {
	return ArrayCreate{replaceIfFound(n.Size, mapping), replaceIfFound(n.Value, mapping)}
}

func (n ArrayGet) UpdateNames(mapping map[string]string) Node {
	return ArrayGet{replaceIfFound(n.Array, mapping), replaceIfFound(n.Index, mapping)}
}

func (n ArrayPut) UpdateNames(mapping map[string]string) Node {
	return ArrayPut{replaceIfFound(n.Array, mapping), replaceIfFound(n.Index, mapping), replaceIfFound(n.Value, mapping)}
}

func (n ReadInt) UpdateNames(mapping map[string]string) Node   { return n }
func (n ReadFloat) UpdateNames(mapping map[string]string) Node { return n }

func (n PrintInt) UpdateNames(mapping map[string]string) Node {
	return PrintInt{replaceIfFound(n.Arg, mapping)}
}

func (n PrintChar) UpdateNames(mapping map[string]string) Node {
	return PrintChar{replaceIfFound(n.Arg, mapping)}
}
func (n IntToFloat) UpdateNames(mapping map[string]string) Node {
	return IntToFloat{replaceIfFound(n.Arg, mapping)}
}
func (n FloatToInt) UpdateNames(mapping map[string]string) Node {
	return FloatToInt{replaceIfFound(n.Arg, mapping)}
}
func (n Sqrt) UpdateNames(mapping map[string]string) Node {
	return Sqrt{replaceIfFound(n.Arg, mapping)}
}

func (n TupleGet) UpdateNames(mapping map[string]string) Node {
	return TupleGet{replaceIfFound(n.Tuple, mapping), n.Index}
}

func copyStringSet(original map[string]struct{}) map[string]struct{} {
	s := map[string]struct{}{}

	for k := range original {
		s[k] = struct{}{}
	}

	return s
}

func (n Variable) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}

	if _, ok := bound[n.Name]; !ok {
		ret = append(ret, n.Name)
	}

	return ret
}

func (n Unit) FreeVariables(bound map[string]struct{}) []string {
	return []string{}
}

func (n Int) FreeVariables(bound map[string]struct{}) []string {
	return []string{}
}

func (n Bool) FreeVariables(bound map[string]struct{}) []string {
	return []string{}
}

func (n Float) FreeVariables(bound map[string]struct{}) []string {
	return []string{}
}

func (n Add) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}
	if _, ok := bound[n.Left]; !ok {
		ret = append(ret, n.Left)
	}
	if _, ok := bound[n.Right]; !ok {
		ret = append(ret, n.Right)
	}
	return ret
}

func (n Sub) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}
	if _, ok := bound[n.Left]; !ok {
		ret = append(ret, n.Left)
	}
	if _, ok := bound[n.Right]; !ok {
		ret = append(ret, n.Right)
	}
	return ret
}

func (n FloatAdd) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}
	if _, ok := bound[n.Left]; !ok {
		ret = append(ret, n.Left)
	}
	if _, ok := bound[n.Right]; !ok {
		ret = append(ret, n.Right)
	}
	return ret
}

func (n FloatSub) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}
	if _, ok := bound[n.Left]; !ok {
		ret = append(ret, n.Left)
	}
	if _, ok := bound[n.Right]; !ok {
		ret = append(ret, n.Right)
	}
	return ret
}

func (n FloatDiv) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}
	if _, ok := bound[n.Left]; !ok {
		ret = append(ret, n.Left)
	}
	if _, ok := bound[n.Right]; !ok {
		ret = append(ret, n.Right)
	}
	return ret
}

func (n FloatMul) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}
	if _, ok := bound[n.Left]; !ok {
		ret = append(ret, n.Left)
	}
	if _, ok := bound[n.Right]; !ok {
		ret = append(ret, n.Right)
	}
	return ret
}

func (n IfEqual) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}

	if _, ok := bound[n.Left]; !ok {
		ret = append(ret, n.Left)
	}

	if _, ok := bound[n.Right]; !ok {
		ret = append(ret, n.Right)
	}

	ret = append(ret, n.True.FreeVariables(bound)...)
	ret = append(ret, n.False.FreeVariables(bound)...)

	return ret
}

func (n IfLessThanOrEqual) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}
	if _, ok := bound[n.Left]; !ok {
		ret = append(ret, n.Left)
	}

	if _, ok := bound[n.Right]; !ok {
		ret = append(ret, n.Right)
	}

	ret = append(ret, n.True.FreeVariables(bound)...)
	ret = append(ret, n.False.FreeVariables(bound)...)

	return ret
}

func (n ValueBinding) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}
	ret = append(ret, n.Value.FreeVariables(bound)...)
	bound = copyStringSet(bound)
	bound[n.Name] = struct{}{}
	ret = append(ret, n.Next.FreeVariables(bound)...)
	return ret
}

func (n Application) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}

	for _, arg := range n.Args {
		if _, ok := bound[arg]; !ok {
			ret = append(ret, arg)
		}
	}

	return ret
}

func (n Tuple) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}
	for _, element := range n.Elements {
		if _, ok := bound[element]; !ok {
			ret = append(ret, element)
		}
	}
	return ret
}

func (n ArrayCreate) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}
	if _, ok := bound[n.Size]; !ok {
		ret = append(ret, n.Size)
	}
	if _, ok := bound[n.Value]; !ok {
		ret = append(ret, n.Value)
	}
	return ret
}
func (n ArrayGet) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}
	if _, ok := bound[n.Array]; !ok {
		ret = append(ret, n.Array)
	}
	if _, ok := bound[n.Index]; !ok {
		ret = append(ret, n.Index)
	}
	return ret
}
func (n ArrayPut) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}
	if _, ok := bound[n.Array]; !ok {
		ret = append(ret, n.Array)
	}
	if _, ok := bound[n.Index]; !ok {
		ret = append(ret, n.Index)
	}
	if _, ok := bound[n.Value]; !ok {
		ret = append(ret, n.Value)
	}
	return ret
}

func (n ReadInt) FreeVariables(bound map[string]struct{}) []string   { return []string{} }
func (n ReadFloat) FreeVariables(bound map[string]struct{}) []string { return []string{} }

func (n PrintInt) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}

	if _, ok := bound[n.Arg]; !ok {
		ret = append(ret, n.Arg)
	}

	return ret
}

func (n PrintChar) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}

	if _, ok := bound[n.Arg]; !ok {
		ret = append(ret, n.Arg)
	}

	return ret
}
func (n IntToFloat) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}

	if _, ok := bound[n.Arg]; !ok {
		ret = append(ret, n.Arg)
	}

	return ret
}
func (n FloatToInt) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}

	if _, ok := bound[n.Arg]; !ok {
		ret = append(ret, n.Arg)
	}

	return ret
}
func (n Sqrt) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}

	if _, ok := bound[n.Arg]; !ok {
		ret = append(ret, n.Arg)
	}

	return ret
}

func (n TupleGet) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}

	if _, ok := bound[n.Tuple]; !ok {
		ret = append(ret, n.Tuple)
	}

	return ret
}
