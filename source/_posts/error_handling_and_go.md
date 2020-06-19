---
title: Go 语言中的错误处理 
date: 2020-06-19 11:28:43
tags:
    - go
categories: 技术博文
comments:  true
---

在使用 Go 语言编写应用程序时，会调用内建函数或者进行远程 RPC 调用，调用这些函数时可能会返回错误。Go 语言里没有使用 try...catch 这类机制，而是使用 error 来表示错误。以 os.Open 函数为例，当打开文件出错时会返回一次错误。

```go
func Open(name string) (file *File, err error)
```

对于这种错误，典型的处理方式是判断是否发生错误，如果发生了错误便将错误打印出来或者将错误返回给上层进行处理，代码如下。
```go
f, err := os.Open("filename.ext")
if err != nil {
    log.Fatal(err)
}
// do something with the open *File f
```

在代码结构比较简单的情况，使用这种方式处理错误基本能够满足要求，但是随着代码逻辑复杂化，这种处理错误的方式会遇到一些问题，本文想系统的探讨一下在 Go 语言里如何处理错误。<!-- more -->

## 错误类型

error 是一个 interface，而且是一个内建类型，这个 interface 只实现了一个方法 Error，该方法返回一段字符串来描述错误信息。

```go
type error interface {
    Error() string
}
```

### 默认类型

使用的最多的 error 实现是 errors 包中定义的 errorString，使用一个 string 字段来保存所有的错误信息。
```go
// errorString is a trivial implementation of error.
type errorString struct {
    s string
}

func (e *errorString) Error() string {
    return e.s
}
```

errorString 是一个非导出类型，errors 包提供了 New 方法来构建该对象。

```go
// New returns an error that formats as the given text.
func New(text string) error {
    return &errorString{text}
}
```

New 方法只能传入固定的字符串，不够灵活，因此 fmt 包里提供了另外一个函数 fmt.Errorf，该函数会调用 errors.New 来生成 errorString 对象。在实际的编码过程中，这个函数使用得更加频繁。

```go
func Errorf(format string, a ...interface{}) error {
	p := newPrinter() // 通过 sync.Pool 池化内存分配
  p.wrapErrs = true
	p.doPrintf(format, a)
	s := string(p.buf)
	var err error
	if p.wrappedErr == nil {
		err = errors.New(s)
	} else {
		err = &wrapError{s, p.wrappedErr}
	}
	p.free()
	return err
}
```

仍然以打开文件这个函数为例，此时处理错误的方式可能是这样。

```go
f, err := os.Open("filename.ext")
if err != nil {
    log.Fatal(err)
    return fmt.Errorf("open filename.ext failed, err: %v", err)
}
```

使用默认类型时，可以通过定义全局的错误变量来区分不同的错误，如 EOF 错误的定义与处理如下。

```go 
var EOF = errors.New("EOF")

buf := make([]byte, 100)
n, err := r.Read(buf)
buf = buf[:n]
if err == io.EOF {
    log.Fatal("read failed:", err)
}
```

### 自定义类型

默认类型错误适合简单场景和基础库来使用，在实际的业务场景下，可能发生的错误千奇百怪，不同的错误可能处理方式会不一样。这个时候默认 errorString 类型就非常难以满足现在要求了。比如调用了一个 RPC 方法，返回超时就进行重试，返回资源不存在则打印错误日志并返回。对于这种情况 errorString 类型的错误，只能像下面这段代码这样，通过判断 error 字符串中是否存在某个关键词来判断错误，这实在不是一个好主意。

```go
rsp, err := rpc.Call(req)
if err != nil {
    if strings.Contains(err.Error(), "timeout") {
        // try again
    } else {
        log.Error(err)
        return err
    }
}
```

因为 error 只是一个 interface，就是为了能够放开可以让用户自定义错误类型。自定义错误类型通常有两种思路，第一种是给每一种类型的错误都定义的结构，第二种是通过错误码区分不同的错误。

#### 通过错误类型区分错误

比如对于 HTTP 调用失败的类型可以定义一个类型：

```go
type HttpError struct {
    code int
    message string
}
```

对于内建函数调用失败可以定义一个类型：

```go
type BuiltInError struct {
    message string
}
```

当处理错误的时候，通过 assert 错误对象来判断是哪种错误，当然也可以使用 switch case 的方法来对不同的类型进行处理。

```go
err := f()
if err != nil {
    if _, ok := err.(*HttpError); ok {
        // 是 HTTP 的错误
    }
}
```

