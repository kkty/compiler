package mir

type Node interface {
	mirNode()
	FreeVariables(bound map[string]struct{}) []string
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

func (n Variable) mirNode()          {}
func (n Unit) mirNode()              {}
func (n Int) mirNode()               {}
func (n Bool) mirNode()              {}
func (n Float) mirNode()             {}
func (n Add) mirNode()               {}
func (n Sub) mirNode()               {}
func (n FloatAdd) mirNode()          {}
func (n FloatSub) mirNode()          {}
func (n FloatDiv) mirNode()          {}
func (n FloatMul) mirNode()          {}
func (n IfEqual) mirNode()           {}
func (n IfLessThanOrEqual) mirNode() {}
func (n ValueBinding) mirNode()      {}
func (n FunctionBinding) mirNode()   {}
func (n Application) mirNode()       {}
func (n Tuple) mirNode()             {}
func (n TupleBinding) mirNode()      {}
func (n ArrayCreate) mirNode()       {}
func (n ArrayGet) mirNode()          {}
func (n ArrayPut) mirNode()          {}
func (n ReadInt) mirNode()           {}
func (n ReadFloat) mirNode()         {}
func (n PrintInt) mirNode()          {}
func (n PrintChar) mirNode()         {}
func (n IntToFloat) mirNode()        {}
func (n FloatToInt) mirNode()        {}
func (n Sqrt) mirNode()              {}
func (n Neg) mirNode()               {}

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

func (n FunctionBinding) FreeVariables(bound map[string]struct{}) []string {
	bound = copyStringSet(bound)
	bound[n.Name] = struct{}{}
	for _, arg := range n.Args {
		bound[arg] = struct{}{}
	}
	return n.Body.FreeVariables(bound)
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

func (n TupleBinding) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}
	if _, ok := bound[n.Tuple]; !ok {
		ret = append(ret, n.Tuple)
	}
	bound = copyStringSet(bound)
	for _, name := range n.Names {
		bound[name] = struct{}{}
	}
	return append(ret, n.Next.FreeVariables(bound)...)
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
func (n Neg) FreeVariables(bound map[string]struct{}) []string {
	ret := []string{}

	if _, ok := bound[n.Arg]; !ok {
		ret = append(ret, n.Arg)
	}

	return ret
}
