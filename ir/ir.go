package ir

type Function struct {
	Name string
	Args []string
	Body Node
}

func (f Function) FreeVariables() map[string]struct{} {
	bound := map[string]struct{}{}

	bound[f.Name] = struct{}{}

	for _, arg := range f.Args {
		bound[arg] = struct{}{}
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
	UpdateNames(mapping map[string]string)
	FreeVariables(bound map[string]struct{}) map[string]struct{}
	FloatValues() []float32
	Clone() Node
	HasSideEffects() bool
	Applications() []*Application
	irNode()
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

type IfEqual struct {
	Left, Right string
	True, False Node
}

type IfEqualZero struct {
	Inner       string
	True, False Node
}

type IfLessThan struct {
	Left, Right string
	True, False Node
}

type IfLessThanZero struct {
	Inner       string
	True, False Node
}

type ValueBinding struct {
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

func (n *Variable) irNode()         {}
func (n *Unit) irNode()             {}
func (n *Int) irNode()              {}
func (n *Bool) irNode()             {}
func (n *Float) irNode()            {}
func (n *Add) irNode()              {}
func (n *AddImmediate) irNode()     {}
func (n *Sub) irNode()              {}
func (n *SubFromZero) irNode()      {}
func (n *FloatAdd) irNode()         {}
func (n *FloatSub) irNode()         {}
func (n *FloatSubFromZero) irNode() {}
func (n *FloatDiv) irNode()         {}
func (n *FloatMul) irNode()         {}
func (n *IfEqual) irNode()          {}
func (n *IfEqualZero) irNode()      {}
func (n *IfLessThan) irNode()       {}
func (n *IfLessThanZero) irNode()   {}
func (n *ValueBinding) irNode()     {}
func (n *Application) irNode()      {}
func (n *Tuple) irNode()            {}
func (n *ArrayCreate) irNode()      {}
func (n *ArrayGet) irNode()         {}
func (n *ArrayPut) irNode()         {}
func (n *ReadInt) irNode()          {}
func (n *ReadFloat) irNode()        {}
func (n *PrintInt) irNode()         {}
func (n *PrintChar) irNode()        {}
func (n *IntToFloat) irNode()       {}
func (n *FloatToInt) irNode()       {}
func (n *Sqrt) irNode()             {}
func (n *TupleGet) irNode()         {}

func replaceIfFound(k string, m map[string]string) string {
	if v, ok := m[k]; ok {
		return v
	}

	return k
}

func (n *Variable) UpdateNames(mapping map[string]string) {
	n.Name = replaceIfFound(n.Name, mapping)
}

func (n *Unit) UpdateNames(mapping map[string]string)  {}
func (n *Int) UpdateNames(mapping map[string]string)   {}
func (n *Bool) UpdateNames(mapping map[string]string)  {}
func (n *Float) UpdateNames(mapping map[string]string) {}

func (n *Add) UpdateNames(mapping map[string]string) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *AddImmediate) UpdateNames(mapping map[string]string) {
	n.Left = replaceIfFound(n.Left, mapping)
}

func (n *Sub) UpdateNames(mapping map[string]string) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *SubFromZero) UpdateNames(mapping map[string]string) {
	n.Inner = replaceIfFound(n.Inner, mapping)
}

func (n *FloatAdd) UpdateNames(mapping map[string]string) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *FloatSub) UpdateNames(mapping map[string]string) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *FloatSubFromZero) UpdateNames(mapping map[string]string) {
	n.Inner = replaceIfFound(n.Inner, mapping)
}

func (n *FloatDiv) UpdateNames(mapping map[string]string) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *FloatMul) UpdateNames(mapping map[string]string) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
}

func (n *IfEqual) UpdateNames(mapping map[string]string) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
	n.True.UpdateNames(mapping)
	n.False.UpdateNames(mapping)
}

