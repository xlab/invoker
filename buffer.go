package invoker

import (
	"bytes"
	"sync"
)

// SafeBuffer introduces a mutex lock, making all operations safe
// and non-conflicting: read operation cannot modify buffer until write is finished.
//
// Used in buffer polling contexts, see NewWatchedSafeBuffer of this package.
type SafeBuffer struct {
	buf *bytes.Buffer
	mux *sync.Mutex

	discarded bool
}

// NewSafeBuffer constructs a new safe buffer.
func NewSafeBuffer() *SafeBuffer {
	return &SafeBuffer{
		buf: new(bytes.Buffer),
		mux: new(sync.Mutex),

		discarded: false,
	}
}

func (b *SafeBuffer) Read(p []byte) (n int, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.notDiscarded()

	return b.buf.Read(p)
}

func (b *SafeBuffer) Write(p []byte) (n int, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.notDiscarded()

	return b.buf.Write(p)
}

func (b *SafeBuffer) String() string {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.notDiscarded()

	return b.buf.String()
}

func (b *SafeBuffer) Bytes() []byte {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.notDiscarded()

	return b.buf.Bytes()
}

func (b *SafeBuffer) Discard() {
	b.mux.Lock()
	defer b.mux.Unlock()

	if b.discarded {
		return
	}

	b.buf.Reset()
	b.buf = nil
	b.discarded = true
}

func (b *SafeBuffer) notDiscarded() {
	if b.discarded {
		panic("access to discarded buffer")
	}
}
