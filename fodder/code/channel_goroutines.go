package main

import (
	"fmt"
)

func main() {
	ch := make(chan struct{}, 2)
	go func() {
		defer func() {
			ch <- struct{}{}
		}()
		fmt.Println("running in goroutine 1")
	}()

	go func() {
		defer func() {
			ch <- struct{}{}
		}()
		fmt.Println("running in goroutine 2")
	}()

	for i := 0; i < 2; i++ {
		<-ch
	}
	fmt.Println("main exit")
}
