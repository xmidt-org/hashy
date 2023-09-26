package hashy

import "sync"

type BufferPool struct {
	size int
	pool sync.Pool
}

func (bp *BufferPool) newBuffer() any {
	return make([]byte, bp.size)
}

func (bp *BufferPool) Get() []byte {
	return bp.pool.Get().([]byte)
}

func (bp *BufferPool) Put(b []byte) {
	if cap(b) != bp.size {
		return
	}

	bp.pool.Put(b[:cap(b)])
}

func NewBufferPool(bufferSize int) *BufferPool {
	bp := &BufferPool{
		size: bufferSize,
	}

	bp.pool.New = bp.newBuffer
	return bp
}
