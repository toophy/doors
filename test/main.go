// main.go
package main

import (
	// "fmt"
	"github.com/toophy/doors/help"
	"sync"
	"time"
)

var G_thread_msg_pool ThreadMsgPool

// Go程间消息存放处
type ThreadMsgPool struct {
	lock       []sync.Mutex     // silk消息池有一个独立的锁
	cond       []sync.Cond      // 条件锁
	header     []help.DListNode // silk消息池
	lock_count uint             // Mutex锁邮箱
	cond_count uint             // Cond锁邮箱
}

// 初始化
func (this *ThreadMsgPool) Init(lock_count, cond_count uint) {
	this.lock = make([]sync.Mutex, lock_count+cond_count)
	this.cond = make([]sync.Cond, lock_count+cond_count)
	this.header = make([]help.DListNode, lock_count+cond_count)
	this.lock_count = lock_count
	this.cond_count = cond_count

	for i := uint(0); i < lock_count+cond_count; i++ {
		this.header[i].Init(nil)
	}
	for i := uint(0); i < lock_count+cond_count; i++ {
		this.cond[i].L = &this.lock[i]
	}
}

// 投递Go程间消息, PostMsg和GetMsg是一对
func (this *ThreadMsgPool) PostMsg(tid uint, e *help.DListNode) bool {
	if e != nil && !e.IsEmpty() && tid < this.lock_count+this.cond_count {
		this.lock[tid].Lock()
		defer this.lock[tid].Unlock()

		header := &this.header[tid]

		e_pre := e.Pre
		e_next := e.Next

		e.Init(nil)

		old_pre := header.Pre

		header.Pre = e_pre
		e_pre.Next = header

		e_next.Pre = old_pre
		old_pre.Next = e_next

		return true
	}
	return false
}

// 获取Go程间消息, PostMsg和GetMsg是一对
func (this *ThreadMsgPool) GetMsg(tid uint, e *help.DListNode) bool {
	if e != nil && tid < this.lock_count+this.cond_count {
		this.lock[tid].Lock()
		defer this.lock[tid].Unlock()

		header := &this.header[tid]

		if !header.IsEmpty() {
			header_pre := header.Pre
			header_next := header.Next

			header.Init(nil)

			old_pre := e.Pre

			e.Pre = header_pre
			header_pre.Next = e

			header_next.Pre = old_pre
			old_pre.Next = header_next
		}

		return true
	}
	return false
}

// 投递Go程间消息, PushMsg 和 WaitMsg 是一对, 投递和等待消息
func (this *ThreadMsgPool) PushMsg(tid uint, e *help.DListNode) bool {

	if e != nil && !e.IsEmpty() && tid < this.lock_count+this.cond_count {
		this.lock[tid].Lock()
		defer this.lock[tid].Unlock()

		println("PushMsg e")

		header := &this.header[tid]

		e_pre := e.Pre
		e_next := e.Next

		e.Init(nil)

		old_pre := header.Pre

		header.Pre = e_pre
		e_pre.Next = header

		e_next.Pre = old_pre
		old_pre.Next = e_next

		this.cond[tid].Signal()

		println("Push Msg : ", tid)

		return true
	}
	if e == nil {
		println("e null")
	}
	if e.IsEmpty() {
		println("e empty")
	}
	if tid >= this.lock_count+this.cond_count {
		println("tid out")
	}
	return false
}

// 等待Go程间消息, PushMsg 和 WaitMsg 是一对, 投递和等待消息
func (this *ThreadMsgPool) WaitMsg(tid uint, e *help.DListNode) bool {
	if e != nil && tid >= 0 && tid < this.lock_count+this.cond_count {
		this.lock[tid].Lock()
		defer this.lock[tid].Unlock()

		this.cond[tid].Wait()

		header := &this.header[tid]

		if !header.IsEmpty() {
			header_pre := header.Pre
			header_next := header.Next

			header.Init(nil)

			old_pre := e.Pre

			e.Pre = header_pre
			header_pre.Next = e

			header_next.Pre = old_pre
			old_pre.Next = header_next
		}

		return true
	}
	return false
}

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
	SildId uint
}

func (this *WriteSilk) Init(id uint) {
	this.SildId = id
}

type MsgX struct {
	Data int
}

func (this *WriteSilk) Run() {
	go func() {
		for {
			header := help.DListNode{}
			header.Init(nil)

			G_thread_msg_pool.WaitMsg(this.SildId, &header)

			for {
				// 每次得到链表第一个事件(非)
				n := header.Next
				if n.IsEmpty() {
					continue
				}

				// 执行事件, 删除这个事件
				e := n.Data.(*MsgX)
				println(e.Data)
				n.Pop()
			}
		}
	}()
}

func (this *WriteSilk) PostData(d int) {
	m := &MsgX{Data: d}
	n := &help.DListNode{}
	n.Init(m)
	println("PostData b")
	G_thread_msg_pool.PushMsg(this.SildId, n)
}

const (
	SilkCount = 10
)

func main() {

	G_thread_msg_pool.Init(10, 1200)

	xw := make([]WriteSilk, SilkCount)
	for i := uint(0); i < SilkCount; i++ {
		xw[i].Init(i + 1)
		xw[i].Run()
	}

	println("Wait silk")

	for i := 0; i < 100; i++ {
		time.Sleep(2 * time.Second)

		for k := 0; k < SilkCount; k++ {
			println("PostData")
			xw[k].PostData(i + 1)
		}
	}
}
