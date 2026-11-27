package main

import (
	"fmt"
	"time"
)

func divmod(a, b int) (q, r int) {
	q = a / b
	r = a % b
	return
}

func add(a, b int) int {
	return a + b
}

type BinOp func(int, int) int

func apply(a, b int, op BinOp) int {
	return op(a, b)
}

func sum() int {
	count := 0
	defer func() {
		count++
		fmt.Println("count:", count)
	}()

	count++

	return count
}

func measureTime() {
	start := time.Now()
	defer func() {
		fmt.Println("耗时：", time.Since(start))
	}()
}

func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			// panic(r) 
		}
	}()

	return a / b, nil
}

func main() {
	var funcs []func()

	for i := 0; i < 3; i++ {
		funcs = append(funcs, func() {
			fmt.Println(i)
		})
	}

	for _, f := range funcs {
		f()
	}

	q, r := divmod(2, 1)

	fmt.Println(q, r)

	fmt.Println(apply(1, 2, add))

	sum()

	measureTime()

	result, err := safeDivide(10, 0)
	fmt.Println(result, err) // 0 panic: runtime error: integer divide by zero

}
