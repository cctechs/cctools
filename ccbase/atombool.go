package ccbase

import "sync/atomic"

// 封装bool类型的原子变量
type AtomBool struct {
	val int32
}

func(this*AtomBool) Set(flag bool){
	var i int32 = 0
	if flag{
		i = 1
	}
	atomic.StoreInt32(&this.val, i)
}

func(this*AtomBool) Get() bool{
	i := atomic.LoadInt32(&this.val)
	if 0 == i{
		return false
	}
	return true
}