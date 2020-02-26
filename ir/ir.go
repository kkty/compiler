package ir

import (
	"math"

	"github.com/kkty/compiler/stringmap"
	"github.com/kkty/compiler/stringset"
)

type Function struct {
	Name string
	Args []string
	Body Node
}

func FunctionsWithoutSideEffects(functions []*Function) stringset.Set {
	functionsWithoutSideEffects := stringset.New()
	n := 0
	for {
		for _, function := range functions {
			functionsWithoutSideEffects.Add(function.Name)

			if function.Body.HasSideEffects(functionsWithoutSideEffects) {
				delete(functionsWithoutSideEffects, function.Name)
			}
		}

		if len(functionsWithoutSideEffects) > n {
			n = len(functionsWithoutSideEffects)
		} else {
			break
		}
	}
	return functionsWithoutSideEffects
}

func (f Function) FreeVariables() stringset.Set {
	bound := stringset.New()

	bound.Add(f.Name)

	for _, arg := range f.Args {
		bound.Add(arg)
	}

	return f.Body.FreeVariables(bound)
}

func (f *Function) IsRecursive() bool {
	applications := f.Body.Applications()

	for _, application := range applications {
		if application.Function == f.Name {
			return true
		}
	}

	return false
}

type Node interface {
	UpdateNames(mapping stringmap.Map)
	FreeVariables(bound stringset.Set) stringset.Set
	FloatValues() []float32
	Clone() Node
	HasSideEffects(functionsWithoutSideEffects stringset.Set) bool
	Applications() []*Application
	Size() int
	Evaluate(map[string]interface{}, []*Function) interface{}
}

type Variable struct{ Name string }
type Unit struct{}
type Int struct{ Value int32 }
type Bool struct{ Value bool }
type Float struct{ Value float32 }

type Add struct{ Left, Right string }

type AddImmediate struct {
	Left  string
	Right int32
}

type Sub struct{ Left, Right string }
type SubFromZero struct{ Inner string }
type FloatAdd struct{ Left, Right string }
type FloatSub struct{ Left, Right string }
type FloatSubFromZero struct{ Inner string }
type FloatDiv struct{ Left, Right string }
type FloatMul struct{ Left, Right string }

type Not struct{ Inner string }
type Equal struct{ Left, Right string }
type EqualZero struct{ Inner string }
type LessThan struct{ Left, Right string }
type LessThanFloat struct{ Left, Right string }
type LessThanZero struct{ Inner string }
type LessThanZeroFloat struct{ Inner string }
type GreaterThanZero struct{ Inner string }
type GreaterThanZeroFloat struct{ Inner string }

type IfEqual struct {
	Left, Right string
	True, False Node
}

type IfEqualZero struct {
	Inner       string
	True, False Node
}

type IfEqualTrue struct {
	Inner       string
	True, False Node
}

type IfLessThan struct {
	Left, Right string
	True, False Node
}

type IfLessThanFloat struct {
	Left, Right string
	True, False Node
}

type IfLessThanZero struct {
	Inner       string
	True, False Node
}

type IfLessThanZeroFloat struct {
	Inner       string
	True, False Node
}

type Assignment struct {
	Name        string
	Value, Next Node
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

type ArrayCreate struct{ Length, Value string }

type ArrayCreateImmediate struct {
	Length int32
	Value  string
}

type ArrayGet struct{ Array, Index string }

type ArrayGetImmediate struct {
	Array string
	Index int32
}

type ArrayPut struct{ Array, Index, Value string }

type ArrayPutImmediate struct {
	Array string
	Index int32
	Value string
}

type ReadInt struct{}
type ReadFloat struct{}
type WriteByte struct{ Arg string }
type IntToFloat struct{ Arg string }
type FloatToInt struct{ Arg string }
type Sqrt struct{ Arg string }

func replaceIfFound(k string, m stringmap.Map) string {
	if v, ok := m[k]; ok {
		return v
	}
	return k
}

func (n *Variable) UpdateNames(mapping stringmap.Map) {
	n.Name = replaceIfFound(n.Name, mapping)
}

func (n *Unit) UpdateNames(mapping stringmap.Map)  {}
func (n *Int) UpdateNames(mapping stringmap.Map)   {}
func (n *Bool) UpdateNames(mapping stringmap.Map)  {}
func (n *Float) UpdateNames(mapping stringmap.Map) {}

func (n *Add) UpdateNames(mapping stringmap.Map) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *AddImmediate) UpdateNames(mapping stringmap.Map) {
	n.Left = replaceIfFound(n.Left, mapping)
}

func (n *Sub) UpdateNames(mapping stringmap.Map) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *SubFromZero) UpdateNames(mapping stringmap.Map) {
	n.Inner = replaceIfFound(n.Inner, mapping)
}

func (n *FloatAdd) UpdateNames(mapping stringmap.Map) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *FloatSub) UpdateNames(mapping stringmap.Map) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *FloatSubFromZero) UpdateNames(mapping stringmap.Map) {
	n.Inner = replaceIfFound(n.Inner, mapping)
}

func (n *FloatDiv) UpdateNames(mapping stringmap.Map) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *FloatMul) UpdateNames(mapping stringmap.Map) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *Not) UpdateNames(mapping stringmap.Map) {
	n.Inner = replaceIfFound(n.Inner, mapping)
}

func (n *Equal) UpdateNames(mapping stringmap.Map) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *EqualZero) UpdateNames(mapping stringmap.Map) {
	n.Inner = replaceIfFound(n.Inner, mapping)
}

