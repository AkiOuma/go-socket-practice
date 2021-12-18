# TCP Socket Practice

# 作业说明
* 总结几种 socket 粘包的解包方式：fix length/delimiter based/length field based frame decoder。尝试举例其应用。
	|问题|文档位置|代码位置|
	|-|-|-|
	|模拟粘包场景|[模拟粘包情景](#模拟粘包情景)|`situtaion/stickyPackage`|
	|fix length|[固定接受与发送端的缓冲区大小](#固定接受与发送端的缓冲区大小)|`situtaion/fixLength`|
	|delimiter based|[使用特定的分隔符](#使用特定的分隔符)|`situtaion/fixLength`|
	|length field based frame decoder|[在数据包中添加长度字段](#在数据包中添加长度字段)|`situtaion/lengthField`|
* 实现一个从socket connection中解码出`goim`协议的解码器
	|文档说明位置|代码位置|
	|-|-|
	|[模拟goim解码](#GOIM)|`goim-simulate`|


# 粘包

## 问题
总结几种 socket 粘包的解包方式：fix length/delimiter based/length field based frame decoder。尝试举例其应用。

## 定义
粘包问题是指当客户端发送多条消息时，比如发送了 ABC 和 DEF，但服务端接收到的却是 ABCDE，而不是分别收到ABC和DEF，像这种一次性读取了多条数据的情况就叫做粘包



## 产生原因
因为 TCP 是面向连接的传输协议，TCP 传输的数据是以流的形式，而流数据是没有明确的开始结尾边界，所以 TCP 也没办法判断哪一段流属于一个消息。基于这个原因，粘包主要是因为：
* 发送方每次写入数据小于Socket buffer的大小；
* 接收方读取socket buffer的数据不够及时



## 模拟粘包情景



服务端代码：`situtaion/stickyPackage/server/main.go`, 启动命令`make sti-ser`
* 开启一个`tcp`监听器，监听`0.0.0.0:8080`
* 接受到请求后直接把buffer的内容取出后转成字符串打印到控制台
```go
func handler(c net.Conn) {
	defer c.Close()
	buf := make([]byte, config.BufferSize)
	result := bytes.NewBuffer(nil)
	for {
		n, err := c.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println("ending: " + err.Error())
				return
			} else {
				log.Println("Read error: " + err.Error())
				break
			}
		}
		result.Write(buf[0:n])
		log.Printf("recevie size[%d]: %v", n, result.String())
		result.Reset()
	}
}

```



客户端代码：`situtaion/stickyPackage/client/main.go`, 启动命令`make sti-cli`

* 每秒向pipeline中写入两个hello信息，并带上信息的序号（如：`hello[0]`），持续5秒后关闭连接，可知每次写入pipeline的信息大小为8个字节
```go
func main() {
	after := time.After(5 * time.Second)
	iter := 0
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer conn.Close()

	// sending message
	for {
		select {
		case <-after:
			log.Println("time out")
			return
		default:
			for i := 0; i < 10; i++ {
				content := "hello[" + strconv.Itoa(iter) + "]"
				_, err = conn.Write([]byte(content))
				if err != nil {
					log.Fatal(err.Error())
				}
				iter++
			}
			time.Sleep(1 * time.Second)
		}
	}
}
```

运行结果：
* 服务端：
```shell
[root@workspace socket-practice]# make sti-ser
go run ./situtaion/stickyPackage/server/
2021/12/17 23:22:49 listeing from 0.0.0.0:8080
2021/12/17 23:22:54 recevie size[80]: hello[0]hello[1]hello[2]hello[3]hello[4]hello[5]hello[6]hello[7]hello[8]hello[9]
2021/12/17 23:22:55 recevie size[63]: hello[10]hello[11]hello[12]hello[13]hello[14]hello[15]hello[16]
2021/12/17 23:22:55 recevie size[27]: hello[17]hello[18]hello[19]
2021/12/17 23:22:56 recevie size[90]: hello[20]hello[21]hello[22]hello[23]hello[24]hello[25]hello[26]hello[27]hello[28]hello[29]
2021/12/17 23:22:57 recevie size[72]: hello[30]hello[31]hello[32]hello[33]hello[34]hello[35]hello[36]hello[37]
2021/12/17 23:22:57 recevie size[18]: hello[38]hello[39]
2021/12/17 23:22:58 recevie size[81]: hello[40]hello[41]hello[42]hello[43]hello[44]hello[45]hello[46]hello[47]hello[48]
2021/12/17 23:22:58 recevie size[9]: hello[49]
2021/12/17 23:22:59 ending: EOF
```

可以看到，服务端是一次把发送的多个包连在一起读取的（多个hello信息被连到了一起），而不是按照客户端发送一次信息然后读取一次信息的这种模式



# 粘包的解决方案



## 固定接受与发送端的缓冲区大小
固定接受与发送端的缓冲区大小，当发送端的字符长度不足时使用空字符补足



### 实现思路
* 服务端：`situtaion/fixLength/server/main.go`, 启动命令`make fix-ser`
  与粘包场景的服务端代码基本一致

* 客户端：`situtaion/fixLength/client/main.go`, 启动命令`make fix-cli`
  在粘包场景的客户端代码的基础上，新增了一个patch函数，用来填补长度不足时的空字符
```go
// patching empty byte into origin message
func patch(message string) []byte {
	res := make([]byte, fixlength.BufferSize)
	copy(res, []byte(message))
	return res
}

```



### 运行结果

* 服务端：
```shell
[root@workspace socket-practice]# make fix-ser
go run ./situtaion/fixLength/server/
2021/12/17 23:24:35 listeing from 0.0.0.0:8080
2021/12/17 23:24:37 recevie size[1024]: hello[0]
2021/12/17 23:24:37 recevie size[1024]: hello[1]
2021/12/17 23:24:37 recevie size[1024]: hello[2]
...省略中间的结果
2021/12/17 23:24:41 recevie size[1024]: hello[48]
2021/12/17 23:24:41 recevie size[1024]: hello[49]
2021/12/17 23:24:42 ending: EOF
```

可以看到，因为客户端发出的包的大小与服务端的buffer size的大小保持了一致，因此粘包的问题被解决了，达到了客户端每发送一次消息服务端能正确打印出该消息的效果



### 分析
虽然这种方式可以解决粘包和半包的问题，但这种固定缓冲区大小的方式增加了不必要的数据传输，因为这种方式当发送的数据比较小时会使用空字符来弥补，所以这种方式就大大的增加了网络传输的负担，所以它也不是最佳的解决方案。



## 使用特定的分隔符

发送端和接收端双方约定好的一个分隔符

在发送消息时，在消息的结尾带上约定好的分隔符

在接受消息时，当遇到分隔符，则按分隔符切割前面收到的消息



### 实现思路

* 服务端 `situtaion/delimeter/server/main.go`, 启动命令`make del-ser`

  在读取buffer的数据时，检查buffer中是否出现了特定的分隔符。若检测到分隔符，则

```go
func handler(c net.Conn) {
	defer c.Close()
	buf := make([]byte, config.BufferSize)
	result := bytes.NewBuffer(nil)
	for {
		n, err := c.Read(buf)
		if err != nil {
			// 省去错误处理代码
		}
		result.Write(buf[0:n])

		// 设置缓冲区的阅读指针
		var start int
		var end int
		for k, v := range result.Bytes() {
			// 当遇到约定好的分隔符时，将分隔符前的内容提取出来，然后移动内容的开头的指针
			if v == config.Delimeter {
				end = k
				log.Printf("recevie: %v", string(result.Bytes()[start:end]))
				// 移动开头指针
				start = end + 1
			}
		}
		result.Reset()
	}
}

```



* 客户端：`situtaion/delimeter/client/main.go`, 启动命令`make del-cli`

  定义一个方法，为内容的结尾追加一个约定好的分隔符

```go
func patchDelimeter(content string) []byte {
	data := []byte(content)
	data = append(data, config.Delimeter)
	return data
}
```





### 运行结果

* 服务端

```shell
[root@workspace socket-practice]# make del-ser
go run ./situtaion/delimeter/server/
2021/12/17 23:50:54 recevie: hello[0]
2021/12/17 23:50:54 recevie: hello[1]
2021/12/17 23:50:54 recevie: hello[2]
2021/12/17 23:50:54 recevie: hello[3]
2021/12/17 23:50:54 recevie: hello[4]
2021/12/17 23:50:54 recevie: hello[5]
2021/12/17 23:50:54 recevie: hello[6]
...省略中间的结果
2021/12/17 23:50:58 recevie: hello[48]
2021/12/17 23:50:58 recevie: hello[49]
2021/12/17 23:50:59 ending: EOF
```

可以看到，服务端接受的结果均能被正确切割



### 分析

使用分隔符的方式与上面的固定发送与接受端的缓冲区大小相比，解决了每次发送的package较大的问题，但是同时缺引入了一个新的问题，有可能在数据正文中包含了分隔符的内容，这时就会出现非预期的切割行为，导致没有正确获取预期的数据



## 在数据包中添加长度字段

发送端在待发的package的头部追加指定长度的信息，填充上该package的长度信息

接收端在接受到package时，先尝试获取该package的头部的长度信息，如果获取成功，则根据读取到的长度信息去读取接下来的buffer的该长度的数据，然后再对读取到的数据进行解析



### 实现思路

定义package的前四个字节储存package正文内容的长度

* 服务端`situtaion/lengthField/server/main.go`，启动命令`make len-ser`

  与上面的两个方法不同，尝试使用`bufio`中的`Reader`去直接读取`net.Conn`（实现了`io.Reader`接口）的内容。

  首先尝试从Reader中获取前四个字符，尝试解码，获取package大小

  若获取成功，读取包含header的该package的全部信息，打印header以外的信息


```go
func handler(c net.Conn) {
	defer c.Close()
	reader := bufio.NewReader(c)
	for {
		peek, err := reader.Peek(4)
		if err != nil {
			// 省略错误信息
			break
		}
		buffer := bytes.NewBuffer(peek)
		var size int32
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			log.Println(err)
		}
		if int32(reader.Buffered()) < size+4 {
			continue
		}
		data := make([]byte, size+4)
		if _, err := reader.Read(data); err != nil {
			log.Println(err.Error())
			continue
		}
		log.Printf("recevie : %v", string(data[4:]))
	}
}
```

  

* 客户端`situtaion/lengthField/client/main.go`，启动命令`make len-cli`

  实现一个package前追加含有长度信息的header的方法

```go
// add length field
func addLengthField(content []byte) ([]byte, error) {
	size := len(content)
	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.BigEndian, int32(size)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, content); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
```



### 运行结果

* 服务端

```shell
[root@workspace socket-practice]# make len-ser
go run ./situtaion/lengthField/server/
2021/12/18 00:31:28 listeing from 0.0.0.0:8080
2021/12/18 00:31:34 recevie : hello[0]
2021/12/18 00:31:34 recevie : hello[1]
2021/12/18 00:31:34 recevie : hello[2]
2021/12/18 00:31:34 recevie : hello[3]
2021/12/18 00:31:34 recevie : hello[4]
2021/12/18 00:31:34 recevie : hello[5]
2021/12/18 00:31:34 recevie : hello[6]
2021/12/18 00:31:34 recevie : hello[7]
...省略中间的结果
2021/12/18 00:31:38 recevie : hello[48]
2021/12/18 00:31:38 recevie : hello[49]
2021/12/18 00:31:39 ending.
```

可以看到服务端能正确切分出信息，消除了粘包的影响



### 分析

在数据包中添加长度字段的做法综合了固定缓冲区大小和分隔符两种做法的优点，既不会导致每个数据包过大，也不会产生数据正文可能出现分隔符导致切分错误的情况，是解决`tcp`粘包的比较推荐的做法





# GOIM

## 问题
实现一个从socket connection中解码出`goim`协议的解码器



## 协议结构

根据毛大的视频课中得知`goim`的数据包采取了以下格式

| Type             | Size              |
| ---------------- | ----------------- |
| Package Length   | 4 bytes           |
| Header Length    | 2 bytes           |
| Protocol Version | 2 bytes           |
| Operation        | 4 bytes           |
| Sequence Id      | 4 bytes           |
| Body             | PackLen-HeaderLen |



## 实现思路

实现一个编码数据包和解码数据包的方法`goim-simulate/pkg/simulate.go`

* 根据协议定义一个数据包的对象

```go
// define package
type Pack struct {
	Length          int
	HeaderLength    int
	ProtocolVersion int
	OperationCode   int
	Seq             int
	Content         []byte
}
```



* 编码数据包方法

  定义一个字节切片，根据协议中不同段的信息的大小，使用`binary.BigEndian.PutUint16`（2字节区域）或`binary.BigEndian.PutUint32`（4字节区域）为字节切片插入数据包结构体中对应的内容

  最后将正文内容使用copy方法直接拷贝到字节切片的请求头之后的区域之中

```go
func Encoder(pack *Pack) []byte {
	res := make([]byte, pack.Length)

	// set package length
	binary.BigEndian.PutUint32(
		res[:headerLengthStart()],
		uint32(pack.Length),
	)
	// set header length
	binary.BigEndian.PutUint16(
		res[headerLengthStart():protocolVersionStart()],
		uint16(headerSize()),
	)
	// set protocol version
	binary.BigEndian.PutUint16(
		res[protocolVersionStart():operationCodeStart()],
		uint16(pack.ProtocolVersion),
	)
	// set operation code
	binary.BigEndian.PutUint32(
		res[operationCodeStart():sequenceIdStart()],
		uint32(pack.OperationCode),
	)
	// set sequence id
	binary.BigEndian.PutUint32(
		res[sequenceIdStart():sequenceIdStart()+sequenceIdFieldSize],
		uint32(pack.Seq),
	)
	// set body
	copy(res[headerSize():], pack.Content)

	return res
}
```



* 解码数据包方法

  首先判断该数据包是否大于协议中的请求头的大小（16字节），若小于该大小则放弃本次解码

  根据协议在数据包中使用`binary.BigEndian.Uint32`或者`binary.BigEndian.Uint16`取出相应的内容去还原Pack结构体的内容

```go
func Decoder(msg []byte) (*Pack, error) {
	if len(msg) < headerSize()+1 {
		return nil, ErrPackInComplete
	}
	// get package length
	packageLength := binary.BigEndian.Uint32(msg[:headerLengthStart()])
	// get header length
	headerLength := binary.BigEndian.Uint16(msg[headerLengthStart():protocolVersionStart()])
	// get protocol version
	protocolVersion := binary.BigEndian.Uint16(msg[protocolVersionStart():operationCodeStart()])
	// get operation code
	operationCode := binary.BigEndian.Uint32(msg[operationCodeStart():sequenceIdStart()])
	// get sequence id
	sequenceId := binary.BigEndian.Uint32(msg[sequenceIdStart() : sequenceIdStart()+sequenceIdFieldSize])
	// get data
	content := msg[headerSize():]
	return &Pack{
		Length:          int(packageLength),
		HeaderLength:    int(headerLength),
		ProtocolVersion: int(protocolVersion),
		OperationCode:   int(operationCode),
		Seq:             int(sequenceId),
		Content:         content,
	}, nil
}

```



* 服务端实现`goim-simulate/server/main.go`，启动命令`make im-ser`

  实现思路与添加长度字段的拆包方法类似，我们先尝试读取包长度的字段，若获取成功，则根据获取到的长度读取出该数据包，然后调用上面定义的Decode方法还原回Pack结构体

  

```go
func handler(c net.Conn) {
	defer c.Close()
	reader := bufio.NewReader(c)
	for {
		peek, err := reader.Peek(pkg.PackageLengthSize())
		if err != nil {
			// 省略错误处理
			break
		}
        // 获取数据包header部分的长度信息
		buffer := bytes.NewBuffer(peek)
		var size int32
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			log.Println(err)
		}
		if int32(reader.Buffered()) < size {
			continue
		}
        // 根据获取到的长度读取数据包
		data := make([]byte, size)
		if _, err := reader.Read(data); err != nil {
			log.Println(err.Error())
			continue
		}
        // 对数据包进行解码,还原回pkg.Pack对象
		content, err := pkg.Decoder(data)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		log.Println(string(content.Content))
	}
}

```





* 客户端实现`goim-simulate/client/main.go`，启动命令`make im-cli`

  使用上面定义的Encode方法去封装想要发送的数据正文即可

```go
func main() {
	seq := 0
	version := 1
	code := 1
	timeout := time.After(5 * time.Second)
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer conn.Close()
	for {
		select {
		case <-timeout:
			log.Println("timeout")
			return
		default:
			for i := 0; i < 5; i++ {
				data := pkg.Encoder(pkg.NewPack(version, code, seq, []byte("hello"+strconv.Itoa(seq))))
				seq++
				conn.Write(data)
			}
			time.Sleep(time.Second)
		}
	}
}
```



## 运行结果

* 客户端

```shell
[root@workspace socket-practice]# make im-ser
go run ./goim-simulate/server/
2021/12/18 01:13:32 hello0
2021/12/18 01:13:32 hello1
2021/12/18 01:13:32 hello2
2021/12/18 01:13:32 hello3
...省略中间的结果
2021/12/18 01:13:36 hello22
2021/12/18 01:13:36 hello23
2021/12/18 01:13:36 hello24
2021/12/18 01:13:37 ending...
2021/12/18 01:13:37 
```

可以看到，服务端可以正确拆分客户端发来的模拟`goim`协议的数据包