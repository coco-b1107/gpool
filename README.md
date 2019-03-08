# gpool

Go语言实现，用于快速构建资源池的库

## 使用方法

### 创建池元素struct并实现Item接口方法

需要实现的方法签名列表

```go
Initial(map[string]string) error
Destory() error
Check() error
```

#### 示例

```go
//Connection 连接池对象
type Connection struct {
	TCPConn net.Conn
}

//Initial 初始化
func (c *Connection) Initial(params map[string]string) error {
	con, err := net.Dial("tcp", params["host"]+":"+params["port"])
	if err != nil {
		return err
	}
	c.TCPConn = con
	return nil
}

//Destory 销毁连接
func (c *Connection) Destory() error {
	return c.TCPConn.Close()
}

//Check 检查元素连接是否可用
func (c *Connection) Check() error {
	fmt.Println("检查连接可用")
	return nil
}
```

### 实现创建池元素工厂方法

```go
//NewConnection 获取新连接
func NewConnection() gpool.Item {
	return &Connection{}
}
```

### 创建单例模式的Pool

```go
var (
	pool *gpool.Pool
	once sync.Once
)

func init() {
	once.Do(func() {
		pool = gpool.DefaultPool()
		pool.Config.LoadToml("general.toml")

		fmt.Println(pool.Config)
		pool.NewFunc = NewConnection
		pool.Initial()

	})
}
```

### 实现获取池元素与归还池元素的方法

```go
//GetConnection 获取连接
func GetConnection() (net.Conn, error) {
	item, err := pool.GetOne()
	if err != nil {
		return nil, err
	}
	con, ok := item.(*Connection)
	if ok {
		return con.TCPConn, nil
	}
	return nil, errors.New("类型转换错误")
}

//CloseConnection 关闭连接
func CloseConnection(conn net.Conn) {
	pool.BackOne(&Connection{
		TCPConn: conn,
	})
}
```

### 实现关闭池方法

```
//ClosePool 关闭连接池
func ClosePool() {
	pool.Shutdown()
}
```

### 使用

略

## 配置说明

| 名称                 | 说明                                                     | 类型              | 默认值 |
| -------------------- | -------------------------------------------------------- | ----------------- | ------ |
| InitialPoolSize      | 初始化池中元素数量，取值应在MinPoolSize与MaxPoolSize之间 | int               | 5      |
| MinPoolSize          | 池中保留的最小元素数量                                   | int               | 2      |
| MaxPoolSize          | 池中保留的最大连元素数量                                 | int               | 15     |
| AcquireRetryAttempts | 定义在新连接失败后重复尝试的次数                         | int               | 5      |
| AcquireIncrement     | 当池中的元素耗尽时，一次同时创建的元素数                 | int               | 5      |
| TestDuration         | 连接有效性检查间隔，单位毫秒                             | int               | 1000   |
| TestOnGetItem        | 如果设为true那么在取得元素的同时将校验元素的有效性       | bool              | false  |
| Debug                | 显示调试信息                                             | bool              | false  |
| Params               | 元素初始化参数                                           | map[string]string |        |

## 示例

下面用一个示例展示这个库的用法

### 目录结构

```
.
|-- cmd
|   |-- client
|   |   `-- client.go
|   `-- server
|       `-- server.go
|-- dial
|   `-- dial.go
|-- general.toml
|-- go.mod
|-- go.sum
`-- gpool
    |-- config.go
    `-- gpool.go
```

**cmd/client/client.go**

```
package main

import (
	"app/general/dial"
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {
	wg := sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				send(i, j)
			}

		}(i)
	}
	wg.Wait()
	dial.ClosePool()
	end := time.Now()
	fmt.Println(end.Sub(start))
}

func send(i, j int) {
	time.Sleep(time.Second)
	conn, err := dial.GetConnection()
	if err != nil {
		log.Fatalf("第%d个线程获取连接失败%v", i, err)
	}
	defer dial.CloseConnection(conn)
	_, err = conn.Write([]byte(fmt.Sprintf("%d %d\n", i, j)))
	if err != nil {
		log.Fatal(err)
	}
}
```
**cmd/server/server.go**

```
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:11211")
	if err != nil {
		log.Fatal(err)
	}
	i := 0
	for {
		conn, err := listener.Accept()
		i++
		if err != nil {
			log.Fatal(err)
		}
		go Proc(conn, i)
	}

}

//Proc Proc
func Proc(conn net.Conn, i int) {
	defer conn.Close()
	buf := bufio.NewReader(conn)
	for {
		bs, _, err := buf.ReadLine()
		if err != nil {
			return
		}
		v := string(bs)
		slist := strings.Split(v, " ")

		if len(slist) == 2 {
			if slist[1] == "999" {
				fmt.Printf("第%d个连接收到消息 : 线程 %s 第 %s 次发送\n", i, slist[0], slist[1])
			}
		} else {
			fmt.Println(v)
		}
	}
}
```

**dial/dial.go**

```
package dial

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/cloudfstrife/gpool"
)

var (
	pool *gpool.Pool
	once sync.Once
)

func init() {
	once.Do(func() {
		pool = gpool.DefaultPool()
		pool.Config.LoadToml("general.toml")

		fmt.Println(pool.Config)
		pool.NewFunc = NewConnection
		pool.Initial()

	})
}

//NewConnection 获取新连接
func NewConnection() gpool.Item {
	return &Connection{}
}

//Connection 连接池对象
type Connection struct {
	TCPConn net.Conn
}

//Initial 初始化
func (c *Connection) Initial(params map[string]string) error {
	con, err := net.Dial("tcp", params["host"]+":"+params["port"])
	if err != nil {
		return err
	}
	c.TCPConn = con
	return nil
}

//Destory 销毁连接
func (c *Connection) Destory() error {
	return c.TCPConn.Close()
}

//Check 检查元素连接是否可用
func (c *Connection) Check() error {
	fmt.Println("检查连接可用")
	return nil
}

//GetConnection 获取连接
func GetConnection() (net.Conn, error) {
	item, err := pool.GetOne()
	if err != nil {
		return nil, err
	}
	con, ok := item.(*Connection)
	if ok {
		return con.TCPConn, nil
	}
	return nil, errors.New("类型转换错误")
}

//CloseConnection 关闭连接
func CloseConnection(conn net.Conn) {
	pool.BackOne(&Connection{
		TCPConn: conn,
	})
}

//ClosePool 关闭连接池
func ClosePool() {
	pool.Shutdown()
}
```

**general.toml**

```
InitialPoolSize = 5
MinPoolSize = 2
MaxPoolSize = 15
AcquireRetryAttempts = 5
AcquireIncrement = 5
TestDuration = 1000
TestOnGetItem = false
Debug = false

[Params]
  host = "127.0.0.1"
  port = "11211"
```

**go.mod**

```
module app/general

go 1.12

require github.com/BurntSushi/toml latest

require github.com/cloudfstrife/gpool latest
```

### 运行与测试

```
go build app/general/cmd/server
./server

go build  app/general/cmd/client
./client
```