func (n *LessThan) UpdateNames(mapping stringmap.Map) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *LessThanFloat) UpdateNames(mapping stringmap.Map) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *LessThanZero) UpdateNames(mapping stringmap.Map) {
	n.Inner = replaceIfFound(n.Inner, mapping)
}

func (n *LessThanZeroFloat) UpdateNames(mapping stringmap.Map) {
	n.Inner = replaceIfFound(n.Inner, mapping)
}

func (n *GreaterThanZero) UpdateNames(mapping stringmap.Map) {
	n.Inner = replaceIfFound(n.Inner, mapping)
}

func (n *GreaterThanZeroFloat) UpdateNames(mapping stringmap.Map) {
	n.Inner = replaceIfFound(n.Inner, mapping)
}

func (n *IfEqual) UpdateNames(mapping stringmap.Map) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
	n.True.UpdateNames(mapping)
	n.False.UpdateNames(mapping)
}

func (n *IfEqualZero) UpdateNames(mapping stringmap.Map) {
	n.Inner = replaceIfFound(n.Inner, mapping)
	n.True.UpdateNames(mapping)
	n.False.UpdateNames(mapping)
}

func (n *IfEqualTrue) UpdateNames(mapping stringmap.Map) {
	n.Inner = replaceIfFound(n.Inner, mapping)
	n.True.UpdateNames(mapping)
	n.False.UpdateNames(mapping)
}

func (n *IfLessThan) UpdateNames(mapping stringmap.Map) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
	n.True.UpdateNames(mapping)
	n.False.UpdateNames(mapping)
}

func (n *IfLessThanFloat) UpdateNames(mapping stringmap.Map) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
	n.True.UpdateNames(mapping)
	n.False.UpdateNames(mapping)
}

func (n *IfLessThanZero) UpdateNames(mapping stringmap.Map) {
	n.Inner = replaceIfFound(n.Inner, mapping)
	n.True.UpdateNames(mapping)
	n.False.UpdateNames(mapping)
}

func (n *IfLessThanZeroFloat) UpdateNames(mapping stringmap.Map) {
	n.Inner = replaceIfFound(n.Inner, mapping)
	n.True.UpdateNames(mapping)
	n.False.UpdateNames(mapping)
}

func (n *Assignment) UpdateNames(mapping stringmap.Map) {
	n.Name = replaceIfFound(n.Name, mapping)
	n.Value.UpdateNames(mapping)
	n.Next.UpdateNames(mapping)
}

func (n *Application) UpdateNames(mapping stringmap.Map) {
	for i := range n.Args {
		n.Args[i] = replaceIfFound(n.Args[i], mapping)
	}
}

func (n *Tuple) UpdateNames(mapping stringmap.Map) {
	for i := range n.Elements {
		n.Elements[i] = replaceIfFound(n.Elements[i], mapping)
	}
}

func (n *ArrayCreate) UpdateNames(mapping stringmap.Map) {
	n.Length = replaceIfFound(n.Length, mapping)
	n.Value = replaceIfFound(n.Value, mapping)
}

func (n *ArrayCreateImmediate) UpdateNames(mapping stringmap.Map) {
	n.Value = replaceIfFound(n.Value, mapping)
}

func (n *ArrayGet) UpdateNames(mapping stringmap.Map) {
	n.Array = replaceIfFound(n.Array, mapping)
	n.Index = replaceIfFound(n.Index, mapping)
}

func (n *ArrayGetImmediate) UpdateNames(mapping stringmap.Map) {
	n.Array = replaceIfFound(n.Array, mapping)
}

func (n *ArrayPut) UpdateNames(mapping stringmap.Map) {
	n.Array = replaceIfFound(n.Array, mapping)
	n.Index = replaceIfFound(n.Index, mapping)
	n.Value = replaceIfFound(n.Value, mapping)
}

func (n *ArrayPutImmediate) UpdateNames(mapping stringmap.Map) {
	n.Array = replaceIfFound(n.Array, mapping)
	n.Value = replaceIfFound(n.Value, mapping)
}

func (n *ReadInt) UpdateNames(mapping stringmap.Map)   {}
func (n *ReadFloat) UpdateNames(mapping stringmap.Map) {}

func (n *WriteByte) UpdateNames(mapping stringmap.Map) {
	n.Arg = replaceIfFound(n.Arg, mapping)
}
func (n *IntToFloat) UpdateNames(mapping stringmap.Map) {
	n.Arg = replaceIfFound(n.Arg, mapping)
}
func (n *FloatToInt) UpdateNames(mapping stringmap.Map) {
	n.Arg = replaceIfFound(n.Arg, mapping)
}
func (n *Sqrt) UpdateNames(mapping stringmap.Map) {
	n.Arg = replaceIfFound(n.Arg, mapping)
}

func (n *TupleGet) UpdateNames(mapping stringmap.Map) {
	n.Tuple = replaceIfFound(n.Tuple, mapping)
}

func (n *Variable) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Name) {
		ret.Add(n.Name)
	}
	return ret
}

func (n *Unit) FreeVariables(bound stringset.Set) stringset.Set {
	return stringset.New()
}

func (n *Int) FreeVariables(bound stringset.Set) stringset.Set {
	return stringset.New()
}

func (n *Bool) FreeVariables(bound stringset.Set) stringset.Set {
	return stringset.New()
}

func (n *Float) FreeVariables(bound stringset.Set) stringset.Set {
	return stringset.New()
}

func (n *Add) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Left) {
		ret.Add(n.Left)
	}
	if !bound.Has(n.Right) {
		ret.Add(n.Right)
	}
	return ret
}

func (n *AddImmediate) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Left) {
		ret.Add(n.Left)
	}
	return ret
}

