---
title: Thrift 
date: 2020-01-29 14:28:43
tags:
    - go
categories: 技术博文
comments:  true
---

Thrift 是一个跨语言的RPC框架，支持20多种语言。Thrift 抽象出数据传输层，数据序列化协议层和应用层，并一一对其进行了实现。Thrift 定义了一套通用的基本类型，兼容各种语言。代码生成工具，根据 IDL 文件生成对应语言的代码。

Thrift 整体架构

![img](https://bytedance.feishu.cn/space/api/box/stream/download/asynccode/?code=a54bd7835318e61cc0277faabf7f0953_8f118824ce50c961_boxcnoP6baBXTz6dIwEkOlf6bnh_fh0MtNkA9zrkw70p0r3wuSkP932M8qsz)

Thrift 的分层架构图

![img](https://bytedance.feishu.cn/space/api/box/stream/download/asynccode/?code=8b7db830655485c114fdc8d15d1540ac_8f118824ce50c961_boxcnxFPfYu3SiNYK2NXub5491x_eICajZodwFI14ZMJx9SdKydzEjh2XwNh)

- 传输层和协议层的代码由 thrift 的基础类库提供。

- 通过 IDL 生成用户自定义的类型，以及服务打桩

- 最后客户端程序和服务端的程序由业务方自行编写来完成

> The classes within the Apache Thrift libraries are typically named with a leading capital T, for example, TTransport, TProtocol, and TServer.

传输层

![img](https://bytedance.feishu.cn/space/api/box/stream/download/asynccode/?code=8b7db830655485c114fdc8d15d1540ac_8f118824ce50c961_boxcnxFPfYu3SiNYK2NXub5491x_eICajZodwFI14ZMJx9SdKydzEjh2XwNh)

传输层封装了对字节的读写接口，传输层不关心底层设备是什么，写入的设备对象可以是文件、内存和网络等。

传输层interface定义

| open()    | Prepares the transport for I/O operations                    |
| --------- | ------------------------------------------------------------ |
| close()   | Shuts down the transport                                     |
| isOpen()  | Returns true if the transport is open, false otherwise       |
| read()    | Reads bytes from the transport                               |
| readAll() | Reads an exact number of bytes from the transport (blocking or erroring if they aren’t available) |
| write()   | Writes bytes to the transport (the transport may buffer the operation) |
| flush()   | Forces any buffered bytes to be written to the underlying device |

TBufferedTransport与TFramedTransport区别

TBufferedTransport是直接对网络IO的封装，TFramedTransport是在TBufferedTransport的基础上给每桢数据添加了一个每桢数据大小的头，对于使用非阻塞的服务器需要使用TFramedTransport模式。这也是为什么公司里生成C++的服务时需要带上--framed选项，因为C++服务器大多数是采用IO多路复用的形式，而Go服务器直接使用协程。

协议层

Thrift 基本语法

**基本类型** 

- bool: 布尔类型，true或者false

- byte: 字节类型，长度为8位，有符号

- i16: 16位有符号整形

- i32: 32位有符号整形

- i64: 64位有符号整形

- double: 64位浮点数

- binary: byte数组

- string 一串编码未知的文本或二进制串 

**容器** 

Thrift支持三种容器：

- list 列表容器，对应go的slice

- set 元素无序唯一的集合容器，对应go的

- map<type1, type2> 一个key唯一的key-value映射map，对应go的map

**结构体** 

Thrift struct 类似于 C 语言的结构体，定义了一个用于多语言的通用对象，基本上相当于面向对象语言中的一个类。一个结构体的例子如下：

```
struct Example {
  1:i32 number=10,
  2:i64 bigNumber,
  3:double decimals,
  4:string name="thrifty"
}
```

结构体里的字段可以设定默认值。

**异常** 

异常的语法和结构体基本相同，除了异常使用的关键字是exception而不是struct。go语言没有异常，异常会转为go的error。

**服务** 

服务是基于 Thrift 类型定义的。服务的定义在语义上大概相当于面向对象语言中定义一个接口或一个纯虚抽象类。Thrift编译器会生成实现该接口的全功能客户端和服务端桩代码。

服务定义类似如下：

```
service  {
   ()
  [throws ()]
  ...
}
```

实例如下:

```
service StringCache {
  void set(1:i32 key, 2:string value),
  string get(1:i32 key) throws (1:KeyNotFound knf),
  void delete(1:i32 key)
}
```

**枚举** 

可以像 C/C++ 那样定义枚举类型，如：

```
enum TweetType {
  TWEET,
  RETWEET = 2,
  DM = 0xa,
  REPLY
}
struct Tweet {
  1: required i32 userId;
  2: required string userName;
  3: required string text;
  4: optional Location loc;
  5: optional TweetType tweetType = TweetType.TWEET
  6: optional string language = "english"
}
```

说明：

- 编译器默认从0开始赋值

- 可以将某个整数赋予某个常量

- 允许常量是十六进制整数

- 末尾没有逗号

- 给常量赋缺省值时，使用常量的全称

**注释** 

Thrift支持Shell注释风格，也支持C/C++的单行和多行注释风格。

```
# This is a valid comment. shell style
*/\* * This is a multi-line comment. \* Just like in C. \*/*
*// C++/Java style single-line comments work just as well.*
```

**命名空间** 

Thrift中的命名空间同C++中的namespace和java中的package类似，它们均提供了一种组织（隔离）代码的方式。因为每种语言均有自己的命名空间定义方式（如python中有module），thrift允许开发者针对特定语言定义namespace：

```
namespace cpp com.example.project  // a
namespace java com.example.project // b 说明： a．  转化成
namespace go com.example.project // go 的 namespace 就对应目录的路径
```

- 会被转化成命名空间：namespace com { namespace example { namespace project { 

- 或者转化成包名：package com.example.project

**文件包含** 

Thrift允许thrift文件包含，用户需要使用thrift文件名作为前缀访问被包含的对象，如：

```
#include "tweet.thrift"
...
struct TweetSearchResult {
  1: list tweets;
} * 注意包含时的tweet前缀
```

**常量** 

Thrift允许用户定义常量，复杂的类型和结构体可使用JSON形式表示。

```
const i32 INT_CONST = 1234;   // 分号是可选的
const map MAP_CONST = {"hello": "world", "goodnight": "moon"} // 复杂结构可以使用JSON表示
```

PS

- thrift的基本类型中没有无符号整形，这是为了兼容多种语言

Thrift RPC 数据交换

数据交换过程

Thrift 一次完整的数据交换过程如下图所示，分为四个步骤：

![img](https://bytedance.feishu.cn/space/api/box/stream/download/asynccode/?code=3273dfa8d20de4aa2141bb9c0b385f2f_8f118824ce50c961_boxcnTrP4OdSGG3VjiRXONSs1Tc_tiLlfFg6nZegftWUk4ZM9LzA0pTKKI5r)

1. 首先是客户端发起请求。客户端会先发送一段本次请求的元数据信息，称为Message。Message有四种类型（call，oneway，reply，exception），步骤一种Message的类型可以是call或者oneway。关于Message里具体的信息，后面会进行具体说明。

1. 客户端发送请求的结构体，该结构体是有代码生成工具生成的，保存的是请求方法的入参，每一个入参对应结构体里的一个字段。

1. 服务接收到请求并处理完成后需要返回结果给客户端。先也是发送一段元数据信息，元数据信息的类型会是reply或者exception。

1. 然后发送返回数据的结构体或者是一个exception。

> 1. Client sends a Message (type Call or Oneway). The TMessage contains some metadata and the name of the method to invoke.

> 1. Client sends method arguments (a struct defined by the generate code).

> 1. Server sends a Message (type Reply or Exception) to start the response.

> 1. Server sends a struct containing the method result or exception.

> The pattern is a simple half duplex protocol where the parties alternate in sending a Message followed by a struct. What these are is described below.

> Although the standard Apache Thrift Java clients do not support pipelining (sending multiple requests without waiting for an response), the standard Apache Thrift Java servers do support it.

Message结构

Message 主要包括三部分内容，分别是：

- 名称，是一个 string 类型

- Message type，有四种类型：call、oneway、reply、exception

- 序列号，是一个 int32 类型

其中序列号是由客户端生成，服务端在返回的时候会原封不动的回传这个序列号，客户端通过序列号来识别是哪个请求返回了。序列号可以是负值，当序列号溢出后，直接使用溢出的数据继续循环递增。

Call 和 oneway 出现在步骤1中，reply 和 exception 出现在步骤3中。reply 类型表示服务端成功处理并返回，exception 类型服务方法抛出了异常时用，但是 go 语言没有 exception，exception就是go语言的error。这也是为什么kite框架会规定业务错误不要通过err来返回，而是通过baseresp结构体来返回。因为业务错误不应该用异常来表示。

Oneway 类型从 0.9.3 版本才开始使用，更早的版本不会使用 oneway 类型。

Request struct

紧跟着message发送的是一个结构体，这个结构体并不是 IDL 里定义的结构体，而是 thrift_gen 生成的，以方法名+Args为名称，这个结构体里的每一个字段对应方法里的每一个参数。

官方文档里给出的例子，IDL里定义了一个add方法：

```
struct SharedStruct {
 1: i32 key
 2: string value
}
service SharedService {
 SharedStruct getStruct(1: i32 key)
}
```

生成的代码如下：

```
type GetStructArgs struct {
```

`  Key int32 `thrift:"key,1" json:"key"``

```
}
```

Response struct

返回的结构体就有一个ID标识为0的success字段，success字段的类型就是IDL中定义的返回类型。如果有定义异常处理，还有有其他另一个字段。

```
type GetStructResult struct {
```

`  Success *SharedStruct `thrift:"success,0" json:"success"``

```
}
```

序列化协议

序列化协议接口定义

使用Thrift时使用最多的就是binary协议，主要来分析binary协议的实现。

thrift协议的接口非常直接，因为就传输了message和struct，那么序列化协议的接口只要能序列化Message和任意结构的struct就可以满足任意传输的需求了。

为此thrift定义了一套通用的接口，如果要新建一种协议只要按照这个interface实现一套就可以创建一套新的协议，但是其它层可以继续复用。

```
type TProtocol interface {
  // 写Message
  WriteMessageBegin(name string, typeId TMessageType, seqid int32) error
  WriteMessageEnd() error
  // 写结构体
  WriteStructBegin(name string) error
  WriteStructEnd() error
  WriteFieldBegin(name string, typeId TType, id int16) error
  WriteFieldEnd() error
  WriteFieldStop() error
  WriteMapBegin(keyType TType, valueType TType, size int) error
  WriteMapEnd() error
  WriteListBegin(elemType TType, size int) error
  WriteListEnd() error
  WriteSetBegin(elemType TType, size int) error
  WriteSetEnd() error
  WriteBool(value bool) error
  WriteByte(value byte) error
  WriteI16(value int16) error
  WriteI32(value int32) error
  WriteI64(value int64) error
  WriteDouble(value float64) error
  WriteString(value string) error
  WriteBinary(value []byte) error
  // 读message
  ReadMessageBegin() (name string, typeId TMessageType, seqid int32, err error)
  ReadMessageEnd() error
  // 读结构体
  ReadStructBegin() (name string, err error)
  ReadStructEnd() error
  ReadFieldBegin() (name string, typeId TType, id int16, err error)
  ReadFieldEnd() error
  ReadMapBegin() (keyType TType, valueType TType, size int, err error)
  ReadMapEnd() error
  ReadListBegin() (elemType TType, size int, err error)
  ReadListEnd() error
  ReadSetBegin() (elemType TType, size int, err error)
  ReadSetEnd() error
  ReadBool() (value bool, err error)
  ReadByte() (value byte, err error)
  ReadI16() (value int16, err error)
  ReadI32() (value int32, err error)
  ReadI64() (value int64, err error)
  ReadDouble() (value float64, err error)
  ReadString() (value string, err error)
  ReadBinary() (value []byte, err error)
  Skip(fieldType TType) (err error)
  Flush() (err error)
  Transport() TTransport
}  
```

Binary序列化协议

基础类型编码

thrift传输的数据是message和struct，但是message和struct都是由基础类型组成的，因此需要先分析基础类型是如何序列化的。

**整形类型编码** 

整形是定长编码，采用网络字节序（大端模式），这里整形包括i8，i16，i32，i64，bool。bool会被转成i8进行存储。以i32为典型例子来分析整形的具体序列化过程。其他整形类型的序列化过程于此类似，只是字节大小不同而已。

> - 计算机电路先处理低位字节，效率比较高，因为计算都是从低位开始的。所以，计算机的内部处理都是小端字节序。

> - 人类还是习惯读写大端字节序。所以，除了计算机的内部处理，其他的场合几乎都是大端字节序，比如网络传输和文件储存。

```
// 序列化int32
func (p *TBinaryProtocol) WriteI32(value int32) error {
  // buffer是预先分配好的内存，[]byte类型，取slice的四个长度
  v := p.buffer[0:4]
  // 将value转成网络字节序，写入到v中
  binary.BigEndian.PutUint32(v, uint32(value))
  // 调用transport层的Write方法写数据
  _, e := p.writer.Write(v)
  return NewTProtocolException(e)
}
// 这里将int类型都转成了uint类型来处理，是因为负数在右移的时候会补1，转了无符号数来处理就可以很简洁的处理。
// 负数转uint也不会有什么问题，在反序列出来的时候，再将uint转成int时会溢出变成原来的负数
func (bigEndian) PutUint32(b []byte, v uint32) {
  _ = b[3] // early bounds check to guarantee safety of writes below
  b[0] = byte(v >> 24)
  b[1] = byte(v >> 16)
  b[2] = byte(v >> 8)
  b[3] = byte(v)
}
// 反序列化int32
func (p *TBinaryProtocol) ReadI32() (value int32, err error) {
  buf := p.buffer[0:4]
  // 调用transport层的接口，读且只读len(buf)个长度的字节数据
  err = p.readAll(buf)
  value = int32(binary.BigEndian.Uint32(buf))
  return value, err
}
```

| **类型** | **占用字节数** | **类型ID** |
| -------- | -------------- | ---------- |
| i8       | 1              |            |
| i16      | 2              | 6          |
| i32      | 4              | 8          |
| i64      | 8              | 10         |
| bool     | 1              | 2          |

**binary类型编码** 

binary 类型实际上就是 []byte，因为数组长度不固定，因此需要存储数组的长度，前面四个字节存储byte 数组的长度，后面存储 byte 数组的数据

```
Binary protocol, binary data, 4+ bytes:
+--------+--------+--------+--------+--------+...+--------+
| byte length            | bytes         |
+--------+--------+--------+--------+--------+...+--------+
func (p *TBinaryProtocol) WriteBinary(value []byte) error {
  e := p.WriteI32(int32(len(value)))
  if e != nil {
    return e
  }
  _, err := p.writer.Write(value)
  return NewTProtocolException(err)
}
```

**string类型编码** 

字符串会先使用UTF-8进行编码，然后转成binary类型进行发送

**double类型编码** 

**bool类型编码** 

bool类型会转成i8类型处理

Message编码

Message就是thrift协议每次请求需要带上的元信息，对message编码分为两种：严格模式和非严格模式。

严格模式

严格模式下需要占用12以上的字节，具体格式如下：

![img](https://bytedance.feishu.cn/space/api/box/stream/download/asynccode/?code=9554c42919c5f63a5ff7d3197791960d_8f118824ce50c961_boxcnXGmAxWjHfZFyjoZ95xKKdg_iL2emw82f9GHRaVkwoDD35aKTKXZSC2Y)

- 第一个字节中的第一位一定为1，用来区分是严格模式，非严格模式的第一位是0。前两个字节中，出去第一位剩下的15位表示版本号。也就是上图中的vvvvvvvvvvvvv。

- unused暂时没有使用

- mmm用来表示message type，一个无符号的3bit整形。前面5位必须是0。

| Message type | value |
| ------------ | ----- |
| call         | 001   |
| oneway       | 010   |
| reply        | 011   |
| exception    | 100   |

- Name length 是int32来表示name的长度，大端存储

- name 就是方法名，UTF-8编码

- Seq id 序列ID，是int32，大端编码

非严格模式

非严格模式需要占用9个以上的字节。每个字段的含义与严格模式一直。由于name length一定是非负整数，所以name length的第一位一定是0.

![img](https://bytedance.feishu.cn/space/api/box/stream/download/asynccode/?code=eda0e91e3737d0920c21f71140bd3276_8f118824ce50c961_boxcngOq00xPQbD8YQT4DfAbnCC_FemH6AXJXm5wblRD27FVIShlO2KbU6El)

Struct 编码

结构体每一个字段编码都会有一个前缀，一个字节表示字段的类型，两个字节表示字段的序列ID，就是IDL里定义的。最后整个结构体编码完成后，写入一个全0的字节表示结束。因为结构体第一个字段是类型，但是类型不可能为0，因此也就不会和结束符冲突。

```
Binary protocol field header and field value:
+--------+--------+--------+--------+...+--------+
|tttttttt| field id     | field value     |
+--------+--------+--------+--------+...+--------+
Binary protocol stop field:
+--------+
|00000000|
+--------+
```

- tttttttt是一个int8，存储类型ID

- Field id 是一个int16，存字段ID，这里说明字段ID可以为负。

- fieldValue 就存储具体的数据类型。如果字段继续是一个结构体，那么递归的解析就好了。

- 字段的名称不会被序列化进去，因此如果字段名称发生变更是没有任何影响的。

- 字段序列化的顺序也没有关系，因为是使用ID来识别的。

类型ID的定义如下：

- BOOL, encoded as 2

- BYTE, encoded as 3

- DOUBLE, encoded as 4

- I16, encoded as 6

- I32, encoded as 8

- I64, encoded as 10

- STRING, used for binary and string fields, encoded as 11

- STRUCT, used for structs and union fields, encoded as 12

- MAP, encoded as 13

- SET, encoded as 14

- LIST, encoded as 15

List 和 Set

list的set序列化的结构相同

```
Binary protocol list (5+ bytes) and elements:
+--------+--------+--------+--------+--------+--------+...+--------+
|tttttttt| size                | elements       |
+--------+--------+--------+--------+--------+--------+...+--------+
```

- tttttttt一个字节表示类型

- 四个字节表示元素个数和元素的类型

- elements就是具体的数据

> The element-type values are the same as field-types. The full list is included in the struct section above.

> The maximum list/set size is configurable. By default there is no limit (meaning the limit is the maximum int32 value: 2147483647).

Map

```
Binary protocol map (6+ bytes) and key value pairs:
+--------+--------+--------+--------+--------+--------+--------+...+--------+
|kkkkkkkk|vvvvvvvv| size                | key value pairs   |
+--------+--------+--------+--------+--------+--------+--------+...+--------+
```

- kkkkkkkk表示key类型，int8

- vvvvvvvv表示value类型，int8

- size是map的键值对个数，int32

- key value pairs 就是具体的数据

> The element-type values are the same as field-types. The full list is included in the struct section above.

> The maximum map size is configurable. By default there is no limit (meaning the limit is the maximum int32 value: 2147483647).

> Framed vs. unframed transport

> The first thrift binary wire format was unframed. This means that information is sent out in a single stream of bytes. With unframed transport the (generated) processors will read directly from the socket (though Apache Thrift does try to grab all available bytes from the socket in a buffer when it can).

> Later, Thrift introduced the framed transport.

> With framed transport the full request and response (the TMessage and the following struct) are first written to a buffer. Then when the struct is complete (transport method flush is hijacked for this), the length of the buffer is written to the socket first, followed by the buffered bytes. The combination is called a *frame*. On the receiver side the complete frame is first read in a buffer before the message is passed to a processor.

> The length prefix is a 4 byte signed int, send in network (big endian) order. The following must be true: 0 <= length <= 16384000 (16M).

> Framed transport was introduced to ease the implementation of async processors. An async processor is only invoked when all data is received. Unfortunately, framed transport is not ideal for large messages as the entire frame stays in memory until the message has been processed. In addition, the java implementation merges the incoming data to a single, growing byte array. Every time the byte array is full it needs to be copied to a new larger byte array.

> Framed and unframed transports are not compatible with each other.

Compact 序列化协议

> Compact序列化也是一种二进制的序列化，不同于Binary的点主要在于整数类型采用了zigzag 和 varint压缩编码实现，这里简要介绍下zigzag 和 varint整数编码。

> **VarInt**

> 对于一个整形数字，一般使用 4 个字节来表示一个整数值。但是经过研究发现，消息传递中大部分使用的整数值都是很小的非负整数，如果全部使用 4 个字节来表示一个整数会很浪费。比如数字1用四字节表示就是这样:

> 00000000 00000000 00000000 00000001

> 对较小整数来说，这种固定字节数编码很浪费bit。所以人们就发明了一个类型叫变长整数varint。数值非常小时，只需要使用一个字节来存储，数值稍微大一点可以使用 2 个字节，再大一点就是 3 个字节，它还可以超过 4 个字节用来表达长整形数字。

> 其原理也很简单，就是保留每个字节的最高位的bit来标识后一个字节是否属于该bit，1表示属于，0表示不属于。

示意图如下:

![img](https://bytedance.feishu.cn/space/api/box/stream/download/asynccode/?code=c1fe079cab201e6875915bdf033d4701_8f118824ce50c961_boxcnTfbbBi2kbj7JZhkpLSylSd_spNWW6NdxLDYkw0YKr2hDCZN8tCm4cHD)

由于大多数时候使用的是较小的整数，所以总体上来说，Varint编码的方式可以有效的压缩多字节整数。

那么对于负数怎么办呢?大家知道负数在计算机中是以补码的形式存在的。

```
10000000 00000000 00000000 00000001  -1的原码
11111111 11111111 11111111 11111110  -1的反码
11111111 11111111 11111111 11111111  -1的补码
```

所以-1在计算机中就是11111111 11111111 11111111,如果按照Varint编码，那么需要6个字节才能存的下，但是在现实生活中，-1却是个常用的整数。越大的负数越常见，编码需要的字节数越大，这显然是不能容忍的。为了解决这个问题, 就要使用ZigZag编码压缩技术了。

**ZigZag**

zigzag 编码专门用来解决负数编码问题。zigzag 编码将整数范围一一映射到自然数范围，然后再进行 varint 编码。

```
0 => 0
-1 => 1
1 => 2
-2 => 3
2 => 4
-3 => 5
3 => 6
```

zigzag 将负数编码成正奇数，正数编码成偶数。解码的时候遇到偶数直接除 2 就是原值，遇到奇数就加 1 除 2 再取负就是原值。

**Compact细节**

Compact编码——压缩的二进制。Compact序列化的实现大致逻辑和Binary序列化实现是一样的，就是将i16、i32、i64三种类型使用zigzag+varint编码实现，string、binary、map、list、set复合类型的长度只采用varint编码没使用zigzag，其他逻辑几乎一样。

> Comparing binary and compact protocol

> The binary protocol is fairly simple and therefore easy to process. The compact protocol needs less bytes to send the same data at the cost of additional processing. As bandwidth is usually the bottleneck, the compact protocol is almost always slightly faster.

实现

协议层Interface定义 

在thrift的 protocol.go 文件里可以发现，为协议定义了一套interface：

```
type TProtocol interface {
  WriteMessageBegin(name string, typeId TMessageType, seqid int32) error
  WriteMessageEnd() error
  WriteStructBegin(name string) error
  WriteStructEnd() error
  WriteFieldBegin(name string, typeId TType, id int16) error
  WriteFieldEnd() error
  WriteFieldStop() error
  WriteMapBegin(keyType TType, valueType TType, size int) error
  WriteMapEnd() error
  WriteListBegin(elemType TType, size int) error
  WriteListEnd() error
  WriteSetBegin(elemType TType, size int) error
  WriteSetEnd() error
  WriteBool(value bool) error
  WriteByte(value byte) error
  WriteI16(value int16) error
  WriteI32(value int32) error
  WriteI64(value int64) error
  WriteDouble(value float64) error
  WriteString(value string) error
  WriteBinary(value []byte) error
  ReadMessageBegin() (name string, typeId TMessageType, seqid int32, err error)
  ReadMessageEnd() error
  ReadStructBegin() (name string, err error)
  ReadStructEnd() error
  ReadFieldBegin() (name string, typeId TType, id int16, err error)
  ReadFieldEnd() error
  ReadMapBegin() (keyType TType, valueType TType, size int, err error)
  ReadMapEnd() error
  ReadListBegin() (elemType TType, size int, err error)
  ReadListEnd() error
  ReadSetBegin() (elemType TType, size int, err error)
  ReadSetEnd() error
  ReadBool() (value bool, err error)
  ReadByte() (value byte, err error)
  ReadI16() (value int16, err error)
  ReadI32() (value int32, err error)
  ReadI64() (value int64, err error)
  ReadDouble() (value float64, err error)
  ReadString() (value string, err error)
  ReadBinary() (value []byte, err error)
  Skip(fieldType TType) (err error)
  Flush() (err error)
  Transport() TTransport
}
```

那么不管使用什么协议binary，compact还是json，只要实现了这一套interface就可以完成序列化的工作了。

以binary协议为例

每一种结构进行序列化并不复杂，以i32为例

```
func (p *TBinaryProtocol) WriteI32(value int32) error {
  v := p.buffer[0:4]
  // 这里如果是个负数转uint32会溢出，但是ReadI32时从uint32强转回来时也会溢出，又回到原来的负数，所以这里就不会有影响了
  binary.BigEndian.PutUint32(v, uint32(value))
  _, e := p.writer.Write(v)
  return NewTProtocolException(e)
}
func (bigEndian) PutUint32(b []byte, v uint32) {
  _ = b[3] // early bounds check to guarantee safety of writes below
  // 计算机是使用小端存储，这里调换前后顺序，完成大端存储
  b[0] = byte(v >> 24)
  b[1] = byte(v >> 16)
  b[2] = byte(v >> 8)
  b[3] = byte(v)
}
```

​                

版本管理

结构体字段标识             

Thrift 在每一个结构体前用一个唯一的数字标识符来标识这个字段，这个字段比使用版本号的方式要强太多，这个标识符也会被序列化传输到服务端，服务端接受到数据后，就可以通过这个标识符来判断将改字段数据序列化到对应的字段。那么客户端和服务端和字段可以任意删除和添加，只要不修改标识符就行。

```
struct Example {
  1:i32 number=10,
  2:i64 bigNumber,
  3:double decimals,
  4:string name="thrifty"
}
```

同样这样的标识符也用来标识函数参数

```
service StringCache {
  void set(1:i32 key, 2:string value),
  string get(1:i32 key) throws (1:KeyNotFound knf),
  void delete(1:i32 key)                     
}                     
```

唯一的数字标识符其实与使用变量名称是一样的，只是使用数字标识符会减小序列化数据的大小，而且性能更高。

字段标识符建议手动指定，从1开始。如果不指定，会自动指定，自动指定从-1开始递减。避免和手动指定的值冲突。

case分析

1. 新增字段，旧client，新server。

参考             

1. [Thrift: The Missing Guide](https://diwakergupta.github.io/thrift-missing-guide/#_versioning_compatibility)

1. [Thrift: Scalable Cross-Language Services Implementation](http://thrift.apache.org/static/files/thrift-20070401.pdf)      

1. https://github.com/apache/thrift/blob/master/doc/specs/thrift-rpc.md    

1. https://andrewpqc.github.io/2019/02/24/thrift/         

​     