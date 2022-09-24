package invoker

import (
	"context"
	"io"
	"io/ioutil"
	"os/exec"
	"sync"
	"time"
)

type InvokeResult struct {
	Error error

	stdErr *SafeBuffer
	stdOut *SafeBuffer
}

func (res *InvokeResult) StdErr() []byte {
	data, _ := ioutil.ReadAll(res.stdErr)
	return data
}

func (res *InvokeResult) StdOut() []byte {
	data, _ := ioutil.ReadAll(res.stdOut)
	return data
}

func (res *InvokeResult) Discard() {
	if res.stdErr != nil {
		res.stdErr.Discard()
	}

	if res.stdOut != nil {
		res.stdOut.Discard()
	}
}

type Invoker interface {
	Run(ctx context.Context, args ...string) <-chan InvokeResult
	RunWithIO(
		ctx context.Context,
		stdIn io.Reader,
		stdOut, stdErr io.Writer,
		args ...string,
	) <-chan InvokeResult
	WorkingDir() string
	Close()
}

func NewInvoker(binPath, workDir string) Invoker {
	return &invoker{
		wg:      new(sync.WaitGroup),
		binPath: binPath,
		workDir: workDir,
	}
}

type invoker struct {
	binPath string
	workDir string

	wg *sync.WaitGroup
}

func (inv *invoker) WorkingDir() string {
	return inv.workDir
}

func (inv *invoker) Run(ctx context.Context, args ...string) <-chan InvokeResult {
	out := make(chan InvokeResult, 1)

	inv.wg.Add(1)

	go func() {
		defer inv.wg.Done()

		stdErr := NewSafeBuffer()
		stdOut := NewSafeBuffer()
		result := InvokeResult{
			stdOut: stdOut,
			stdErr: stdErr,
		}

		defer func() {
			out <- result
			close(out)
		}()

		cmd := exec.CommandContext(ctx, inv.binPath, args...)
		cmd.Dir = inv.workDir
		cmd.Stdout = stdOut
		cmd.Stderr = stdErr
		result.Error = cmd.Run()
	}()

	return out
}

func (inv *invoker) RunWithIO(
	ctx context.Context,
	stdIn io.Reader,
	stdOut, stdErr io.Writer,
	args ...string,
) <-chan InvokeResult {
	out := make(chan InvokeResult, 1)

	inv.wg.Add(1)

	go func() {
		defer inv.wg.Done()

		ownStdOut := NewSafeBuffer()
		ownStdErr := NewSafeBuffer()
		result := InvokeResult{
			stdOut: ownStdOut,
			stdErr: ownStdErr,
		}

		defer func() {
			out <- result
			close(out)
		}()

		cmd := exec.CommandContext(ctx, inv.binPath, args...)
		cmd.Dir = inv.workDir

		if stdIn != nil {
			cmd.Stdin = stdIn
		}

		if stdOut != nil {
			// write to both
			cmd.Stdout = io.MultiWriter(stdOut, ownStdOut)
		} else {
			cmd.Stdout = ownStdOut
		}

		if stdErr != nil {
			// write to both
			cmd.Stderr = io.MultiWriter(stdErr, ownStdErr)
		} else {
			cmd.Stderr = ownStdErr
		}

		result.Error = cmd.Run()
		// enough time to flush buffers
		time.Sleep(time.Second)
	}()

	return out
}

func (inv *invoker) Close() {
	inv.wg.Wait()
}