func (n *Sub) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Left) {
		ret.Add(n.Left)
	}
	if !bound.Has(n.Right) {
		ret.Add(n.Right)
	}
	return ret
}

func (n *SubFromZero) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Inner) {
		ret.Add(n.Inner)
	}
	return ret
}

func (n *FloatAdd) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Left) {
		ret.Add(n.Left)
	}
	if !bound.Has(n.Right) {
		ret.Add(n.Right)
	}
	return ret
}

func (n *FloatSub) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Left) {
		ret.Add(n.Left)
	}
	if !bound.Has(n.Right) {
		ret.Add(n.Right)
	}
	return ret
}

func (n *FloatSubFromZero) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Inner) {
		ret.Add(n.Inner)
	}
	return ret
}

func (n *FloatDiv) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Left) {
		ret.Add(n.Left)
	}
	if !bound.Has(n.Right) {
		ret.Add(n.Right)
	}
	return ret
}

func (n *FloatMul) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Left) {
		ret.Add(n.Left)
	}
	if !bound.Has(n.Right) {
		ret.Add(n.Right)
	}
	return ret
}

func (n *Not) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Inner) {
		ret.Add(n.Inner)
	}
	return ret
}

func (n *Equal) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Left) {
		ret.Add(n.Left)
	}
	if !bound.Has(n.Right) {
		ret.Add(n.Right)
	}
	return ret
}

func (n *EqualZero) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Inner) {
		ret.Add(n.Inner)
	}
	return ret
}

func (n *LessThan) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Left) {
		ret.Add(n.Left)
	}
	if !bound.Has(n.Right) {
		ret.Add(n.Right)
	}
	return ret
}

func (n *LessThanZero) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Inner) {
		ret.Add(n.Inner)
	}
	return ret
}

func (n *LessThanZeroFloat) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Inner) {
		ret.Add(n.Inner)
	}
	return ret
}

func (n *GreaterThanZero) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Inner) {
		ret.Add(n.Inner)
	}
	return ret
}

func (n *GreaterThanZeroFloat) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Inner) {
		ret.Add(n.Inner)
	}
	return ret
}

func (n *LessThanFloat) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Left) {
		ret.Add(n.Left)
	}
	if !bound.Has(n.Right) {
		ret.Add(n.Right)
	}
	return ret
}

func (n *IfEqual) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Left) {
		ret.Add(n.Left)
	}
	if !bound.Has(n.Right) {
		ret.Add(n.Right)
	}
	for v := range n.True.FreeVariables(bound) {
		ret.Add(v)
	}
	for v := range n.False.FreeVariables(bound) {
		ret.Add(v)
	}
	return ret
}

func (n *IfEqualZero) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Inner) {
		ret.Add(n.Inner)
	}
	for v := range n.True.FreeVariables(bound) {
		ret.Add(v)
	}
	for v := range n.False.FreeVariables(bound) {
		ret.Add(v)
	}
	return ret
}

func (n *IfEqualTrue) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Inner) {
		ret.Add(n.Inner)
	}
	for v := range n.True.FreeVariables(bound) {
		ret.Add(v)
	}
	for v := range n.False.FreeVariables(bound) {
		ret.Add(v)
	}
	return ret
}

func (n *IfLessThan) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Left) {
		ret.Add(n.Left)
	}
	if !bound.Has(n.Right) {
		ret.Add(n.Right)
	}
	for v := range n.True.FreeVariables(bound) {
		ret.Add(v)
	}
	for v := range n.False.FreeVariables(bound) {
		ret.Add(v)
	}
	return ret
}

func (n *IfLessThanFloat) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Left) {
		ret.Add(n.Left)
	}
	if !bound.Has(n.Right) {
		ret.Add(n.Right)
	}
	for v := range n.True.FreeVariables(bound) {
		ret.Add(v)
	}
	for v := range n.False.FreeVariables(bound) {
		ret.Add(v)
	}
	return ret
}

func (n *IfLessThanZero) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Inner) {
		ret.Add(n.Inner)
	}
	for v := range n.True.FreeVariables(bound) {
		ret.Add(v)
	}
	for v := range n.False.FreeVariables(bound) {
		ret.Add(v)
	}
	return ret
}

func (n *IfLessThanZeroFloat) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Inner) {
		ret.Add(n.Inner)
	}
	for v := range n.True.FreeVariables(bound) {
		ret.Add(v)
	}
	for v := range n.False.FreeVariables(bound) {
		ret.Add(v)
	}
	return ret
}

func (n *Assignment) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	for v := range n.Value.FreeVariables(bound) {
		ret.Add(v)
	}
	bound.Add(n.Name)
	for v := range n.Next.FreeVariables(bound) {
		ret.Add(v)
	}
	delete(bound, n.Name)
	return ret
}

func (n *Application) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	for _, arg := range n.Args {
		if !bound.Has(arg) {
			ret.Add(arg)
		}
	}
	return ret
}

func (n *Tuple) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	for _, element := range n.Elements {
		if !bound.Has(element) {
			ret.Add(element)
		}
	}
	return ret
}

func (n *ArrayCreate) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Length) {
		ret.Add(n.Length)
	}
	if !bound.Has(n.Value) {
		ret.Add(n.Value)
	}
	return ret
}

func (n *ArrayCreateImmediate) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Value) {
		ret.Add(n.Value)
	}
	return ret
}

func (n *ArrayGet) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Array) {
		ret.Add(n.Array)
	}
	if !bound.Has(n.Index) {
		ret.Add(n.Index)
	}
	return ret
}

func (n *ArrayGetImmediate) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Array) {
		ret.Add(n.Array)
	}
	return ret
}

