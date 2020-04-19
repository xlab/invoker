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
	Success bool
	Error   error

	stdErr io.Reader
	stdOut io.Reader
}

func (res *InvokeResult) StdErr() []byte {
	data, _ := ioutil.ReadAll(res.stdErr)
	return data
}

func (res *InvokeResult) StdOut() []byte {
	data, _ := ioutil.ReadAll(res.stdOut)
	return data
}

type Invoker interface {
	Run(ctx context.Context, args ...string) <-chan InvokeResult
	RunWithOutputs(
		ctx context.Context,
		stdErr, stdOut io.ReadWriter,
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
			stdErr: stdErr,
			stdOut: stdOut,
		}

		defer func() {
			out <- result
			close(out)
		}()

		cmd := exec.CommandContext(ctx, inv.binPath, args...)
		cmd.Dir = inv.workDir
		cmd.Stderr = stdErr
		cmd.Stdout = stdOut
		result.Error = cmd.Run()
	}()

	return out
}

func (inv *invoker) RunWithOutputs(
	ctx context.Context,
	stdOut, stdErr io.ReadWriter,
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
