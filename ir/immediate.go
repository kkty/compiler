package ir

import (
	"github.com/kkty/compiler/stringmap"
)

// Immediate applies immediate-value optimization.
func Immediate(main Node, functions []*Function) Node {
	functionsWithoutSideEffects := FunctionsWithoutSideEffects(functions)

	// Updates a node to use immediate values, and evaluates the value of each node at the
	// same time. nil is used for unknown values.
	var update func(node Node, values map[string]interface{}) Node
	update = func(node Node, values map[string]interface{}) Node {
		if !node.HasSideEffects(functionsWithoutSideEffects) {
			value := node.Evaluate(values, functions)
			if v, ok := value.(int32); ok {
				return &Int{v}
			}

			if v, ok := value.(float32); ok {
				return &Float{v}
			}

			if v, ok := value.(bool); ok {
				return &Bool{v}
			}
		}

		switch n := node.(type) {
		case *Add:
			if left, ok := values[n.Left].(int32); ok {
				return &AddImmediate{n.Right, left}
			}
			if right, ok := values[n.Right].(int32); ok {
				return &AddImmediate{n.Left, right}
			}
		case *Sub:
			if left, ok := values[n.Left].(int32); ok {
				if left == 0 {
					return &SubFromZero{n.Right}
				}
			}
		case *FloatSub:
			if left, ok := values[n.Left].(float32); ok {
				if left == 0 {
					return &FloatSubFromZero{n.Right}
				}
			}
		case *FloatMul:
			if left, ok := values[n.Left].(float32); ok {
				if left == -1 {
					return &FloatSubFromZero{n.Right}
				}
			}
			if right, ok := values[n.Right].(float32); ok {
				if right == -1 {
					return &FloatSubFromZero{n.Left}
				}
			}
		case *Equal:
			if left, ok := values[n.Left].(int32); ok && left == 0 {
				return &EqualZero{n.Right}
			} else if right, ok := values[n.Right].(int32); ok && right == 0 {
				return &EqualZero{n.Left}
			}

			if left, ok := values[n.Left].(float32); ok && left == 0 {
				return &EqualZero{n.Right}
			} else if right, ok := values[n.Right].(float32); ok && right == 0 {
				return &EqualZero{n.Left}
			}
		case *LessThan:
			if left, ok := values[n.Left].(int32); ok && left == 0 {
				return &GreaterThanZero{n.Right}
			} else if right, ok := values[n.Right].(int32); ok && right == 0 {
				return &LessThanZero{n.Left}
			}
		case *LessThanFloat:
			if left, ok := values[n.Left].(float32); ok && left == 0 {
				return &GreaterThanZeroFloat{n.Right}
			} else if right, ok := values[n.Right].(float32); ok && right == 0 {
				return &LessThanZeroFloat{n.Left}
			}
		case *IfEqual:
			if left, ok := values[n.Left].(int32); ok {
				if right, ok := values[n.Right].(int32); ok {
					if left == right {
						return n.True
					} else {
						return n.False
					}
				} else {
					if left == 0 {
						return &IfEqualZero{n.Right, n.True, n.False}
					}
				}
			} else {
				if right, ok := values[n.Right].(int32); ok {
					if right == 0 {
						return &IfEqualZero{n.Left, n.True, n.False}
					}
				}
			}

			if left, ok := values[n.Left].(float32); ok {
				if right, ok := values[n.Right].(float32); ok {
					if left == right {
						return n.True
					} else {
						return n.False
					}
				} else {
					if left == 0 {
						return &IfEqualZero{n.Right, n.True, n.False}
					}
				}
			} else {
				if right, ok := values[n.Right].(float32); ok {
					if right == 0 {
						return &IfEqualZero{n.Left, n.True, n.False}
					}
				}
			}

			if left, ok := values[n.Left].(bool); ok {
				if right, ok := values[n.Right].(bool); ok {
					if left == right {
						return n.True
					} else {
						return n.False
					}
				} else {
					if left {
						return &IfEqualTrue{n.Right, n.True, n.False}
					} else {
						return &IfEqualTrue{n.Right, n.False, n.True}
					}
				}
			} else {
				if right, ok := values[n.Right].(bool); ok {
					if right {
						return &IfEqualTrue{n.Left, n.True, n.False}
					} else {
						return &IfEqualTrue{n.Left, n.False, n.True}
					}
				}
			}

			n.True = update(n.True, values)
			n.False = update(n.False, values)
		case *IfEqualZero:
			if inner, ok := values[n.Inner].(int32); ok {
				if inner == 0 {
					return n.True
				} else {
					return n.False
				}
			}

			if inner, ok := values[n.Inner].(float32); ok {
				if inner == 0 {
					return n.True
				} else {
					return n.False
				}
			}

			n.True = update(n.True, values)
			n.False = update(n.False, values)
		case *IfEqualTrue:
			if inner, ok := values[n.Inner].(bool); ok {
				if inner {
					return n.True
				} else {
					return n.False
				}
			}

			n.True = update(n.True, values)
			n.False = update(n.False, values)
		case *IfLessThan:
			if left, ok := values[n.Left].(int32); ok {
				if right, ok := values[n.Right].(int32); ok {
					if left < right {
						return n.True
					} else {
						return n.False
					}
				}
			} else {
				if right, ok := values[n.Right].(int32); ok {
					if right == 0 {
						return &IfLessThanZero{n.Left, n.True, n.False}
					}
				}
			}

			n.True = update(n.True, values)
			n.False = update(n.False, values)
		case *IfLessThanFloat:
			if left, ok := values[n.Left].(float32); ok {
				if right, ok := values[n.Right].(float32); ok {
					if left < right {
						return n.True
					} else {
						return n.False
					}
				}
			} else {
				if right, ok := values[n.Right].(float32); ok {
					if right == 0 {
						return &IfLessThanZeroFloat{n.Left, n.True, n.False}
					}
				}
			}

			n.True = update(n.True, values)
			n.False = update(n.False, values)
		case *IfLessThanZero:
			n.True = update(n.True, values)
			n.False = update(n.False, values)
		case *IfLessThanZeroFloat:
			n.True = update(n.True, values)
			n.False = update(n.False, values)
		case *Assignment:
			n.Value = update(n.Value, values)

			switch value := n.Value.(type) {
			case *Int:
				for k, v := range values {
					if v == value.Value {
						n.Next.UpdateNames(stringmap.Map{n.Name: k})
						return update(n.Next, values)
					}
				}
			case *Float:
				for k, v := range values {
					if v == value.Value {
						n.Next.UpdateNames(stringmap.Map{n.Name: k})
						return update(n.Next, values)
					}
				}
			case *Bool:
				for k, v := range values {
					if v == value.Value {
						n.Next.UpdateNames(stringmap.Map{n.Name: k})
						return update(n.Next, values)
					}
				}
			}

			valuesExtended := map[string]interface{}{}
			for k, v := range values {
				valuesExtended[k] = v
			}
			valuesExtended[n.Name] = n.Value.Evaluate(values, functions)
			n.Next = update(n.Next, valuesExtended)
		case *ArrayCreate:
			if length, ok := values[n.Length].(int32); ok {
				return &ArrayCreateImmediate{length, n.Value}
			}
		case *ArrayGet:
			if index, ok := values[n.Index].(int32); ok {
				return &ArrayGetImmediate{n.Array, index}
			}
		case *ArrayPut:
			if index, ok := values[n.Index].(int32); ok {
				return &ArrayPutImmediate{n.Array, index, n.Value}
			}
		}

		return node
	}

	main = update(main, map[string]interface{}{})
	for _, function := range functions {
		function.Body = update(function.Body, map[string]interface{}{})
	}

	return main
}
