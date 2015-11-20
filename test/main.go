// main.go
package main

import (
	// "fmt"
	"sync"
	"time"
)

type Session struct {
	r ReadSilk
	w WriteSilk
}

// 发送消息给唯一go程
type ReadSilk struct {
	// 从网络接口接收数据
	// 发送数据给合适的go线程->Actor模式(邮箱)
	// 邮箱在哪里? ReadSilk决定还是网络端口决定?
	// 由ReadSilk决定更能解耦网络端口
}

// 接受任何一个go程的消息
type WriteSilk struct {
	writeMutex sync.RWMutex
	writeCond  *sync.Cond
	data       int
}

func (this *WriteSilk) Init() {
	this.writeCond = sync.NewCond(&this.writeMutex)
}

func (this *WriteSilk) PostData(d int) {
	this.writeCond.L.Lock()

	this.data = d

	this.writeCond.Signal()
	this.writeCond.L.Unlock()

	// 写入消息列表
	// 一般较少调用
}

func (this *WriteSilk) Run() {
	go func() {
		for {
			this.writeCond.L.Lock()
			this.writeCond.Wait()

			println(this.data)

			// this.writeCond.Signal()
			this.writeCond.L.Unlock()
		}
	}()
}

const (
	SilkCount = 1000
)

func main() {
	xw := make([]WriteSilk, SilkCount)
	for i := 0; i < SilkCount; i++ {
		xw[i].Init()
		xw[i].Run()
	}

	for i := 0; i < 100; i++ {
		time.Sleep(2 * time.Second)

		for k := 0; k < SilkCount; k++ {
			xw[k].PostData(i)
		}
	}
}
