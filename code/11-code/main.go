package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

)

func worker(id int) {
	fmt.Println("worker", id, "start")
	time.Sleep(100 * time.Millisecond)
	fmt.Println("worker", id, "done")
}

type result struct{ data string; err error}

func fetch(id int) (string, error) {
	time.Sleep(50 * time.Millisecond)   // 模拟耗时操作
	if id <= 0 {
		return "", errors.New("invalid id")
	}
	return fmt.Sprintf("user-%d", id), nil
}

func produce(out chan <- int) {
	for i := 1; i <= 3; i++ {
		out <- i
	}
	close(out)
}

func consume(in <-chan int) {
	for v := range in {
		fmt.Println(v)
	}
}

func main() {
	go worker(1)
	go worker(2)
	time.Sleep(300 * time.Millisecond)
	fmt.Println("main code")

	go fmt.Println("hello")
	time.Sleep(100 * time.Millisecond)

	var wg sync.WaitGroup
	wg.Add(1)
	go func () {
		defer wg.Done()
		fmt.Println("hello")
	}()
	wg.Wait()

	ch := make(chan result, 1)
	go func() {
		data, err := fetch(1)
		ch <- result{data: data, err: err}
	}()

	r := <-ch
	if r.err != nil {
		fmt.Println("failed:", r.err)
		return
	}
	fmt.Println(r.data)

	for i := 1; i <= 3; i++ {
		go func(id int) {fmt.Println(id)}(i)
	}

	go func() {
		defer func(){
			if r := recover();r != nil {
				fmt.Println("recovered", r)
			}
		}()
		panic("boom")
	}()

	time.Sleep(100 * time.Millisecond)

	ch2 := make(chan int, 2)
	ch2 <- 1
	ch2 <- 2
	fmt.Println(len(ch2),cap(ch2))
	fmt.Println(<-ch2, <-ch2)

	ch3 := make(chan int)
	go produce(ch3)
	consume(ch3)

	ch4 := make(chan int, 3)
	ch4 <- 1
	ch4 <- 2
	ch4 <- 0
	close(ch4)

	for v := range ch4 {
		fmt.Println(v)
	}

	done := make(chan struct{})

	go func() {
		<- done
		fmt.Println("worker1 醒来")
	}()
	go func() {
		<- done
		fmt.Println("worker2 醒来")
	}()
	go func() {
		<- done
		fmt.Println("worker3 醒来")
	}()

	time.Sleep(50 * time.Millisecond)
	close(done)

	ch6 := make(chan string)
	ch7 := make(chan string)

	go func ()  {
		time.Sleep(20 * time.Millisecond)
		ch6 <- "one"
	}()
	go func() {
		time.Sleep(10 * time.Millisecond)
		ch7 <- "two"
	}()

	select {
	case v := <-ch6:
		fmt.Println("ch6:", v)
	case v := <-ch7:
		fmt.Println("ch7:", v)
	}

}
