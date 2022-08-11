package mapreduce

import (
	"testing"
)

// run all: go test -v github.com/beyondyyh/libs/mapreduce

type employee struct {
	Name     string
	Age      int
	Vacation int
	Salary   int
}

var employee_list = []employee{
	{"Hao", 44, 0, 8000},
	{"Bob", 34, 10, 5000},
	{"Alice", 23, 5, 9000},
	{"Jack", 26, 0, 4000},
	{"Tom", 48, 9, 7500},
	{"Marry", 29, 0, 6000},
	{"Mike", 32, 8, 4000},
}

// 求和
func sum(a, b int) int {
	return a + b
}

// 乘积
func multiply(a, b int) int {
	return a * b
}

// 阶乘
func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

// run: go test -v -run TestMap
func TestMap(t *testing.T) {
	// string
	l1 := []string{"1", "2", "3", "4", "5"}
	l1_1 := Map(l1, func(s string) string {
		return s + s + s
	})
	t.Logf("Not inplace: %v", l1_1)

	MapInplace(l1, func(s string) string {
		return s + s + s
	})
	t.Logf("Inplace: %v", l1)

	// ints
	l2 := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	MapInplace(l2, func(a int) int {
		return a * 3
	})
	t.Log(l2)

	employee_list_2 := Map(employee_list, func(e employee) employee {
		e.Salary *= 2
		e.Age += 1
		return e
	})
	t.Logf("employee_list_2: %v", employee_list_2)
}

// run: go test -v -run TestReduce
func TestReduce(t *testing.T) {
	const size = 10
	a := make([]int, size)
	for i := range a {
		a[i] = i + 1 // 1...9
	}
	for i := 1; i < 10; i++ {
		out := Reduce(a[:i], multiply, 1).(int) // 从1到i的乘积，也就是i的阶乘
		expect := factorial(i)
		if expect != out {
			t.Fatalf("expected %d got %d", expect, out)
		}
	}
}

// run: go test -v -run TestFilter
func TestFilter(t *testing.T) {
	l3 := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	// 偶数
	even := Filter(l3, func(a int) bool { return a%2 == 0 })
	t.Logf("even: %v", even)
	// 奇数
	odd := Filter(l3, func(a int) bool { return a%2 == 1 })
	t.Logf("odd: %v", odd)

	// 老员工
	old := Filter(employee_list, func(e employee) bool {
		return e.Age > 40
	})
	t.Logf("old people: %v", old)

	// 高收入员工
	high_pay := Filter(employee_list, func(e employee) bool {
		return e.Salary >= 8000
	})
	t.Logf("high pay: %v", high_pay)

	// 无休假的员工
	no_vacation := Filter(employee_list, func(e employee) bool {
		return e.Vacation == 0
	})
	t.Logf("no vacation: %v", no_vacation)

	var sum = func(a, b employee) employee {
		return employee{Salary: a.Salary + b.Salary}
	}

	// 所有员工薪水之和
	total_pay := Reduce(employee_list, sum, 0).(employee).Salary
	t.Logf("total pay: %d", total_pay)

	// 30岁以下员工的薪资总和
	younger_pay := Reduce(
		Filter(employee_list, func(e employee) bool { return e.Age < 30 }).([]employee),
		sum,
		0).(employee).Salary
	t.Logf("younger pay: %d", younger_pay)

	// in-place 筛选年轻的员工，employee_list会被改变
	t.Logf("before younger filter, employee_list: %v", employee_list)
	FilterInplace(&employee_list, func(e employee) bool {
		return e.Age < 30
	})
	t.Logf("after younger filter, employee_list: %v", employee_list)
}
