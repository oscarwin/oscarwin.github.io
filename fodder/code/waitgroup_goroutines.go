package main

import (
	"fmt"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		fmt.Println("running in goroutine 1")
		wg.Done()
	}()

	go func() {
		fmt.Println("running in goroutine 2")
		wg.Done()
	}()

	wg.Wait()
	fmt.Println("main exit")
}
