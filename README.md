# go-coroutine-learning
GO 语言协程知识点学习笔记

#### 协程？

协程并不是 GO 语言特有的机制，像 Lua、Ruby、Python、Kotlin、C/C++ 也都有协程的支持。区别在于有些是从语言层面支持，有的通过插件类库支持。Go 语言的协程是原生语言层面支持的。

#### 协程和进程和线程的对比

##### 进程

* 进程是系统资源分配的最小单位。进程的创建和销毁都是系统资源级别。操作起来代价昂贵。程是抢占式调度，分为三个状态：等待态、就绪态、运行态。进程之间是相互隔离的，拥有各自的系统资源，更加安全。因此存在进程之间通讯不方便的问题

* 进程是线程的载体容器，多个线程除了共享进程的资源还拥有自己的一少部分独立的资源。因此相比进程而言线程更加清凉，进程内的多个线程间的通信比进程容易，但也同样带来了同步和互斥的问题和线程安全的问题（多线程编程的常见问题），尽管多线程编程仍然是当前服务端编程的主流，线程也是 CPU 调度的最小单位，多线程运行时存在线程切换的问题，其状态的转移如图：

![状态的转移](https://github.com/polichan/go-coroutine-learning/blob/main/assets/17012e958099cbf4.jpg?raw=true)

##### 协程

* 协程在有的资料中成为微线程或者用户态轻量级线程，协程调度不需要系统内核参与而是完全由用户态程序来决定，因此协程对于系统而言是无感知的。协程由用户态控制就不存在抢占式调度。抢占式调度会强制让 CPU 控制权切换到其它进线程，多个协程进行协作实调度，协程自己主动把控制权转让出去后，其它协程 才能被执行到，这样就避免了系统切换开销提高了 CPU 的使用效率。

##### 抢占式调度和协作式调度的简单对比图：

![抢占式调度和协作式调度的简单对比图](https://github.com/polichan/go-coroutine-learning/blob/main/assets/17012e95811df2bb.jpg?raw=true)

#### Go 协程

启用一个 Go 协程非常简单，只需要在函数前加上关键字 ``go``，就可以轻易的启用一个新的 Go 协程并发运行。

```golang
package main

import "fmt"

func main() {
	go hello()
	fmt.Println("main function")
}

func hello()  {
	fmt.Println("Hello goroutine")
}
```

 ``go hello()`` 此处启动了一个协程，因此``hello()`` 函数会与 ``main()`` 函数并发执行。``main()`` 函数会运行在一个特有的 Go 协程上，它被称为 Go 的主协程（Main Goroutine)。

 运行结果大概率如下：

```shell
main function

Process finished with exit code 0
```

为什么只输出了 ``main function`` ？我们的 ``Hello goroutine`` 为什么没有输出呢？

* 当启动一个新的协程时，协程的调用会立即返回，与函数不同，程序控制不回去等待 Go 的协程执行完毕，在本例中即 ``hello()`` 函数，因此程序控制会立即返回到代码的下一行，即 ``fmt.Println("main function")`` ，忽略该协程的任何返回值
* 协程的存在与否取决于 ``main()`` 主协程是否存活，如果主协程被销毁（即程序终止了），那么其余用 go 关键字启用的协程，在本例中为 ``hello()`` 协程也被销毁了。这就导致了主协程已经节数，但没有完全等待 ``hello()`` 协程的执行完毕，就被销毁了。因此出现了以上情况。

看到这里，你应该知道为什么运行结果是大概率如下，因为主协程和协程的执行顺序是不一致的，在一定概率下，即主协程正好等到了协程的执行完毕，会输出 ``Hello goroutine``。

让我们来修复下这个问题吧！

```golang
package main

import (
	"fmt"
	"time"
)

func main() {
	go hello()
	time.Sleep(1 * time.Second)
	fmt.Println("main function")
}

func hello()  {
	fmt.Println("Hello goroutine")
}
```

可以看到，我们利用 ``time`` 包中的  ``Sleep`` 方法，去休眠主协程，让其等待 1 秒钟，等待 ``hello()`` 协程执行完毕。那么最终输出结果如下：

```shell
Hello goroutine
main function

Process finished with exit code 0
```
``注意``：在实际业务中，利用协程休眠的方式去控制协程的执行顺序是不可取的，因为这样的预测永远无法 100% 如你所愿。如果需要控制协程的同步，我们会在下面讲到 ``channel`` 即管道或信道。

##### 启用多个 Go 协程
为了更好的理解 Go 协程，我们再编写一个程序，启动多个 Go 协程

```golang
package main

import (
	"fmt"
	"time"
)

func main() {
	go hello()
	go anotherHello()
	time.Sleep(1 * time.Second)
	fmt.Println("main function")
}

func hello()  {
	for i := 0; i <= 3; i++ {
		time.Sleep(250 * time.Millisecond)
		fmt.Println("Hello goroutine")
	}
}

func anotherHello()  {
	for i := 0; i <= 3; i++ {
		time.Sleep(400 * time.Millisecond)
		fmt.Println("Hello another goroutine")
	}
}
```

* ``hello()`` 协程会在第一次输出 ``Hello goroutine`` 后，休眠该协程 250 毫秒，再次输出 ``Hello goroutine`` ，直到 ``i > 3``。
* ``anotherHello()`` 协程会在第一次输出 ``Hello another goroutine`` 后，休眠该协程 400 毫秒，再次输出 ``Hello another goroutine`` ，直到 ``i > 3``。
* 最终主协程会在该两个协程执行完毕后，输出  ``main function`` 

是不是很简单？Go 语言屏蔽了多线程的复杂实现，只需要一个 go 关键字即可轻而易举的创建一个新的协程。

##### 什么是 Channel 管道：

Channel 是 Go 协程之间通信的通信管道，如同管道中的谁会从一端流到另一端，通过使用管道，就可以实现数据从管道的一个端口发送，在另外一个端口接收。

##### 管道声明（无缓冲管道）

声明管道就像声明一个切片一样简单：

```golang
package main

func main() {
	c := make(chan int)
	// 或者
	var c2 chan int
}
```

* ``chan T`` 表示该管道只能运送 ``T ``类型的数据，在本例中为 ``int`` 类型数据

##### 通过管道进行发送与接收

```golang
a := make(chan int)
getDataFromChannelA := <- a // 读取管道 a 发送的内容
a <- 1 // 向管道 a 发送 1 
```
怎么样？是不是很简单。在一开始可能会觉得管道操作符比较难记，其实我们只需要记住简单的方法即可

* `` a <- 1 `` 向管道 a 发送 1，箭头指向管道，即代表发送
* `` <- a `` 从管道 a 读取数据，箭头不指向管道，即代表从管道接收

##### 发送与接收默认是阻塞的

发送于接收默认是阻塞的，这是什么意思呢？当把数据发送至管道时，程序控制会产生阻塞，直到有另外一个协程（可以是主协程）来进行接收，直到其它协程接收了来自管道的数据，反之亦然。否则程序才能继续运行，否则将一直阻塞。

利用该特性，我们可以很容易实现一个不用加锁的同步的协程通信的管道。那么我们如果实现一个发送与接受不阻塞的管道呢？其实只需要将管道增加缓冲即可，详见下面。在这之前，我们先来利用管道进行协程通讯。

```golang
package main

import "fmt"

func main() {
	c := make(chan bool)
	go useChannelToCommunicate(c)
	<- c
	fmt.Println("main function")
}

func useChannelToCommunicate(c chan bool)  {
	fmt.Println("Hello goroutine")
	c <- true
}
```

* 我们首先声明了一个可以存放 ``bool`` 类型的管道。
* 我们将该管道作为参数传递给 ``useChannelToCommunicate`` 函数
* 在 ``useChannelToCommunicate`` 函数中，当执行完打印 ``Hello goroutine`` 时，会向管道 ``c`` 发送 ``true``，即告知主协程，打印已经执行完毕，你可以不用阻塞了。
* 主协程从管道 ``c`` 中接收到了传递的布尔值，主协程继续执行，打印出 ``main function``
* 值得注意的是，在主协程中，我们并没有去读取管道 ``c `` 传递的数据赋值给某个变量，其实这是完全合法的，我们只是利用了管道的阻塞特性，没有必要去特意赋值给某个变量。

怎么样？相较于笨笨地用休眠的方式来控制协程的同步，利用管道是不是更加优雅，并且更加健壮了？

##### 死锁（DeadLock）
在我们使用管道时，我们必须要要考虑到一个问题，那就是死锁问题。如果一个管道只有接收，没有发送，那么就会导致程序无限制的等待，因此为了解决这个问题，我们需要时刻注意是否会出现死锁情况，否则会出现 ``panic``，以下例子会产生一个协程的死锁。

```golang
func main() {
	c := make(chan bool)
	c <- true
	fmt.Println("main function")
}
```
在本例中，管道 ``c`` 一直在发送 ``true`` 至管道，但显然并没有任何一个协程进行接收，也就导致了程序无线等待，产生死锁问题。此时编译器会报告代码存在死锁问题。


```shell
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
main.main()
	D:/Projects/go-coroutine-learning/main.go:7 +0x5f

Process finished with exit code 2
```

##### 单向管道
上面所提到的例子中，展示的都是双向管道，即有接收也有发送，那么我们是否可以创建一个单向管道呢，这种管道只能发送或者接收数据。其实是可以的。

```golang
package main

import "fmt"

func main() {
	c := make(chan<- bool)
	go send(c)
	fmt.Println(<-c)
}

func send(c chan<- bool)  {
	c <- true
}
```
* 利用 ``make(chan<- bool)`` 我们声明了一个只发送的管道，因为箭头指向了管道，在主协程中，我们试图用 ``<-c`` 进行管道接收。显然是不行的，并且编译器提示我们一个只发送的管道不能进行接收。


```shell
.\main.go:8:14: invalid operation: <-c (receive from send-only type chan<- bool)

Compilation finished with exit code 2
```
可是这样有什么意义呢？这就需要使用到管道转换（Channel Conversion），将一个双向管道（即发送与接收）转换成只发送或只接收。值得一提的是，我们不能将只发送或只接收管道转换成一个双向管道。

```golang
package main

import "fmt"

func main() {
	c := make(chan bool)
	go send(c)
	fmt.Println(<-c)
}

func send(c chan<- bool)  {
	c <- true
}
```

* 首先我们创建了一个双向管道，并将其作为参数传递给 ``send``，在 ``send`` 函数的参数中，我们定义了只允许只发送管道，因此 Go 会将其进行一次转换，转换为只发送的管道。
* ``send`` 函数通过管道 ``c`` 发送 ``true`` 至管道
* 在主协程中，我们的管道是双向的，因此我们可以进行接收，将管道中的值给打印出来，因此控制台输出 ``true``

##### 关闭管道以及遍历管道数据
我们定义数据的发送方可以关闭管道，通知接收方该管道已关闭，不再有数据传递过来了。因此接收方可以多用一个变量来检查管道时候已经关闭。

###### 关闭管道
```golang
c := make(chan int)
closed(c)
value, isClosed := <- c
```
* ``close`` 函数代表关闭一个管道
* ``isClosed`` 代表管道是否已经关闭，如果已经关闭，那么 ``value`` 值为该类型的零值，在这里是 0
* ``value`` 代表从管道中接收到的值，如果管道未关闭，那么该值就是从管道接收到的值。

###### 遍历管道数据
```golang
package main

import "fmt"

func main() {
	c := make(chan int)
	go producer(c)
	for {
		value, isClosed := <- c
		if isClosed == false {
			fmt.Println("Channel has been closed!")
			break
		}
		fmt.Println("Value:", value)
	}
}

func producer(c chan int)  {
	for i := 0; i <= 10; i++{
		c <- i
	}
	close(c)
}
```

我们利用 ``isClosed`` 来判断管道是否已经被关闭了，如果被关闭了，那么我们会 ``break`` 出这个无限循环。

###### 利用 ``for range`` 遍历管道数据

``for range`` 循环用于在一个管道关闭之前，从管道接收数据。因此我们可以改写上例的代码如下：

```golang
package main

import "fmt"

func main() {
	c := make(chan int)
	go producer(c)
	for value := range c{
		fmt.Println("value: ", value)
	}
	fmt.Println("Channel has been closed!")
}

func producer(c chan int)  {
	for i := 0; i <= 10; i++{
		c <- i
	}
	close(c)
}
```

##### 缓冲管道

在无缓冲管道中，我们提到了无缓冲管道是阻塞的，因此本小节来介绍带有缓冲的管道，即可以实现无阻塞。

我们只需要在创建管道时，给予管道一个容量大小，就可以声明一个带有缓冲的管道。

```golang
ch := make(chan int, 10)
```

以上声明了一个允许传输 ``int`` 类型的缓冲管道，并且管道的容量大小为 10。

```golang
package main

import (
	"fmt"
	"time"
)

func main() {

	ch := make(chan int, 2)
	go write(ch)
	time.Sleep(2 * time.Second)
	for v := range ch {
		fmt.Println("read value", v, "from ch")
		time.Sleep(2 * time.Second)
	}
}

func write(ch chan int){
	for i := 0; i < 5; i ++{
		ch <- i
		fmt.Println("Success wrote:", i, "to ch")
	}
	close(ch)
}

```

首先我们创建了一个容量大小为 2 的缓冲管道。并且启动了一个协程，此协程不断得向 ``ch`` 传输数据，但是因为我们的容量为 2 ，因此一开始只会写入 ``0``和 ``1`` ，随后主协程进行休眠，接收管道数据，每一次接收就会使容量减少 1，因此协程又会进行写入，直到协程关闭了 ``ch`` 管道。

```shell
Success wrote: 0 to ch
Success wrote: 1 to ch
read value 0 from ch
Success wrote: 2 to ch
read value 1 from ch
Success wrote: 3 to ch
read value 2 from ch
Success wrote: 4 to ch
read value 3 from ch
read value 4 from ch

Process finished with exit code 0
```

##### 死锁

```golang
package main

import (  
    "fmt"
)

func main() {  
    ch := make(chan string, 2)
    ch <- "naveen"
    ch <- "paul"
    ch <- "steve"
    fmt.Println(<-ch)
    fmt.Println(<-ch)
}
```

首先，我们创建了一个容量为 2 的管道，并向其发送了三个数据，分别是 ``naveen``、``paul``、``steve``，但是我们在主协程并没有进行接收，导致只有发送，没有接受，因此产生了死锁。

##### 管道的长度与容量

* 长度是指管道中当前排队的元素（数据）个数
* 容量是指管道可以存储的元素（数据）的数量

例如，容量为 3 的管道可能当前排队了两个元素，排队了 2 个元素的管道容量一定是大于等于 2 的。

#### 参考资料：

[浅谈协程和Go语言的Goroutine](https://juejin.cn/post/6844904056918376456)

[Go 系列教程 —— 21. Go 协程](https://studygolang.com/articles/12342)

[Go 系列教程 —— 22. 信道（channel)](https://studygolang.com/articles/12402)