使用这种方法的痛点在于，对于不同类型的错误，需要维护不同的类型，维护成本会很大。对于更细节的错误仍然不容易区分，比如一个 open 函数，因为文件不存在而打开错误，还是本身系统故障打开错误，这两种错误的处理方式可能也不一样。如果每种都定义一个错误类型，那么仍然是维护成本的问题。

#### 通过错误码区分错误

第二种思路是通过错误码区分，只定义一个错误结构。Code 是错误码，Message 保存具体的错误信息，IsTemporary 表示该错误是否为一个临时性错误，这对于我们决定是否要进行重试提供了主要的判断依据。

```go
const (
    ErrCodeRpcNotFound = 1
    ErrCodeRpcTimeout = 2
)

type CustomError struct {
    Code        int    `json:"code"`
    Message     string `json:"message"`
    IsTemporary bool   `json:"isTemporary"`
}
```

这时候错误处理就变成了：

```go
rsp, err := rpc.Call(req)
if err != nil {
    if e, ok := err.(*CustomError); ok {
        if e.Code == ErrCodeRpcNotFound {
            log.Error(err)
            return nil
        } else if e.Code == ErrCodeRpcTimeout {
            // retry
        } else {
            return err
        }
    }
} 
```

## 错误处理

### 输出更详细的错误信息

```go
func AuthenticateRequest(r *Request) error {
    err := authenticate(r.User)
    if err != nil {
        return err
    }
    return nil
}
```

上面这段程序如果返回一段错误 `“No such file or directory"`，你是否会抓狂，这样一段错误你压根不知道从哪里去定位这样一个错误。

不过还好已经有人给出解决方案，`github.com/pkg/errors` 这个库可以将错误发生时的整个调用栈信息全部打印出来，比如下面这段程序。

```go
package main

import (
	"fmt"

	"github.com/pkg/errors"
)

func fn1() error {
	return errors.New("not found")
}

func fn2() error {
	err := fn1()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := fn2()
	fmt.Printf("err: %+v\n", err)
}
```

打印出来的错误信息如下，那么就容易定位到错误发生的位置，不用在程序里每次处理一次错误就打印一次错误。

```
err: not found
main.fn1
        /Users/liuteng/go/src/mycoding/example/error/error.go:33
main.fn2
        /Users/liuteng/go/src/mycoding/example/error/error.go:39
main.main
        /Users/liuteng/go/src/mycoding/example/error/error.go:50
runtime.main
        /usr/local/go/src/runtime/proc.go:203
runtime.goexit
        /usr/local/go/src/runtime/asm_amd64.s:1357
```

这个包该怎么用，主要有几个函数

```go
// 新建错误
func New(message string) error
func WithStack(err error) error
func WithMessage(err error, message string) error
// 在错误上添加额外信息
func Wrap(err error, message string) error
// 找出最原始的错误信息
func Cause(err error) error
```

### 不漏处理错误也不重复处理错误

错误暴露一些程序运行过程中才会发生的信息，因此不应该被忽略，不漏错误最好的解决方案就是保证在程序的主入口一定处理错误。

不重复处理错误是指对于一个错误同时进行两次相同的处理，如下所示进行了两次错误处理，就存在重复处理错误的问题。在函数 fn2 中应该直接返回该错误，而不需要打印出该错误，这样日志里不会导出充斥着同样的日志，而影响问题的排查。
```go
package main

import (
	"fmt"

	"github.com/pkg/errors"
)

func fn1() error {
	return errors.New("not found")
}

func fn2() error {
    err := fn1()
    if err != nil {
	    // 第一次处理错误
	    log.Error(err)
      return err
    }
    return nil
}

func main() {
    err := fn2()
    // 第二次处理错误
    log.Error("err: %+v\n", err)
}
```

### 断言错误的行为而不是错误的类型

对于多种错误类型往往会采用相同的错误处理方式，比如对于 HTTP 的 400、404、401 几个错误码，都不需要进行重试，因此对于错误只要赋予一个属性——是否是临时性错误。根据是否是临时性错误来决定是否要进行重试。

```go
type CustomError struct {
    Code        int    `json:"code"`
	  Message     string `json:"message"`
	  IsTemporary bool   `json:"isTemporary"`
}

if err := fn(); err != nil {
    if e, ok := errors.Cause(err).(*CustomError); ok {
        if e.IsTemporary {
            // 针对临时错误的处理方案
        } else {
            // 针对非临时错误的处理方法
        }
    }
}
```