func (n *IfEqualZero) UpdateNames(mapping map[string]string) {
	n.Inner = replaceIfFound(n.Inner, mapping)
	n.True.UpdateNames(mapping)
	n.False.UpdateNames(mapping)
}

func (n *IfLessThan) UpdateNames(mapping map[string]string) {
	n.Left = replaceIfFound(n.Left, mapping)
	n.Right = replaceIfFound(n.Right, mapping)
	n.True.UpdateNames(mapping)
	n.False.UpdateNames(mapping)
}

func (n *IfLessThanZero) UpdateNames(mapping map[string]string) {
	n.Inner = replaceIfFound(n.Inner, mapping)
	n.True.UpdateNames(mapping)
	n.False.UpdateNames(mapping)
}

func (n *ValueBinding) UpdateNames(mapping map[string]string) {
	n.Name = replaceIfFound(n.Name, mapping)
	n.Value.UpdateNames(mapping)
	n.Next.UpdateNames(mapping)
}

func (n *Application) UpdateNames(mapping map[string]string) {
	for i := range n.Args {
		n.Args[i] = replaceIfFound(n.Args[i], mapping)
	}
}

func (n *Tuple) UpdateNames(mapping map[string]string) {
	for i := range n.Elements {
		n.Elements[i] = replaceIfFound(n.Elements[i], mapping)
	}
}

func (n *ArrayCreate) UpdateNames(mapping map[string]string) {
	n.Size = replaceIfFound(n.Size, mapping)
	n.Value = replaceIfFound(n.Value, mapping)
}

func (n *ArrayGet) UpdateNames(mapping map[string]string) {
	n.Array = replaceIfFound(n.Array, mapping)
	n.Index = replaceIfFound(n.Index, mapping)
}

func (n *ArrayPut) UpdateNames(mapping map[string]string) {
	n.Array = replaceIfFound(n.Array, mapping)
	n.Index = replaceIfFound(n.Index, mapping)
	n.Value = replaceIfFound(n.Value, mapping)
}

func (n *ReadInt) UpdateNames(mapping map[string]string)   {}
func (n *ReadFloat) UpdateNames(mapping map[string]string) {}

func (n *PrintInt) UpdateNames(mapping map[string]string) {
	n.Arg = replaceIfFound(n.Arg, mapping)
}

func (n *PrintChar) UpdateNames(mapping map[string]string) {
	n.Arg = replaceIfFound(n.Arg, mapping)
}
func (n *IntToFloat) UpdateNames(mapping map[string]string) {
	n.Arg = replaceIfFound(n.Arg, mapping)
}
func (n *FloatToInt) UpdateNames(mapping map[string]string) {
	n.Arg = replaceIfFound(n.Arg, mapping)
}
func (n *Sqrt) UpdateNames(mapping map[string]string) {
	n.Arg = replaceIfFound(n.Arg, mapping)
}

func (n *TupleGet) UpdateNames(mapping map[string]string) {
	n.Tuple = replaceIfFound(n.Tuple, mapping)
}

func (n *Variable) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	if _, ok := bound[n.Name]; !ok {
		return map[string]struct{}{n.Name: struct{}{}}
	}

	return map[string]struct{}{}
}

func (n *Unit) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	return map[string]struct{}{}
}

func (n *Int) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	return map[string]struct{}{}
}

func (n *Bool) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	return map[string]struct{}{}
}

func (n *Float) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	return map[string]struct{}{}
}

func (n *Add) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Left]; !ok {
		ret[n.Left] = struct{}{}
	}
	if _, ok := bound[n.Right]; !ok {
		ret[n.Right] = struct{}{}
	}
	return ret
}

func (n *AddImmediate) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Left]; !ok {
		ret[n.Left] = struct{}{}
	}
	return ret
}

func (n *Sub) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Left]; !ok {
		ret[n.Left] = struct{}{}
	}
	if _, ok := bound[n.Right]; !ok {
		ret[n.Right] = struct{}{}
	}
	return ret
}

