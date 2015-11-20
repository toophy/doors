package silk

import (
	"sync"
)

type Silk struct {
	Id int // 微线ID
}

func (this *Silk) Init(id int) {
}

func (this *Silk) GetId() int {
	return this.Id
}

type SilkRead struct {
	Silk
}

// 发送线程间消息
func (this *Thread) sendMsg(tid int32, a help.IEvent) {
	G_thread_msg_pool.PostMsg(tid, a)
}

type SildWrite struct {
	Silk
}

func (this *SildWrite) readMsg() {

	header := help.DListNode{}
	header.Init(nil)

	G_thread_msg_pool.GetMsg(this.GetId(), &header)

	for {
		// 每次得到链表第一个事件(非)
		n := header.Next
		if n.IsEmpty() {
			break
		}

		// 执行事件, 删除这个事件
		e := n.Data.(help.IEvent)
		e.Exec(this.self)
		e.Destroy()

		// 节点 n, 需要原始线程回收
		old_pre := this.node_preFree[n.SrcTid].Pre

		this.node_preFree[n.SrcTid].Pre = n
		n.Next = this.node_preFree[n.SrcTid]
		n.Pre = old_pre
		old_pre.Next = n
	}
}

type Session struct {
}
