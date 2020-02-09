---
title: GO并发编程简介
date: 2020-01-29 14:28:43
tags:
    - go
categories: 技术博文
comments:  true
---

每当提起 go 的优点时，不得不提的一点是 go 支持原生协程。对于传统的 C 或者 Java 语言，需要通过线程来实现并发，然后通过互斥锁来实现线程同步。而 go 则通过协程(goroutines)来实现并发，通过通道(channels)来实现同步。go 语言这种并发模型被称为 CSP(Com-municating Sequential Processes)。<!-- more -->

go 并发模型里两个最重要的概念是 goroutines 和 channels：
- goroutines：一个 goroutine 就是独立运行的函数，就像是一个单独运行的线程一样，但是它并不是线程，被称为协程，协程可以理解为更轻量级的线程。
- channels：通道就像一个管道一样，协程之前可以通过通道来传递数据，实现同步和数据通信。channels 是协程安全的，因此读写时不需要加锁。

如果你是首次接触 go 语言，那么 go 的并发编程一定会让你惊讶不已，下面实现一个简单的并发程序。

```go
package main

import "fmt"
import "time"

func main() {
	fmt.Println("main start")
	go echo()                          // 调用 echo 函数开启协程
	fmt.Println("main end")
	time.Sleep(500 * time.Millisecond) // 休眠500ms
}

func echo() {
	fmt.Println("echo a line")
}
```

输出：
```
$: go run simple_goroutines.go 
main start
main end
echo a line
```

这个程序是怎么运行的呢？这个程序中，首先 main 函数打印出 “main start”，然后通过一个 go 关键词就开启了一个协程，在这个协程了执行了 echo 函数，开启协程后主程序还继续往下运行打印出 “main end”，之后主程序休眠 500ms 等待协程运行结束。

## 用 runtime.Gosched() 进行协程调度

在上面的例子中，在主程序结束前休眠了 500ms。休眠 500ms 的目的是为了保证在主程序退出前，开启的协程已经执行完毕。当使用 go 关键词开启协程时，实际上协程并非立刻就运行起来，需要等待 go runtime 的调度，至于如何调度那又是另一件事了，这里不展开说。当然使用休眠的方式，代码看起来很 low，于是换另一种高大上一点的方式。

```go
package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("outside a goroutine")
	go func() {
		fmt.Println("inside a goroutine")
	}()
	fmt.Println("outside a goroutine again")
	runtime.Gosched()
}
```

在这一段代码中使用 runtime.Gosched() 这一行代码替换了原来的 sleep 代码，看起来舒服了不少。Gosched 函数的作用就是命令 runtime 挂起当前协程，把运行的机会让给其他协程。

如果你需要开启一个协程执行数据库查询，Gosched 让主协程让出当前的 CPU，查询协程开始与数据库通信进行查询，但是可能查询到一半时又发生了协程调度，查询协程让出 CPU，主协程开始执行并退出了进程，这样导致数据库查询并未完成，对于这种情况，Gosched 不能满足要求，接下来介绍更加靠谱的方案。

另外，除了使用函数来开启协程外，还可以通过一个匿名函数来开启协程，如本例中就使用了一个匿名函数来开启协程。


## 使用 channel 实现并发程序

```go
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
```

channel 称为通道，实现了协程之间的通信，channel 类似于进程间通信方式中的消息队列，而且 channel 是协程安全的。协程安全的意思就是，在读写 channel 时无需加锁。与之相反的是，map 是非协程安全的，在多个协程并发读写同一个 map 时会引发 panic。因此对于并发读写 map 的场景往往采用加互斥锁的方式。

接着再聊聊上面并发程序的实现。本例中首先初始化了一个带缓冲区的 channel，然后开启了两个协程，每个协程退出前往 channel 中写入一个空结构体。最后主协程循环从 channel 中读两次，获取到两个对象说明两个协程都已经运行。

这种并发方式的实现一是不够优雅，二是不够高效。不够高效是因为 channel 的性能并不是非常好。这一点需要注意，不要像 channel 里写入大量的数据，可以向 channel 里写入对象的指针而不是对象本身来提高性能。

## 使用 WaitGroup 实现并发程序

```go
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
```

在 go 语言里实现并发程序最佳的实践是通过 WaitGroup 来实现。WaitGroup 通过标识事件发生与结束来标记事件的状态，类似于信号量。通过 Add 方法将计数增加，Add 函数中的参数表示增加的计数数量，通过 Done 函数将计数减一，最后 Wait 函数等待计数变为 0，表示所有的的事件都结束了。