func (n *SubFromZero) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Inner]; !ok {
		ret[n.Inner] = struct{}{}
	}
	return ret
}

func (n *FloatAdd) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Left]; !ok {
		ret[n.Left] = struct{}{}
	}
	if _, ok := bound[n.Right]; !ok {
		ret[n.Right] = struct{}{}
	}
	return ret
}

func (n *FloatSub) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Left]; !ok {
		ret[n.Left] = struct{}{}
	}
	if _, ok := bound[n.Right]; !ok {
		ret[n.Right] = struct{}{}
	}
	return ret
}

func (n *FloatSubFromZero) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Inner]; !ok {
		ret[n.Inner] = struct{}{}
	}
	return ret
}

func (n *FloatDiv) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Left]; !ok {
		ret[n.Left] = struct{}{}
	}
	if _, ok := bound[n.Right]; !ok {
		ret[n.Right] = struct{}{}
	}
	return ret
}

func (n *FloatMul) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Left]; !ok {
		ret[n.Left] = struct{}{}
	}
	if _, ok := bound[n.Right]; !ok {
		ret[n.Right] = struct{}{}
	}
	return ret
}

func (n *IfEqual) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Left]; !ok {
		ret[n.Left] = struct{}{}
	}
	if _, ok := bound[n.Right]; !ok {
		ret[n.Right] = struct{}{}
	}
	for v := range n.True.FreeVariables(bound) {
		ret[v] = struct{}{}
	}
	for v := range n.False.FreeVariables(bound) {
		ret[v] = struct{}{}
	}
	return ret
}

func (n *IfEqualZero) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Inner]; !ok {
		ret[n.Inner] = struct{}{}
	}
	for v := range n.True.FreeVariables(bound) {
		ret[v] = struct{}{}
	}
	for v := range n.False.FreeVariables(bound) {
		ret[v] = struct{}{}
	}
	return ret
}

func (n *IfLessThan) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Left]; !ok {
		ret[n.Left] = struct{}{}
	}
	if _, ok := bound[n.Right]; !ok {
		ret[n.Right] = struct{}{}
	}
	for v := range n.True.FreeVariables(bound) {
		ret[v] = struct{}{}
	}
	for v := range n.False.FreeVariables(bound) {
		ret[v] = struct{}{}
	}
	return ret
}

func (n *IfLessThanZero) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Inner]; !ok {
		ret[n.Inner] = struct{}{}
	}
	for v := range n.True.FreeVariables(bound) {
		ret[v] = struct{}{}
	}
	for v := range n.False.FreeVariables(bound) {
		ret[v] = struct{}{}
	}
	return ret
}

func (n *ValueBinding) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	for v := range n.Value.FreeVariables(bound) {
		ret[v] = struct{}{}
	}
	bound[n.Name] = struct{}{}
	for v := range n.Next.FreeVariables(bound) {
		ret[v] = struct{}{}
	}
	delete(bound, n.Name)
	return ret
}

func (n *Application) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	for _, arg := range n.Args {
		if _, ok := bound[arg]; !ok {
			ret[arg] = struct{}{}
		}
	}
	return ret
}

func (n *Tuple) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	for _, element := range n.Elements {
		if _, ok := bound[element]; !ok {
			ret[element] = struct{}{}
		}
	}
	return ret
}

func (n *ArrayCreate) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Size]; !ok {
		ret[n.Size] = struct{}{}
	}
	if _, ok := bound[n.Value]; !ok {
		ret[n.Value] = struct{}{}
	}
	return ret
}

func (n *ArrayGet) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Array]; !ok {
		ret[n.Array] = struct{}{}
	}
	if _, ok := bound[n.Index]; !ok {
		ret[n.Index] = struct{}{}
	}
	return ret
}

