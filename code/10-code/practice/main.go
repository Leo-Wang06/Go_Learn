package main

import "fmt"

func Identity[T any](x T) T {
	fmt.Printf("%T", x)
	return x
}

func Max[T int | int64 | float64](a, b T) T {
	return max(a, b)
}

func Zero[T any]() T {
	var zero T
	return zero
}

type Celsius float64

func Add[T ~float64](a, b T) T {
	return  a + b
}

func main() {
	i := 25
	s := "小明"
	iArry := []int{1, 2, 3, 4}

	Identity(i)
	Identity(s)
	Identity(iArry)

	Max(1 ,2)
	Max(1 ,2.0)

	Zero[int]()

	var a Celsius = 22.4
	var b Celsius = 12.4

	fmt.Println(Add(a, b))
}