## go error 与 panic 的一点想法

Go 语言的错误处理是一个被讨论比较多的话题，有人诟病这种充斥着各种错误检查的代码，也有人认为 go 将错误与异常区分开来处理，避免了异常被滥用，是非常好的一种设计思路。Go 语言中的 panic 与其他语言中的异常机制是一致的，如果异常没有被捕获就一直向上抛。

**异常与错误的区别是什么？**

错误本身就是正常逻辑的一部分，因为错误是预期之内的，比如网络超时了，库存扣减失败了，转账失败了，这些错误都是在编写代码时就预料到可能会出现这些错误，而且对于这些错误会采取不同的处理手段，因此错误本身就应该被检查和处理，他是正常代码的一部分。当程序遇到错误时，还可以继续向下执行，这是错误的一个很重要的特点。

而异常应该是指不可预期的错误，比如内存分配失败了，MySQL 创建连接失败了。内存分配失败了，程序根本无法向下跑，所以要抛异常。MySQL 创建连接失败了，那么应用程序无法正常读写数据库，那基本上也做不了事了，即使 recover 住了也并没有什么意义。假设对于一个非常重要的接口，第一次调用失败那叫错误，然后重试了三次，达到了错误上限，如果这个错误应用程序无法容忍，那么就可以升级为异常，如果这个错误可以容忍那么仍然可以以错误的方式处理它。

当然了，我也很苦恼代码里充斥着大量错误检查的逻辑，但是每一种设计哲学的选择都是有得有失的，很难有完美的解决方案。而 go 只是选择了一种，不那么好看，但是挺实用的方案。

## 自定义错误包

通过以上分析，借助 `github.com/pkg/errors` 封装了一套自定义的错误处理库。主要实现了 New，Warp，Warpf，Cause 几个函数。可以通过错误码确定错误的类型，也可以通过行为来判断是否是临时性错误。具体的错误信息和栈信息也都保存了。

```go
package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

// error code define
const (
    HttpNotFound = 1
)

type CustomError struct {
    Code        int    `json:"code"`
	Message     string `json:"message"`
	IsTemporary bool   `json:"isTemporary"`
}

func New(code int, message string, isTemporary bool) error {
	err := &CustomError{
		Code:        code,
		Message:     message,
		IsTemporary: isTemporary,
	}
	return errors.WithStack(err)
}

func Cause(err error) error {
	return errors.Cause(err)
}

func Wrap(err error, message string) error {
	return errors.Wrap(err, message)
}

func Wrapf(err error, message string, args ...interface{}) error {
	return errors.Wrapf(err, message, args...)
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

func (e *CustomError) GetCode() int {
	return e.Code
}

func (e *CustomError) GetMessage() string {
	return e.Message
}

func (e *CustomError) GetIsTemporary() bool {
	return e.IsTemporary
}
```

### 生成错误

生成错误直接调用 New 函数

```go
return New(HttpNotFound, "resource not found", false)
```

### 检查错误

检查错误后，如果发生了错误，该错误无法在当前函数处理的话，则将错误直接返回，不需要探知具体是何种错误发生了。

```go
if err != nil {
    return err
}
```

### 处理错误

错误优先按照错误的行为进行处理，如果行为处理不完全满足要求，就可以按照类型进行处理。打印错误信息的时候需要使用 +v 格式化，这样才能打印出完整的调用栈信息。

```go
if err != nil {
    if e, ok := errors.Cause(err).(*CustomError); ok {
        if e.IsTemporary {
            // 按照行为处理错误
        }
        if e.Code == HttpNotFound {
            // 按照类型处理错误
        }
    }
    // 打印完整的错误信息
    log.Errorf("%+v", err)
}
```

## 结论

1. 不要随意忽略错误，也不要重复处理错误，按照错误的行为处理错误是一个更简洁的方案
2. 默认的错误定义满足不了需求，就自定义错误类型
3. 不处理错误时，无需检查和打印错误的信息，处理错误时，打印出完整的错误调用栈信息


## 参考
[1] [error-handling-and-go](https://blog.golang.org/error-handling-and-go)
[2] [gocon-spring-2016](https://dave.cheney.net/paste/gocon-spring-2016.pdf)
[3] [Go语言中的错误处理](https://ethancai.github.io/2017/12/29/Error-Handling-in-Go/)
[4] [pkg/errors](https://pkg.go.dev/github.com/pkg/errors?tab=doc)


