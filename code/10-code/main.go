package main

import (
	"fmt"
	"strings"
)

func Identity[T any](x T) T {
	return x
}

func Add[T int | int64 | float64](a, b T) T {
	return a + b
}

func Zero[T any]() T {
	var zero T
	return zero
}

func Max[T ~int | ~int64 | ~float64](a, b T) T {
	return max(a, b)
}

type UserID int

type Stringer interface {
	String() string
}

type Person struct {
	Name string
}

func (p Person) String() string {
	return p.Name
}

func Join[T Stringer](items []T, sep string) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, item.String())
	}
	return strings.Join(parts, sep)
}

type Number interface {
	~int | ~float64
}

func Del[T Number](a, b T) T {
	return a - b
}

func Map[T, U any](items []T, fn func(T) U) []U {
	out := make([]U, 0, len(items))
	for _, item := range items {
		out = append(out, fn(item))
	}
	return out
}

func main() {
	fmt.Println(Identity(42))
	fmt.Println(Identity("hello"))
	fmt.Println(10, 20.0)
	fmt.Println(Zero[int]())
	fmt.Println(Zero[*int]())

	var a, b UserID = 1, 2
	fmt.Println(Max(a, b))

	p1 := Person{"小明"}
	p := []Person{{"张三"}, {"李四"}}

	fmt.Println(p1)
	fmt.Println(p)
	fmt.Println(Join(p, ","))

	fmt.Println(Del(2, 1))

	nums := []int{1, 2, 3, 4}
	squares := Map(nums, func(n int) int { return n * n })
	fmt.Println(squares)

	names := Map(nums, func(n int) string {
		return fmt.Sprintf("用户%d", n)
	})

	fmt.Println(names)
}
