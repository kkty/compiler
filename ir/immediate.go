package ir

import (
	"math"
)

func Immediate(main Node, functions []*Function) Node {
	// Updates a node to use immediate values, and evaluates the value of each node at the
	// time. nil is used for unknown values.

	var updateAndEvaluate func(node Node, values map[string]interface{}) (Node, interface{})
	updateAndEvaluate = func(node Node, values map[string]interface{}) (Node, interface{}) {
		copyValues := func() map[string]interface{} {
			copied := map[string]interface{}{}
			for k, v := range values {
				copied[k] = v
			}
			return copied
		}

		switch node.(type) {
		case *Variable:
			n := node.(*Variable)

			if v, ok := values[n.Name].(int32); ok {
				return &Int{v}, v
			}

			if v, ok := values[n.Name].(float32); ok {
				return &Float{v}, v
			}

			return n, nil
		case *Int:
			n := node.(*Int)
			return n, n.Value
		case *Bool:
			n := node.(*Bool)
			return n, n.Value
		case *Float:
			n := node.(*Float)
			return n, n.Value
		case *Add:
			n := node.(*Add)

			if left, ok := values[n.Left].(int32); ok {
				if right, ok := values[n.Right].(int32); ok {
					return &Int{left + right}, left + right
				}

				return &AddImmediate{n.Right, left}, nil
			}

			if right, ok := values[n.Right].(int32); ok {
				return &AddImmediate{n.Left, right}, nil
			}

			return n, nil
		case *AddImmediate:
			n := node.(*AddImmediate)

			if left, ok := values[n.Left].(int32); ok {
				return &Int{left + n.Right}, left + n.Right
			}

			return n, nil
		case *Sub:
			n := node.(*Sub)

			if left, ok := values[n.Left].(int32); ok {
				if right, ok := values[n.Right].(int32); ok {
					return &Int{left - right}, left - right
				}

				if left == 0 {
					return &SubFromZero{n.Right}, nil
				}
			}

			if right, ok := values[n.Right].(int32); ok {
				return &AddImmediate{n.Left, -right}, nil
			}

			return n, nil
		case *SubFromZero:
			n := node.(*SubFromZero)

			if inner, ok := values[n.Inner].(int32); ok {
				return &Int{-inner}, -inner
			}

			return n, nil
		case *FloatAdd:
			n := node.(*FloatAdd)

			if left, ok := values[n.Left].(float32); ok {
				if right, ok := values[n.Right].(float32); ok {
					return &Float{left + right}, left + right
				}
			}

			return n, nil
		case *FloatSub:
			n := node.(*FloatSub)

			if left, ok := values[n.Left].(float32); ok {
				if right, ok := values[n.Right].(float32); ok {
					return &Float{left - right}, left - right
				}

				if left == 0 {
					return &FloatSubFromZero{n.Right}, nil
				}
			}

			return n, nil
		case *FloatDiv:
			n := node.(*FloatDiv)

			if left, ok := values[n.Left].(float32); ok {
				if right, ok := values[n.Right].(float32); ok {
					return &Float{left / right}, left / right
				}
			}

			return n, nil
		case *FloatMul:
			n := node.(*FloatMul)

			if left, ok := values[n.Left].(float32); ok {
				if right, ok := values[n.Right].(float32); ok {
					return &Float{left * right}, left * right
				}
			}

			return n, nil
		case *Not:
			n := node.(*Not)

			if inner, ok := values[n.Inner].(bool); ok {
				return &Bool{!inner}, !inner
			}

			return n, nil
		case *Equal:
			n := node.(*Equal)

			if left, ok := values[n.Left].(int32); ok {
				if right, ok := values[n.Right].(int32); ok {
					v := left == right
					return &Bool{v}, v
				}
			}

			if left, ok := values[n.Left].(bool); ok {
				if right, ok := values[n.Right].(bool); ok {
					v := left == right
					return &Bool{v}, v
				}
			}

			if left, ok := values[n.Left].(float32); ok {
				if right, ok := values[n.Right].(float32); ok {
					v := left == right
					return &Bool{v}, v
				}
			}

			return n, nil
		case *LessThan:
			n := node.(*LessThan)

			if left, ok := values[n.Left].(int32); ok {
				if right, ok := values[n.Right].(int32); ok {
					v := left < right
					return &Bool{v}, v
				}
			}

			if left, ok := values[n.Left].(float32); ok {
				if right, ok := values[n.Right].(float32); ok {
					v := left < right
					return &Bool{v}, v
				}
			}

			return n, nil
		case *IfEqual:
			n := node.(*IfEqual)

			if left, ok := values[n.Left].(int32); ok {
				if right, ok := values[n.Right].(int32); ok {
					if left == right {
						return updateAndEvaluate(n.True, values)
					} else {
						return updateAndEvaluate(n.False, values)
					}
				} else {
					if left == 0 {
						return &IfEqualZero{n.Right, n.True, n.False}, nil
					}
				}
			} else {
				if right, ok := values[n.Right].(int32); ok {
					if right == 0 {
						return &IfEqualZero{n.Left, n.True, n.False}, nil
					}
				}
			}

			if left, ok := values[n.Left].(float32); ok {
				if right, ok := values[n.Right].(float32); ok {
					if left == right {
						return updateAndEvaluate(n.True, values)
					} else {
						return updateAndEvaluate(n.False, values)
					}
				} else {
					if left == 0 {
						return &IfEqualZero{n.Right, n.True, n.False}, nil
					}
				}
			} else {
				if right, ok := values[n.Right].(float32); ok {
					if right == 0 {
						return &IfEqualZero{n.Left, n.True, n.False}, nil
					}
				}
			}

			if left, ok := values[n.Left].(bool); ok {
				if right, ok := values[n.Right].(bool); ok {
					if left == right {
						return updateAndEvaluate(n.True, values)
					} else {
						return updateAndEvaluate(n.False, values)
					}
				} else {
					if left {
						return &IfEqualTrue{n.Right, n.True, n.False}, nil
					} else {
						return &IfEqualTrue{n.Right, n.False, n.True}, nil
					}
				}
			} else {
				if right, ok := values[n.Right].(bool); ok {
					if right {
						return &IfEqualTrue{n.Left, n.True, n.False}, nil
					} else {
						return &IfEqualTrue{n.Left, n.False, n.True}, nil
					}
				}
			}

			n.True, _ = updateAndEvaluate(n.True, values)
			n.False, _ = updateAndEvaluate(n.False, values)

			return n, nil
		case *IfEqualZero:
			n := node.(*IfEqualZero)

			if inner, ok := values[n.Inner].(int32); ok {
				if inner == 0 {
					return updateAndEvaluate(n.True, values)
				} else {
					return updateAndEvaluate(n.False, values)
				}
			}

			if inner, ok := values[n.Inner].(float32); ok {
				if inner == 0 {
					return updateAndEvaluate(n.True, values)
				} else {
					return updateAndEvaluate(n.False, values)
				}
			}

			n.True, _ = updateAndEvaluate(n.True, values)
			n.False, _ = updateAndEvaluate(n.False, values)

			return n, nil
		case *IfEqualTrue:
			n := node.(*IfEqualTrue)

			if inner, ok := values[n.Inner].(bool); ok {
				if inner {
					return updateAndEvaluate(n.True, values)
				} else {
					return updateAndEvaluate(n.False, values)
				}
			}

			n.True, _ = updateAndEvaluate(n.True, values)
			n.False, _ = updateAndEvaluate(n.False, values)

			return n, nil
		case *IfLessThan:
			n := node.(*IfLessThan)

			if left, ok := values[n.Left].(int32); ok {
				if right, ok := values[n.Right].(int32); ok {
					if left < right {
						return updateAndEvaluate(n.True, values)
					} else {
						return updateAndEvaluate(n.False, values)
					}
				}
			} else {
				if right, ok := values[n.Right].(int32); ok {
					if right == 0 {
						return &IfLessThanZero{n.Left, n.True, n.False}, nil
					}
				}
			}

			if left, ok := values[n.Left].(float32); ok {
				if right, ok := values[n.Right].(float32); ok {
					if left < right {
						return updateAndEvaluate(n.True, values)
					} else {
						return updateAndEvaluate(n.False, values)
					}
				}
			} else {
				if right, ok := values[n.Right].(float32); ok {
					if right == 0 {
						return &IfLessThanZero{n.Left, n.True, n.False}, nil
					}
				}
			}

			n.True, _ = updateAndEvaluate(n.True, values)
			n.False, _ = updateAndEvaluate(n.False, values)

			return n, nil
		case *IfLessThanZero:
			n := node.(*IfLessThanZero)

			n.True, _ = updateAndEvaluate(n.True, values)
			n.False, _ = updateAndEvaluate(n.False, values)

			return n, nil
		case *ValueBinding:
			n := node.(*ValueBinding)

			valuesExtended := copyValues()
			n.Value, valuesExtended[n.Name] = updateAndEvaluate(n.Value, values)

			var value interface{}
			n.Next, value = updateAndEvaluate(n.Next, valuesExtended)

			return n, value
		case *Application:
			n := node.(*Application)

			argValues := []interface{}{}
			for _, arg := range n.Args {
				argValues = append(argValues, values[arg])
			}

			return n, nil
		case *ArrayCreate:
			n := node.(*ArrayCreate)

			if length, ok := values[n.Length].(int32); ok {
				return &ArrayCreateImmediate{length, n.Value}, nil
			}

			return n, nil
		case *ArrayGet:
			n := node.(*ArrayGet)

			if index, ok := values[n.Index].(int32); ok {
				return &ArrayGetImmediate{n.Array, index}, nil
			}

			return n, nil
		case *ArrayPut:
			n := node.(*ArrayPut)

			if index, ok := values[n.Index].(int32); ok {
				return &ArrayPutImmediate{n.Array, index, n.Value}, nil
			}

			return n, nil
		case *Sqrt:
			n := node.(*Sqrt)

			if arg, ok := values[n.Arg].(float32); ok {
				v := float32(math.Sqrt(float64(arg)))
				return &Float{v}, v
			}

			return n, nil
		default:
			return node, nil
		}
	}

	main, _ = updateAndEvaluate(main, map[string]interface{}{})
	for _, function := range functions {
		function.Body, _ = updateAndEvaluate(function.Body, map[string]interface{}{})
	}

	return main
}
