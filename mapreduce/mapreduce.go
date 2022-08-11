package mapreduce

import (
	"reflect"
)

// Map 返回新slice，对slice各元素进行映射执行function
// Example:
//	func triple(a int) int { return a*3 }
//	list := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
//	res := Map(list, triple)
func Map(slice, function interface{}) interface{} {
	return _map(slice, function, false)
}

// MapInplace 就地修改slice，对slice各元素进行映射执行function
// Example:
//	func triple(a int) int { return a*3 }
//	list := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
//	MapInplace(list, triple)
func MapInplace(slice, function interface{}) interface{} {
	return _map(slice, function, true)
}

// Reduce 对slice各元素进行pairFunc规约
// 	如果slice为空，返回zero
//  如果slice只有1个元素，则返回该元素
// 	否则对各元素进行pairFunc规约
// Example:
//	func multiply(a, b int) int { return a*b }
//	list := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
//	factorial := Reduce(list, multiply, 1).(int)
func Reduce(slice, pairFunc, zero interface{}) interface{} {
	sliceInType := reflect.ValueOf(slice)
	if sliceInType.Kind() != reflect.Slice {
		panic("reduce: not slice")
	}

	n := sliceInType.Len()
	if n == 0 {
		return zero
	}

	elemType := sliceInType.Type().Elem()
	fn := reflect.ValueOf(pairFunc)
	if !verifyFuncSignature(fn, elemType, elemType, elemType) {
		str := elemType.String()
		panic("reduce: function must be of type func(" + str + ", " + str + ") " + str)
	}

	var ins [2]reflect.Value
	out := sliceInType.Index(0) // By convention, fn(zero, in[0]) = in[0].
	// Run from index 1 to the end.
	for i := 1; i < n; i++ {
		ins[0] = out
		ins[1] = sliceInType.Index(i)
		out = fn.Call(ins[:])[0]
	}

	return out.Interface()
}

// Filter 返回新slice，对slice各元素执行function过滤
// Example:
// 	list := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
//	even := Filter(list, func(a int) bool { return a%2 == 0 })
func Filter(slice, function interface{}) interface{} {
	result, _ := filter(slice, function, false)
	return result
}

// FilterInplace 就地修改slice，对slice各元素执行function过滤
// Example:
// 	list := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
//	odd := Filter(list, func(a int) bool { return a%2 == 1 })
func FilterInplace(slicePtr, function interface{}) {
	sliceInType := reflect.ValueOf(slicePtr)
	if sliceInType.Kind() != reflect.Ptr {
		panic("filterInplace: not a pointer to slice")
	}
	_, n := filter(sliceInType.Elem().Interface(), function, true)
	sliceInType.Elem().SetLen(n)
}

func _map(slice, function interface{}, inplace bool) interface{} {
	// check the slice type is Slice
	sliceInType := reflect.ValueOf(slice)
	if sliceInType.Kind() != reflect.Slice {
		panic("_map: not slice")
	}

	// check the function signature
	fn := reflect.ValueOf(function)
	elemType := sliceInType.Type().Elem()
	if !verifyFuncSignature(fn, elemType, nil) {
		panic("_map: function must be of type func(" + elemType.String() + ") outputElemType")
	}

	sliceOutType := sliceInType
	if !inplace {
		sliceOutType = reflect.MakeSlice(reflect.SliceOf(fn.Type().Out(0)), sliceInType.Len(), sliceInType.Len())
	}
	for i := 0; i < sliceInType.Len(); i++ {
		sliceOutType.Index(i).Set(fn.Call([]reflect.Value{sliceInType.Index(i)})[0])
	}

	return sliceOutType.Interface()
}

var boolType = reflect.ValueOf(true).Type()

func filter(slice, function interface{}, inplace bool) (interface{}, int) {
	// check the slice type is Slice
	sliceInType := reflect.ValueOf(slice)
	if sliceInType.Kind() != reflect.Slice {
		panic("filter: not slice")
	}

	// check the function signature
	fn := reflect.ValueOf(function)
	elemType := sliceInType.Type().Elem()
	if !verifyFuncSignature(fn, elemType, nil) {
		panic("filter: function must be of type func(" + elemType.String() + ") bool")
	}

	var filtered []int
	for i := 0; i < sliceInType.Len(); i++ {
		if fn.Call([]reflect.Value{sliceInType.Index(i)})[0].Bool() {
			filtered = append(filtered, i)
		}
	}

	sliceOutType := sliceInType
	if !inplace {
		sliceOutType = reflect.MakeSlice(sliceInType.Type(), len(filtered), len(filtered))
	}
	for i := range filtered {
		sliceOutType.Index(i).Set(sliceInType.Index(filtered[i]))
	}

	return sliceOutType.Interface(), len(filtered)
}

// verifyFuncSignature 检查函数的参数和返回类型
// NumIn()  检查函数的入参
// NumOut() 检查函数的返回值
func verifyFuncSignature(fn reflect.Value, types ...reflect.Type) bool {
	// check it is a function
	if fn.Kind() != reflect.Func {
		return false
	}

	// NumIn()  - returns a function type's input parameter count.
	// NumOut() - returns a function type's output parameter count.
	if (fn.Type().NumIn() != len(types)-1) || (fn.Type().NumOut() != 1) {
		return false
	}

	// In() - returns the type of a function type's i'th input parameter.
	for i := 0; i < len(types)-1; i++ {
		if fn.Type().In(i) != types[i] {
			return false
		}
	}

	// Out() - returns the type of a function type's i'th output parameter.
	outType := types[len(types)-1]
	if outType != nil && fn.Type().Out(0) != outType {
		return false
	}

	return true
}