func (n *ArrayPut) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Array) {
		ret.Add(n.Array)
	}
	if !bound.Has(n.Index) {
		ret.Add(n.Index)
	}
	if !bound.Has(n.Value) {
		ret.Add(n.Value)
	}
	return ret
}

func (n *ArrayPutImmediate) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Array) {
		ret.Add(n.Array)
	}
	if !bound.Has(n.Value) {
		ret.Add(n.Value)
	}
	return ret
}

func (n *ReadInt) FreeVariables(bound stringset.Set) stringset.Set {
	return stringset.New()
}

func (n *ReadFloat) FreeVariables(bound stringset.Set) stringset.Set {
	return stringset.New()
}

func (n *WriteByte) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Arg) {
		ret.Add(n.Arg)
	}
	return ret
}

func (n *IntToFloat) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Arg) {
		ret.Add(n.Arg)
	}
	return ret
}

func (n *FloatToInt) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Arg) {
		ret.Add(n.Arg)
	}
	return ret
}

func (n *Sqrt) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Arg) {
		ret.Add(n.Arg)
	}
	return ret
}

func (n *TupleGet) FreeVariables(bound stringset.Set) stringset.Set {
	ret := stringset.New()
	if !bound.Has(n.Tuple) {
		ret.Add(n.Tuple)
	}
	return ret
}

func (n *Variable) FloatValues() []float32             { return []float32{} }
func (n *Unit) FloatValues() []float32                 { return []float32{} }
func (n *Int) FloatValues() []float32                  { return []float32{} }
func (n *Bool) FloatValues() []float32                 { return []float32{} }
func (n *Float) FloatValues() []float32                { return []float32{n.Value} }
func (n *Add) FloatValues() []float32                  { return []float32{} }
func (n *AddImmediate) FloatValues() []float32         { return []float32{} }
func (n *Sub) FloatValues() []float32                  { return []float32{} }
func (n *SubFromZero) FloatValues() []float32          { return []float32{} }
func (n *FloatAdd) FloatValues() []float32             { return []float32{} }
func (n *FloatSub) FloatValues() []float32             { return []float32{} }
func (n *FloatSubFromZero) FloatValues() []float32     { return []float32{} }
func (n *FloatDiv) FloatValues() []float32             { return []float32{} }
func (n *FloatMul) FloatValues() []float32             { return []float32{} }
func (n *Not) FloatValues() []float32                  { return []float32{} }
func (n *Equal) FloatValues() []float32                { return []float32{} }
func (n *EqualZero) FloatValues() []float32            { return []float32{} }
func (n *LessThan) FloatValues() []float32             { return []float32{} }
func (n *LessThanFloat) FloatValues() []float32        { return []float32{} }
func (n *LessThanZero) FloatValues() []float32         { return []float32{} }
func (n *LessThanZeroFloat) FloatValues() []float32    { return []float32{} }
func (n *GreaterThanZero) FloatValues() []float32      { return []float32{} }
func (n *GreaterThanZeroFloat) FloatValues() []float32 { return []float32{} }

func (n *IfEqual) FloatValues() []float32 {
	return append(n.True.FloatValues(), n.False.FloatValues()...)
}

func (n *IfEqualZero) FloatValues() []float32 {
	return append(n.True.FloatValues(), n.False.FloatValues()...)
}

func (n *IfEqualTrue) FloatValues() []float32 {
	return append(n.True.FloatValues(), n.False.FloatValues()...)
}

func (n *IfLessThan) FloatValues() []float32 {
	return append(n.True.FloatValues(), n.False.FloatValues()...)
}

func (n *IfLessThanFloat) FloatValues() []float32 {
	return append(n.True.FloatValues(), n.False.FloatValues()...)
}

func (n *IfLessThanZero) FloatValues() []float32 {
	return append(n.True.FloatValues(), n.False.FloatValues()...)
}

func (n *IfLessThanZeroFloat) FloatValues() []float32 {
	return append(n.True.FloatValues(), n.False.FloatValues()...)
}

func (n *Assignment) FloatValues() []float32 {
	return append(n.Value.FloatValues(), n.Next.FloatValues()...)
}

func (n *Application) FloatValues() []float32          { return []float32{} }
func (n *Tuple) FloatValues() []float32                { return []float32{} }
func (n *TupleGet) FloatValues() []float32             { return []float32{} }
func (n *ArrayCreate) FloatValues() []float32          { return []float32{} }
func (n *ArrayCreateImmediate) FloatValues() []float32 { return []float32{} }
func (n *ArrayGet) FloatValues() []float32             { return []float32{} }
func (n *ArrayGetImmediate) FloatValues() []float32    { return []float32{} }
func (n *ArrayPut) FloatValues() []float32             { return []float32{} }
func (n *ArrayPutImmediate) FloatValues() []float32    { return []float32{} }
func (n *ReadInt) FloatValues() []float32              { return []float32{} }
func (n *ReadFloat) FloatValues() []float32            { return []float32{} }
func (n *WriteByte) FloatValues() []float32            { return []float32{} }
func (n *IntToFloat) FloatValues() []float32           { return []float32{} }
func (n *FloatToInt) FloatValues() []float32           { return []float32{} }
func (n *Sqrt) FloatValues() []float32                 { return []float32{} }