func (n *ArrayPut) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Array]; !ok {
		ret[n.Array] = struct{}{}
	}
	if _, ok := bound[n.Index]; !ok {
		ret[n.Index] = struct{}{}
	}
	if _, ok := bound[n.Value]; !ok {
		ret[n.Value] = struct{}{}
	}
	return ret
}

func (n *ReadInt) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	return map[string]struct{}{}
}

func (n *ReadFloat) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	return map[string]struct{}{}
}

func (n *PrintInt) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Arg]; !ok {
		ret[n.Arg] = struct{}{}
	}
	return ret
}

func (n *PrintChar) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Arg]; !ok {
		ret[n.Arg] = struct{}{}
	}
	return ret
}

func (n *IntToFloat) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Arg]; !ok {
		ret[n.Arg] = struct{}{}
	}
	return ret
}

func (n *FloatToInt) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Arg]; !ok {
		ret[n.Arg] = struct{}{}
	}
	return ret
}

func (n *Sqrt) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Arg]; !ok {
		ret[n.Arg] = struct{}{}
	}
	return ret
}

func (n *TupleGet) FreeVariables(bound map[string]struct{}) map[string]struct{} {
	ret := map[string]struct{}{}
	if _, ok := bound[n.Tuple]; !ok {
		ret[n.Tuple] = struct{}{}
	}
	return ret
}

func (n *Variable) FloatValues() []float32         { return []float32{} }
func (n *Unit) FloatValues() []float32             { return []float32{} }
func (n *Int) FloatValues() []float32              { return []float32{} }
func (n *Bool) FloatValues() []float32             { return []float32{} }
func (n *Float) FloatValues() []float32            { return []float32{n.Value} }
func (n *Add) FloatValues() []float32              { return []float32{} }
func (n *AddImmediate) FloatValues() []float32     { return []float32{} }
func (n *Sub) FloatValues() []float32              { return []float32{} }
func (n *SubFromZero) FloatValues() []float32      { return []float32{} }
func (n *FloatAdd) FloatValues() []float32         { return []float32{} }
func (n *FloatSub) FloatValues() []float32         { return []float32{} }
func (n *FloatSubFromZero) FloatValues() []float32 { return []float32{} }
func (n *FloatDiv) FloatValues() []float32         { return []float32{} }
func (n *FloatMul) FloatValues() []float32         { return []float32{} }

func (n *IfEqual) FloatValues() []float32 {
	return append(n.True.FloatValues(), n.False.FloatValues()...)
}

func (n *IfEqualZero) FloatValues() []float32 {
	return append(n.True.FloatValues(), n.False.FloatValues()...)
}

func (n *IfLessThan) FloatValues() []float32 {
	return append(n.True.FloatValues(), n.False.FloatValues()...)
}

func (n *IfLessThanZero) FloatValues() []float32 {
	return append(n.True.FloatValues(), n.False.FloatValues()...)
}

func (n *ValueBinding) FloatValues() []float32 {
	return append(n.Value.FloatValues(), n.Next.FloatValues()...)
}

func (n *Application) FloatValues() []float32 { return []float32{} }
func (n *Tuple) FloatValues() []float32       { return []float32{} }
func (n *TupleGet) FloatValues() []float32    { return []float32{} }
func (n *ArrayCreate) FloatValues() []float32 { return []float32{} }
func (n *ArrayGet) FloatValues() []float32    { return []float32{} }
func (n *ArrayPut) FloatValues() []float32    { return []float32{} }
func (n *ReadInt) FloatValues() []float32     { return []float32{} }
func (n *ReadFloat) FloatValues() []float32   { return []float32{} }
func (n *PrintInt) FloatValues() []float32    { return []float32{} }
func (n *PrintChar) FloatValues() []float32   { return []float32{} }
func (n *IntToFloat) FloatValues() []float32  { return []float32{} }
func (n *FloatToInt) FloatValues() []float32  { return []float32{} }
func (n *Sqrt) FloatValues() []float32        { return []float32{} }

