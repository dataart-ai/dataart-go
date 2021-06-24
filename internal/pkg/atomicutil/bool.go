package atomicutil

import "sync/atomic"

type Bool int32

func (b *Bool) IsSet() bool {
	return atomic.LoadInt32((*int32)(b)) != 0
}

func (b *Bool) SetTrue() {
	atomic.StoreInt32((*int32)(b), 1)
}

func (b *Bool) SetFalse() {
	atomic.StoreInt32((*int32)(b), 0)
}