func (n *Variable) Clone() Node             { return &Variable{n.Name} }
func (n *Unit) Clone() Node                 { return &Unit{} }
func (n *Int) Clone() Node                  { return &Int{n.Value} }
func (n *Bool) Clone() Node                 { return &Bool{n.Value} }
func (n *Float) Clone() Node                { return &Float{n.Value} }
func (n *Add) Clone() Node                  { return &Add{n.Left, n.Right} }
func (n *AddImmediate) Clone() Node         { return &AddImmediate{n.Left, n.Right} }
func (n *Sub) Clone() Node                  { return &Sub{n.Left, n.Right} }
func (n *SubFromZero) Clone() Node          { return &SubFromZero{n.Inner} }
func (n *FloatAdd) Clone() Node             { return &FloatAdd{n.Left, n.Right} }
func (n *FloatSub) Clone() Node             { return &FloatSub{n.Left, n.Right} }
func (n *FloatSubFromZero) Clone() Node     { return &FloatSubFromZero{n.Inner} }
func (n *FloatDiv) Clone() Node             { return &FloatDiv{n.Left, n.Right} }
func (n *FloatMul) Clone() Node             { return &FloatMul{n.Left, n.Right} }
func (n *Not) Clone() Node                  { return &Not{n.Inner} }
func (n *Equal) Clone() Node                { return &Equal{n.Left, n.Right} }
func (n *EqualZero) Clone() Node            { return &EqualZero{n.Inner} }
func (n *LessThan) Clone() Node             { return &LessThan{n.Left, n.Right} }
func (n *LessThanFloat) Clone() Node        { return &LessThanFloat{n.Left, n.Right} }
func (n *LessThanZero) Clone() Node         { return &LessThanZero{n.Inner} }
func (n *LessThanZeroFloat) Clone() Node    { return &LessThanZeroFloat{n.Inner} }
func (n *GreaterThanZero) Clone() Node      { return &GreaterThanZero{n.Inner} }
func (n *GreaterThanZeroFloat) Clone() Node { return &GreaterThanZeroFloat{n.Inner} }

func (n *IfEqual) Clone() Node {
	return &IfEqual{n.Left, n.Right, n.True.Clone(), n.False.Clone()}
}

func (n *IfEqualZero) Clone() Node {
	return &IfEqualZero{n.Inner, n.True.Clone(), n.False.Clone()}
}

func (n *IfEqualTrue) Clone() Node {
	return &IfEqualTrue{n.Inner, n.True.Clone(), n.False.Clone()}
}

func (n *IfLessThan) Clone() Node {
	return &IfLessThan{n.Left, n.Right, n.True.Clone(), n.False.Clone()}
}

func (n *IfLessThanFloat) Clone() Node {
	return &IfLessThanFloat{n.Left, n.Right, n.True.Clone(), n.False.Clone()}
}

func (n *IfLessThanZero) Clone() Node {
	return &IfLessThanZero{n.Inner, n.True.Clone(), n.False.Clone()}
}

func (n *IfLessThanZeroFloat) Clone() Node {
	return &IfLessThanZeroFloat{n.Inner, n.True.Clone(), n.False.Clone()}
}

func (n *Assignment) Clone() Node {
	return &Assignment{n.Name, n.Value.Clone(), n.Next.Clone()}
}

func (n *Application) Clone() Node {
	args := []string{}
	for _, arg := range n.Args {
		args = append(args, arg)
	}
	return &Application{n.Function, args}
}

func (n *Tuple) Clone() Node {
	elements := []string{}
	for _, element := range n.Elements {
		elements = append(elements, element)
	}
	return &Tuple{elements}
}

func (n *TupleGet) Clone() Node {
	return &TupleGet{n.Tuple, n.Index}
}

func (n *ArrayCreate) Clone() Node {
	return &ArrayCreate{n.Length, n.Value}
}

func (n *ArrayCreateImmediate) Clone() Node {
	return &ArrayCreateImmediate{n.Length, n.Value}
}

func (n *ArrayGet) Clone() Node {
	return &ArrayGet{n.Array, n.Index}
}

func (n *ArrayGetImmediate) Clone() Node {
	return &ArrayGetImmediate{n.Array, n.Index}
}

func (n *ArrayPut) Clone() Node {
	return &ArrayPut{n.Array, n.Index, n.Value}
}

func (n *ArrayPutImmediate) Clone() Node {
	return &ArrayPutImmediate{n.Array, n.Index, n.Value}
}

func (n *ReadInt) Clone() Node    { return &ReadInt{} }
func (n *ReadFloat) Clone() Node  { return &ReadFloat{} }
func (n *WriteByte) Clone() Node  { return &WriteByte{n.Arg} }
func (n *IntToFloat) Clone() Node { return &IntToFloat{n.Arg} }
func (n *FloatToInt) Clone() Node { return &FloatToInt{n.Arg} }
func (n *Sqrt) Clone() Node       { return &Sqrt{n.Arg} }

func (n *Variable) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool { return false }
func (n *Unit) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool     { return false }
func (n *Int) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool      { return false }
func (n *Bool) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool     { return false }
func (n *Float) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool    { return false }
func (n *Add) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool      { return false }
func (n *AddImmediate) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return false
}
func (n *Sub) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool { return false }
func (n *SubFromZero) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return false
}
func (n *FloatAdd) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool { return false }
func (n *FloatSub) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool { return false }
func (n *FloatSubFromZero) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return false
}
func (n *FloatDiv) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool      { return false }
func (n *FloatMul) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool      { return false }
func (n *Not) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool           { return false }
func (n *Equal) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool         { return false }
func (n *EqualZero) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool     { return false }
func (n *LessThan) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool      { return false }
func (n *LessThanFloat) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool { return false }
func (n *LessThanZero) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool  { return false }
func (n *LessThanZeroFloat) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return false
}
func (n *GreaterThanZero) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool { return false }
func (n *GreaterThanZeroFloat) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return false
}