func (n *Variable) Clone() Node         { return &Variable{n.Name} }
func (n *Unit) Clone() Node             { return &Unit{} }
func (n *Int) Clone() Node              { return &Int{n.Value} }
func (n *Bool) Clone() Node             { return &Bool{n.Value} }
func (n *Float) Clone() Node            { return &Float{n.Value} }
func (n *Add) Clone() Node              { return &Add{n.Left, n.Right} }
func (n *AddImmediate) Clone() Node     { return &AddImmediate{n.Left, n.Right} }
func (n *Sub) Clone() Node              { return &Sub{n.Left, n.Right} }
func (n *SubFromZero) Clone() Node      { return &SubFromZero{n.Inner} }
func (n *FloatAdd) Clone() Node         { return &FloatAdd{n.Left, n.Right} }
func (n *FloatSub) Clone() Node         { return &FloatSub{n.Left, n.Right} }
func (n *FloatSubFromZero) Clone() Node { return &FloatSubFromZero{n.Inner} }
func (n *FloatDiv) Clone() Node         { return &FloatDiv{n.Left, n.Right} }
func (n *FloatMul) Clone() Node         { return &FloatMul{n.Left, n.Right} }

func (n *IfEqual) Clone() Node {
	return &IfEqual{n.Left, n.Right, n.True.Clone(), n.False.Clone()}
}

func (n *IfEqualZero) Clone() Node {
	return &IfEqualZero{n.Inner, n.True.Clone(), n.False.Clone()}
}

func (n *IfLessThan) Clone() Node {
	return &IfLessThan{n.Left, n.Right, n.True.Clone(), n.False.Clone()}
}

func (n *IfLessThanZero) Clone() Node {
	return &IfLessThanZero{n.Inner, n.True.Clone(), n.False.Clone()}
}

func (n *ValueBinding) Clone() Node {
	return &ValueBinding{n.Name, n.Value.Clone(), n.Next.Clone()}
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
	return &ArrayCreate{n.Size, n.Value}
}

func (n *ArrayGet) Clone() Node {
	return &ArrayGet{n.Array, n.Index}
}

func (n *ArrayPut) Clone() Node {
	return &ArrayPut{n.Array, n.Index, n.Value}
}

func (n *ReadInt) Clone() Node    { return &ReadInt{} }
func (n *ReadFloat) Clone() Node  { return &ReadFloat{} }
func (n *PrintInt) Clone() Node   { return &PrintInt{n.Arg} }
func (n *PrintChar) Clone() Node  { return &PrintChar{n.Arg} }
func (n *IntToFloat) Clone() Node { return &IntToFloat{n.Arg} }
func (n *FloatToInt) Clone() Node { return &FloatToInt{n.Arg} }
func (n *Sqrt) Clone() Node       { return &Sqrt{n.Arg} }

func (n *Variable) HasSideEffects() bool         { return false }
func (n *Unit) HasSideEffects() bool             { return false }
func (n *Int) HasSideEffects() bool              { return false }
func (n *Bool) HasSideEffects() bool             { return false }
func (n *Float) HasSideEffects() bool            { return false }
func (n *Add) HasSideEffects() bool              { return false }
func (n *AddImmediate) HasSideEffects() bool     { return false }
func (n *Sub) HasSideEffects() bool              { return false }
func (n *SubFromZero) HasSideEffects() bool      { return false }
func (n *FloatAdd) HasSideEffects() bool         { return false }
func (n *FloatSub) HasSideEffects() bool         { return false }
func (n *FloatSubFromZero) HasSideEffects() bool { return false }
func (n *FloatDiv) HasSideEffects() bool         { return false }
func (n *FloatMul) HasSideEffects() bool         { return false }

func (n *IfEqual) HasSideEffects() bool {
	return n.True.HasSideEffects() || n.False.HasSideEffects()
}

func (n *IfEqualZero) HasSideEffects() bool {
	return n.True.HasSideEffects() || n.False.HasSideEffects()
}

