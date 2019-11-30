package ir

import (
	"math"

	"github.com/thoas/go-funk"
)

func Immediate(main Node, functions []*Function) Node {
	functionToArgValues := map[string][][]interface{}{}

	// Updates a node to use immediate values, and evaluates the value of each node at the
	// time. nil is used for unknown values.

	var updateAndEvaluate func(node Node, values map[string]interface{}) (Node, interface{})
	updateAndEvaluate = func(node Node, values map[string]interface{}) (Node, interface{}) {
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
		case *IfEqual:
			n := node.(*IfEqual)

			copiedValues := map[string]interface{}{}
			for k, v := range values {
				copiedValues[k] = v
			}

			if left, ok := values[n.Left].(int32); ok {
				if right, ok := values[n.Right].(int32); ok {
					if left == right {
						return updateAndEvaluate(n.True, copiedValues)
					} else {
						return updateAndEvaluate(n.False, copiedValues)
					}
				} else {
					if left == 0 {
						n := &IfEqualZero{n.Right, n.True, n.False}
						var value1, value2 interface{}
						n.True, value1 = updateAndEvaluate(n.True, copiedValues)
						n.False, value2 = updateAndEvaluate(n.False, copiedValues)
						if value1 == value2 {
							return n, value1
						}
						return n, nil
					}
				}
			} else {
				if right, ok := values[n.Right].(int32); ok {
					if right == 0 {
						n := &IfEqualZero{n.Left, n.True, n.False}
						var value1, value2 interface{}
						n.True, value1 = updateAndEvaluate(n.True, copiedValues)
						n.False, value2 = updateAndEvaluate(n.False, copiedValues)
						if value1 == value2 {
							return n, value1
						}
						return n, nil
					}
				}
			}

			if left, ok := values[n.Left].(float32); ok {
				if right, ok := values[n.Right].(float32); ok {
					if left == right {
						return updateAndEvaluate(n.True, copiedValues)
					} else {
						return updateAndEvaluate(n.False, copiedValues)
					}
				} else {
					if left == 0 {
						n := &IfEqualZero{n.Right, n.True, n.False}
						var value1, value2 interface{}
						n.True, value1 = updateAndEvaluate(n.True, copiedValues)
						n.False, value2 = updateAndEvaluate(n.False, copiedValues)
						if value1 == value2 {
							return n, value1
						}
						return n, nil
					}
				}
			} else {
				if right, ok := values[n.Right].(float32); ok {
					if right == 0 {
						n := &IfEqualZero{n.Left, n.True, n.False}
						var value1, value2 interface{}
						n.True, value1 = updateAndEvaluate(n.True, copiedValues)
						n.False, value2 = updateAndEvaluate(n.False, copiedValues)
						if value1 == value2 {
							return n, value1
						}
						return n, nil
					}
				}
			}

			var value1, value2 interface{}
			n.True, value1 = updateAndEvaluate(n.True, copiedValues)
			n.False, value2 = updateAndEvaluate(n.False, copiedValues)
			if value1 == value2 {
				return n, value1
			}

			return n, nil
		case *IfEqualZero:
			n := node.(*IfEqualZero)

			copiedValues := map[string]interface{}{}

			for k, v := range values {
				copiedValues[k] = v
			}

			n.True, _ = updateAndEvaluate(n.True, copiedValues)
			n.False, _ = updateAndEvaluate(n.False, copiedValues)

			return n, nil
		case *IfLessThan:
			n := node.(*IfLessThan)

			copiedValues := map[string]interface{}{}

			for k, v := range values {
				copiedValues[k] = v
			}

			if left, ok := values[n.Left].(int32); ok {
				if right, ok := values[n.Right].(int32); ok {
					if left < right {
						return updateAndEvaluate(n.True, copiedValues)
					} else {
						return updateAndEvaluate(n.False, copiedValues)
					}
				}
			} else {
				if right, ok := values[n.Right].(int32); ok {
					if right == 0 {
						n := &IfLessThanZero{n.Left, n.True, n.False}
						var value1, value2 interface{}
						n.True, value1 = updateAndEvaluate(n.True, copiedValues)
						n.False, value2 = updateAndEvaluate(n.False, copiedValues)
						if value1 == value2 {
							return n, value1
						}
						return n, nil
					}
				}
			}

			if left, ok := values[n.Left].(float32); ok {
				if right, ok := values[n.Right].(float32); ok {
					if left < right {
						return updateAndEvaluate(n.True, copiedValues)
					} else {
						return updateAndEvaluate(n.False, copiedValues)
					}
				}
			} else {
				if right, ok := values[n.Right].(float32); ok {
					if right == 0 {
						n := &IfLessThanZero{n.Left, n.True, n.False}
						var value1, value2 interface{}
						n.True, value1 = updateAndEvaluate(n.True, copiedValues)
						n.False, value2 = updateAndEvaluate(n.False, copiedValues)
						if value1 == value2 {
							return n, value1
						}
						return n, nil
					}
				}
			}

			var value1, value2 interface{}
			n.True, value1 = updateAndEvaluate(n.True, copiedValues)
			n.False, value2 = updateAndEvaluate(n.False, copiedValues)
			if value1 == value2 {
				return n, value1
			}

			return n, nil
		case *IfLessThanZero:
			n := node.(*IfLessThanZero)

			copiedValues := map[string]interface{}{}

			for k, v := range values {
				copiedValues[k] = v
			}

			var value1, value2 interface{}
			n.True, value1 = updateAndEvaluate(n.True, copiedValues)
			n.False, value2 = updateAndEvaluate(n.False, copiedValues)
			if value1 == value2 {
				return n, value1
			}

			return n, nil
		case *ValueBinding:
			n := node.(*ValueBinding)

			valuesExtended := map[string]interface{}{}
			for k, v := range values {
				valuesExtended[k] = v
			}
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

			functionToArgValues[n.Function] = append(
				functionToArgValues[n.Function], argValues)

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

	// Remove arguments whose values are always the same.

	functionToArgValuesCopied := map[string][][]interface{}{}
	for k, v := range functionToArgValues {
		functionToArgValuesCopied[k] = v
	}

	for _, function := range functions {
		argsToRemove := []int{}

		for i, arg := range function.Args {
			possibleValues := funk.Uniq(
				funk.Map(functionToArgValuesCopied[function.Name],
					func(s []interface{}) interface{} { return s[i] })).([]interface{})

			if len(possibleValues) == 1 {
				switch possibleValues[0].(type) {
				case int32:
					function.Body = &ValueBinding{arg, &Int{possibleValues[0].(int32)}, function.Body}
					argsToRemove = append(argsToRemove, i)
				case float32:
					function.Body = &ValueBinding{arg, &Float{possibleValues[0].(float32)}, function.Body}
					argsToRemove = append(argsToRemove, i)
				}
			}
		}

		if len(argsToRemove) == 0 {
			continue
		}

		functions = append(functions, &Function{"main", []string{}, main})
		for _, f := range functions {
			for _, application := range f.Body.Applications() {
				if application.Function == function.Name {
					newArgs := []string{}
					for i, arg := range application.Args {
						if !funk.ContainsInt(argsToRemove, i) {
							newArgs = append(newArgs, arg)
						}
					}

					application.Args = newArgs
				}
			}
		}
		functions = functions[:len(functions)-1]

		newArgs := []string{}
		for i, arg := range function.Args {
			if !funk.ContainsInt(argsToRemove, i) {
				newArgs = append(newArgs, arg)
			}
		}

		function.Args = newArgs
	}

	return main
}