func (n *IfEqual) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return n.True.HasSideEffects(functionsWithoutSideEffects) || n.False.HasSideEffects(functionsWithoutSideEffects)
}

func (n *IfEqualZero) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return n.True.HasSideEffects(functionsWithoutSideEffects) || n.False.HasSideEffects(functionsWithoutSideEffects)
}

func (n *IfEqualTrue) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return n.True.HasSideEffects(functionsWithoutSideEffects) || n.False.HasSideEffects(functionsWithoutSideEffects)
}

func (n *IfLessThan) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return n.True.HasSideEffects(functionsWithoutSideEffects) || n.False.HasSideEffects(functionsWithoutSideEffects)
}

func (n *IfLessThanFloat) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return n.True.HasSideEffects(functionsWithoutSideEffects) || n.False.HasSideEffects(functionsWithoutSideEffects)
}

func (n *IfLessThanZero) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return n.True.HasSideEffects(functionsWithoutSideEffects) || n.False.HasSideEffects(functionsWithoutSideEffects)
}

func (n *IfLessThanZeroFloat) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return n.True.HasSideEffects(functionsWithoutSideEffects) || n.False.HasSideEffects(functionsWithoutSideEffects)
}

func (n *Assignment) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return n.Value.HasSideEffects(functionsWithoutSideEffects) || n.Next.HasSideEffects(functionsWithoutSideEffects)
}

func (n *Application) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return !functionsWithoutSideEffects.Has(n.Function)
}

func (n *Tuple) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool    { return false }
func (n *TupleGet) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool { return false }

func (n *ArrayCreate) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return false
}

func (n *ArrayCreateImmediate) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return false
}

func (n *ArrayGet) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool { return false }

func (n *ArrayGetImmediate) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return false
}

func (n *ArrayPut) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool { return true }

func (n *ArrayPutImmediate) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return true
}

func (n *ReadInt) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool   { return true }
func (n *ReadFloat) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool { return true }
func (n *WriteByte) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool { return true }
func (n *IntToFloat) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return false
}
func (n *FloatToInt) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool {
	return false
}
func (n *Sqrt) HasSideEffects(functionsWithoutSideEffects stringset.Set) bool { return false }

func (n *Variable) Applications() []*Application             { return []*Application{} }
func (n *Unit) Applications() []*Application                 { return []*Application{} }
func (n *Int) Applications() []*Application                  { return []*Application{} }
func (n *Bool) Applications() []*Application                 { return []*Application{} }
func (n *Float) Applications() []*Application                { return []*Application{} }
func (n *Add) Applications() []*Application                  { return []*Application{} }
func (n *AddImmediate) Applications() []*Application         { return []*Application{} }
func (n *Sub) Applications() []*Application                  { return []*Application{} }
func (n *SubFromZero) Applications() []*Application          { return []*Application{} }
func (n *FloatAdd) Applications() []*Application             { return []*Application{} }
func (n *FloatSub) Applications() []*Application             { return []*Application{} }
func (n *FloatSubFromZero) Applications() []*Application     { return []*Application{} }
func (n *FloatDiv) Applications() []*Application             { return []*Application{} }
func (n *FloatMul) Applications() []*Application             { return []*Application{} }
func (n *Not) Applications() []*Application                  { return []*Application{} }
func (n *Equal) Applications() []*Application                { return []*Application{} }
func (n *EqualZero) Applications() []*Application            { return []*Application{} }
func (n *LessThan) Applications() []*Application             { return []*Application{} }
func (n *LessThanFloat) Applications() []*Application        { return []*Application{} }
func (n *LessThanZero) Applications() []*Application         { return []*Application{} }
func (n *LessThanZeroFloat) Applications() []*Application    { return []*Application{} }
func (n *GreaterThanZero) Applications() []*Application      { return []*Application{} }
func (n *GreaterThanZeroFloat) Applications() []*Application { return []*Application{} }

func (n *IfEqual) Applications() []*Application {
	return append(n.True.Applications(), n.False.Applications()...)
}

func (n *IfEqualZero) Applications() []*Application {
	return append(n.True.Applications(), n.False.Applications()...)
}

func (n *IfEqualTrue) Applications() []*Application {
	return append(n.True.Applications(), n.False.Applications()...)
}

func (n *IfLessThan) Applications() []*Application {
	return append(n.True.Applications(), n.False.Applications()...)
}

func (n *IfLessThanFloat) Applications() []*Application {
	return append(n.True.Applications(), n.False.Applications()...)
}

func (n *IfLessThanZero) Applications() []*Application {
	return append(n.True.Applications(), n.False.Applications()...)
}

func (n *IfLessThanZeroFloat) Applications() []*Application {
	return append(n.True.Applications(), n.False.Applications()...)
}

func (n *Assignment) Applications() []*Application {
	return append(n.Value.Applications(), n.Next.Applications()...)
}

func (n *Application) Applications() []*Application {
	return []*Application{n}
}

func (n *Tuple) Applications() []*Application                { return []*Application{} }
func (n *TupleGet) Applications() []*Application             { return []*Application{} }
func (n *ArrayCreate) Applications() []*Application          { return []*Application{} }
func (n *ArrayCreateImmediate) Applications() []*Application { return []*Application{} }
func (n *ArrayGet) Applications() []*Application             { return []*Application{} }
func (n *ArrayGetImmediate) Applications() []*Application    { return []*Application{} }
func (n *ArrayPut) Applications() []*Application             { return []*Application{} }
func (n *ArrayPutImmediate) Applications() []*Application    { return []*Application{} }
func (n *ReadInt) Applications() []*Application              { return []*Application{} }
func (n *ReadFloat) Applications() []*Application            { return []*Application{} }
func (n *WriteByte) Applications() []*Application            { return []*Application{} }
func (n *IntToFloat) Applications() []*Application           { return []*Application{} }
func (n *FloatToInt) Applications() []*Application           { return []*Application{} }
func (n *Sqrt) Applications() []*Application                 { return []*Application{} }