func (n *IfLessThan) HasSideEffects() bool {
	return n.True.HasSideEffects() || n.False.HasSideEffects()
}

func (n *IfLessThanZero) HasSideEffects() bool {
	return n.True.HasSideEffects() || n.False.HasSideEffects()
}

func (n *ValueBinding) HasSideEffects() bool {
	return n.Value.HasSideEffects() || n.Next.HasSideEffects()
}

func (n *Application) HasSideEffects() bool { return true }
func (n *Tuple) HasSideEffects() bool       { return false }
func (n *TupleGet) HasSideEffects() bool    { return false }
func (n *ArrayCreate) HasSideEffects() bool { return false }
func (n *ArrayGet) HasSideEffects() bool    { return false }
func (n *ArrayPut) HasSideEffects() bool    { return true }
func (n *ReadInt) HasSideEffects() bool     { return true }
func (n *ReadFloat) HasSideEffects() bool   { return true }
func (n *PrintInt) HasSideEffects() bool    { return true }
func (n *PrintChar) HasSideEffects() bool   { return true }
func (n *IntToFloat) HasSideEffects() bool  { return false }
func (n *FloatToInt) HasSideEffects() bool  { return false }
func (n *Sqrt) HasSideEffects() bool        { return false }

func (n *Variable) Applications() []*Application         { return []*Application{} }
func (n *Unit) Applications() []*Application             { return []*Application{} }
func (n *Int) Applications() []*Application              { return []*Application{} }
func (n *Bool) Applications() []*Application             { return []*Application{} }
func (n *Float) Applications() []*Application            { return []*Application{} }
func (n *Add) Applications() []*Application              { return []*Application{} }
func (n *AddImmediate) Applications() []*Application     { return []*Application{} }
func (n *Sub) Applications() []*Application              { return []*Application{} }
func (n *SubFromZero) Applications() []*Application      { return []*Application{} }
func (n *FloatAdd) Applications() []*Application         { return []*Application{} }
func (n *FloatSub) Applications() []*Application         { return []*Application{} }
func (n *FloatSubFromZero) Applications() []*Application { return []*Application{} }
func (n *FloatDiv) Applications() []*Application         { return []*Application{} }
func (n *FloatMul) Applications() []*Application         { return []*Application{} }

func (n *IfEqual) Applications() []*Application {
	return append(n.True.Applications(), n.False.Applications()...)
}

func (n *IfEqualZero) Applications() []*Application {
	return append(n.True.Applications(), n.False.Applications()...)
}

func (n *IfLessThan) Applications() []*Application {
	return append(n.True.Applications(), n.False.Applications()...)
}

func (n *IfLessThanZero) Applications() []*Application {
	return append(n.True.Applications(), n.False.Applications()...)
}

func (n *ValueBinding) Applications() []*Application {
	return append(n.Value.Applications(), n.Next.Applications()...)
}

func (n *Application) Applications() []*Application {
	return []*Application{n}
}

func (n *Tuple) Applications() []*Application       { return []*Application{} }
func (n *TupleGet) Applications() []*Application    { return []*Application{} }
func (n *ArrayCreate) Applications() []*Application { return []*Application{} }
func (n *ArrayGet) Applications() []*Application    { return []*Application{} }
func (n *ArrayPut) Applications() []*Application    { return []*Application{} }
func (n *ReadInt) Applications() []*Application     { return []*Application{} }
func (n *ReadFloat) Applications() []*Application   { return []*Application{} }
func (n *PrintInt) Applications() []*Application    { return []*Application{} }
func (n *PrintChar) Applications() []*Application   { return []*Application{} }
func (n *IntToFloat) Applications() []*Application  { return []*Application{} }
func (n *FloatToInt) Applications() []*Application  { return []*Application{} }
func (n *Sqrt) Applications() []*Application        { return []*Application{} }
