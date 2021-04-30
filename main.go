package main

import (
	"github.com/shubhamrai94/go-future/future"
	"errors"
	"fmt"
	"time"
)

func test(a, b, c int) (int, error) {
	time.Sleep(2 * time.Second)

	if c == 0 {
		return 0, errors.New("Divison by zero")
	}

	return (a + b) / c, nil
}

func testCallable1(f *future.Future) {
	time.Sleep(1 * time.Second)

	f.SetResult(10)
}

func testCallable2(f *future.Future) {
	f.SetResult(20)
}

func main() {
	f := future.New(test)(1, 2, 3)
	f.AddDoneCallback(testCallable1)

	fmt.Printf("Running: %v\n", f.Running())
	fmt.Printf("Done: %v\n", f.Done())
	fmt.Println(f.Result(1 * time.Second))

	// fmt.Printf("Cancel: %v\n", f.Cancel())
	// fmt.Printf("Cancelled: %v\n", f.Cancelled())

	fmt.Printf("Running: %v\n", f.Running())
	fmt.Printf("Done: %v\n", f.Done())
	fmt.Println(f.Result(1 * time.Second))

	fmt.Printf("Cancel: %v\n", f.Cancel())
	fmt.Printf("Cancelled: %v\n", f.Cancelled())

	fmt.Printf("Running: %v\n", f.Running())
	fmt.Printf("Done: %v\n", f.Done())
	fmt.Println(f.Result(1 * time.Second))

	time.Sleep(5 * time.Second)
	fmt.Println(f.Result(0))

	f.AddDoneCallback(testCallable2)

	time.Sleep(1 * time.Second)
	fmt.Println(f.Result(0))
}