func (n *Variable) Size() int             { return 1 }
func (n *Unit) Size() int                 { return 1 }
func (n *Int) Size() int                  { return 1 }
func (n *Bool) Size() int                 { return 1 }
func (n *Float) Size() int                { return 1 }
func (n *Add) Size() int                  { return 1 }
func (n *AddImmediate) Size() int         { return 1 }
func (n *Sub) Size() int                  { return 1 }
func (n *SubFromZero) Size() int          { return 1 }
func (n *FloatAdd) Size() int             { return 1 }
func (n *FloatSub) Size() int             { return 1 }
func (n *FloatSubFromZero) Size() int     { return 1 }
func (n *FloatDiv) Size() int             { return 1 }
func (n *FloatMul) Size() int             { return 1 }
func (n *Not) Size() int                  { return 1 }
func (n *Equal) Size() int                { return 1 }
func (n *EqualZero) Size() int            { return 1 }
func (n *LessThan) Size() int             { return 1 }
func (n *LessThanFloat) Size() int        { return 1 }
func (n *LessThanZero) Size() int         { return 1 }
func (n *LessThanZeroFloat) Size() int    { return 1 }
func (n *GreaterThanZero) Size() int      { return 1 }
func (n *GreaterThanZeroFloat) Size() int { return 1 }
func (n *IfEqual) Size() int              { return n.True.Size() + n.False.Size() }
func (n *IfEqualZero) Size() int          { return n.True.Size() + n.False.Size() }
func (n *IfEqualTrue) Size() int          { return n.True.Size() + n.False.Size() }
func (n *IfLessThan) Size() int           { return n.True.Size() + n.False.Size() }
func (n *IfLessThanFloat) Size() int      { return n.True.Size() + n.False.Size() }
func (n *IfLessThanZero) Size() int       { return n.True.Size() + n.False.Size() }
func (n *IfLessThanZeroFloat) Size() int  { return n.True.Size() + n.False.Size() }
func (n *Assignment) Size() int           { return n.Value.Size() + n.Next.Size() }
func (n *Application) Size() int          { return 1 }
func (n *Tuple) Size() int                { return 1 }
func (n *TupleGet) Size() int             { return 1 }
func (n *ArrayCreate) Size() int          { return 1 }
func (n *ArrayCreateImmediate) Size() int { return 1 }
func (n *ArrayGet) Size() int             { return 1 }
func (n *ArrayGetImmediate) Size() int    { return 1 }
func (n *ArrayPut) Size() int             { return 1 }
func (n *ArrayPutImmediate) Size() int    { return 1 }
func (n *ReadInt) Size() int              { return 1 }
func (n *ReadFloat) Size() int            { return 1 }
func (n *WriteByte) Size() int            { return 1 }
func (n *IntToFloat) Size() int           { return 1 }
func (n *FloatToInt) Size() int           { return 1 }
func (n *Sqrt) Size() int                 { return 1 }

func (n *Variable) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return values[n.Name]
}

func (n *Unit) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return nil
}

func (n *Int) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return n.Value
}

func (n *Float) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return n.Value
}

func (n *Bool) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return n.Value
}

func (n *Add) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if left, ok := values[n.Left].(int32); ok {
		if right, ok := values[n.Right].(int32); ok {
			return left + right
		}
	}

	return nil
}

func (n *AddImmediate) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if left, ok := values[n.Left].(int32); ok {
		return left + n.Right
	}

	return nil
}

func (n *Sub) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if left, ok := values[n.Left].(int32); ok {
		if right, ok := values[n.Right].(int32); ok {
			return left - right
		}
	}

	return nil
}

func (n *SubFromZero) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if inner, ok := values[n.Inner].(int32); ok {
		return -inner
	}

	return nil
}

func (n *FloatAdd) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if left, ok := values[n.Left].(float32); ok {
		if right, ok := values[n.Right].(float32); ok {
			return left + right
		}
	}

	return nil
}

func (n *FloatSub) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if left, ok := values[n.Left].(float32); ok {
		if right, ok := values[n.Right].(float32); ok {
			return left - right
		}
	}

	return nil
}

func (n *FloatSubFromZero) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if inner, ok := values[n.Inner].(float32); ok {
		return -inner
	}

	return nil
}

func (n *FloatDiv) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if left, ok := values[n.Left].(float32); ok {
		if right, ok := values[n.Right].(float32); ok {
			return left / right
		}
	}

	return nil
}

func (n *FloatMul) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if left, ok := values[n.Left].(float32); ok {
		if right, ok := values[n.Right].(float32); ok {
			return left * right
		}
	}

	return nil
}

func (n *Not) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if inner, ok := values[n.Inner].(bool); ok {
		return !inner
	}

	return nil
}

func (n *Equal) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if left, ok := values[n.Left].(int32); ok {
		if right, ok := values[n.Right].(int32); ok {
			return left == right
		}
	}

	if left, ok := values[n.Left].(float32); ok {
		if right, ok := values[n.Right].(float32); ok {
			return left == right
		}
	}

	if left, ok := values[n.Left].(bool); ok {
		if right, ok := values[n.Right].(bool); ok {
			return left == right
		}
	}

	return nil
}

