package invoker

import (
	"bufio"
	"context"
	"io"
	"time"
)

// DrainOut insures that no leaks from invoke result channel.
func DrainOut(out <-chan InvokeResult) {
	for res := range out {
		res.Discard()
	}
}

// NewWatchedSafeBuffer creates a safe buffer and starts to scan it,
// using splitFn that by-default reads lines, and scanFn that processes
// each scanned chunk.
func NewWatchedSafeBuffer(
	ctx context.Context,
	scanFn func(data []byte) (stop bool),
	splitFn bufio.SplitFunc,
) (buf *SafeBuffer) {
	buf = NewSafeBuffer()

	scanBuffer := func(buf io.Reader) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				s := bufio.NewScanner(buf)
				if splitFn != nil {
					s.Split(splitFn)
				}

				for s.Scan() {
					if scanFn(s.Bytes()) {
						return
					}
				}

				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	go scanBuffer(buf)

	return buf
}