func (n *EqualZero) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if inner, ok := values[n.Inner].(int32); ok {
		return inner == 0
	}

	if inner, ok := values[n.Inner].(float32); ok {
		return inner == 0
	}

	return nil
}

func (n *LessThan) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if left, ok := values[n.Left].(int32); ok {
		if right, ok := values[n.Right].(int32); ok {
			return left < right
		}
	}

	return nil
}

func (n *LessThanFloat) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if left, ok := values[n.Left].(float32); ok {
		if right, ok := values[n.Right].(float32); ok {
			return left < right
		}
	}

	return nil
}

func (n *LessThanZero) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if inner, ok := values[n.Inner].(int32); ok {
		return inner < 0
	}

	return nil
}

func (n *LessThanZeroFloat) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if inner, ok := values[n.Inner].(float32); ok {
		return inner < 0
	}

	return nil
}

func (n *GreaterThanZero) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if inner, ok := values[n.Inner].(int32); ok {
		return inner > 0
	}

	return nil
}

func (n *GreaterThanZeroFloat) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if inner, ok := values[n.Inner].(float32); ok {
		return inner > 0
	}

	return nil
}

func (n *IfEqual) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if left, ok := values[n.Left].(int32); ok {
		if right, ok := values[n.Right].(int32); ok {
			if left == right {
				return n.True.Evaluate(values, functions)
			} else {
				return n.False.Evaluate(values, functions)
			}
		}
	}

	if left, ok := values[n.Left].(float32); ok {
		if right, ok := values[n.Right].(float32); ok {
			if left == right {
				return n.True.Evaluate(values, functions)
			} else {
				return n.False.Evaluate(values, functions)
			}
		}
	}

	if left, ok := values[n.Left].(bool); ok {
		if right, ok := values[n.Right].(bool); ok {
			if left == right {
				return n.True.Evaluate(values, functions)
			} else {
				return n.False.Evaluate(values, functions)
			}
		}
	}

	return nil
}

func (n *IfEqualZero) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if inner, ok := values[n.Inner].(int32); ok {
		if inner == 0 {
			return n.True.Evaluate(values, functions)
		} else {
			return n.False.Evaluate(values, functions)
		}
	}

	if inner, ok := values[n.Inner].(float32); ok {
		if inner == 0 {
			return n.True.Evaluate(values, functions)
		} else {
			return n.False.Evaluate(values, functions)
		}
	}

	return nil
}

func (n *IfEqualTrue) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if inner, ok := values[n.Inner].(bool); ok {
		if inner {
			return n.True.Evaluate(values, functions)
		} else {
			return n.False.Evaluate(values, functions)
		}
	}

	return nil
}

func (n *IfLessThan) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if left, ok := values[n.Left].(int32); ok {
		if right, ok := values[n.Right].(int32); ok {
			if left < right {
				return n.True.Evaluate(values, functions)
			} else {
				return n.False.Evaluate(values, functions)
			}
		}
	}

	return nil
}

func (n *IfLessThanFloat) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if left, ok := values[n.Left].(float32); ok {
		if right, ok := values[n.Right].(float32); ok {
			if left < right {
				return n.True.Evaluate(values, functions)
			} else {
				return n.False.Evaluate(values, functions)
			}
		}
	}

	return nil
}

func (n *IfLessThanZero) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if inner, ok := values[n.Inner].(int32); ok {
		if inner < 0 {
			return n.True.Evaluate(values, functions)
		} else {
			return n.False.Evaluate(values, functions)
		}
	}

	return nil
}

func (n *IfLessThanZeroFloat) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if inner, ok := values[n.Inner].(float32); ok {
		if inner < 0 {
			return n.True.Evaluate(values, functions)
		} else {
			return n.False.Evaluate(values, functions)
		}
	}

	return nil
}

func (n *Assignment) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	valuesExtended := map[string]interface{}{}
	for k, v := range values {
		valuesExtended[k] = v
	}

	valuesExtended[n.Name] = n.Value.Evaluate(values, functions)

	return n.Next.Evaluate(valuesExtended, functions)
}

func (n *Application) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	argValues := []interface{}{}
	for _, arg := range n.Args {
		v := values[arg]
		if v == nil {
			return nil
		}

		argValues = append(argValues, v)
	}

	for _, function := range functions {
		if function.Name == n.Function {
			values := map[string]interface{}{}
			for i, arg := range function.Args {
				values[arg] = argValues[i]
			}
			return function.Body.Evaluate(values, functions)
		}
	}

	return nil
}

func (n *Tuple) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	tuple := []interface{}{}
	for _, element := range n.Elements {
		tuple = append(tuple, values[element])
	}
	return tuple
}

func (n *TupleGet) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if tuple, ok := values[n.Tuple].([]interface{}); ok {
		return tuple[n.Index]
	}

	return nil
}

func (n *ArrayCreate) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return nil
}

func (n *ArrayCreateImmediate) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return nil
}

func (n *ArrayGet) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return nil
}

func (n *ArrayGetImmediate) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return nil
}

func (n *ArrayPut) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return nil
}

func (n *ArrayPutImmediate) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return nil
}

func (n *ReadInt) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return nil
}

func (n *ReadFloat) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return nil
}

func (n *WriteByte) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return nil
}

func (n *IntToFloat) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return nil
}

func (n *FloatToInt) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	return nil
}

func (n *Sqrt) Evaluate(values map[string]interface{}, functions []*Function) interface{} {
	if arg, ok := values[n.Arg].(float32); ok {
		return float32(math.Sqrt(float64(arg)))
	}

	return nil
}
