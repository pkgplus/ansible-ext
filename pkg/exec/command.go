package exec

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	osexec "os/exec"
	"sync"
	"syscall"
)

type Command struct {
	execCmd    *osexec.Cmd
	cmd        string
	args       []string
	outChan    chan interface{}
	StderrRet  []byte
	ReturnCode int
	wg         sync.WaitGroup
}

func NewCommand(cmd string, args ...string) *Command {
	return &Command{
		cmd:     cmd,
		args:    args,
		outChan: make(chan interface{}, 1),
	}
}

type StreamExecuter interface {
	Execute([]byte, chan interface{}, bool)
}

func (c *Command) RunStreaming(se StreamExecuter) error {
	cmd := osexec.Command(c.cmd, c.args...)
	c.execCmd = cmd

	// stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.Wait()
		return err
	}

	// stderr
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	defer func() {
		c.StderrRet = errbuf.Bytes()
	}()

	// start run command
	err = cmd.Start()
	if err != nil {
		c.Wait()
		return err
	}

	// read stdout
	c.wg.Add(1)
	go func() {
		defer func() {
			c.wg.Done()
			c.Close()
			stdout.Close()
			c.Wait()
		}()

		bio := bufio.NewReader(stdout)
		for {
			line, err := bio.ReadSlice('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "read from cmd '%s %+v' err: %v\n", c.cmd, c.args, err)
				}
				se.Execute(line, c.outChan, true)
				break
			} else {
				se.Execute(line, c.outChan, false)
			}
		}

	}()

	return nil
}

func (c *Command) Messages() <-chan interface{} {
	return c.outChan
}

func (c *Command) Wait() error {
	defer log.Printf("finish running %s ,returnCode:%d", c.cmd, c.ReturnCode)
	c.wg.Wait()

	err := c.execCmd.Wait()
	if err != nil {
		if exiterr, ok := err.(*osexec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				c.ReturnCode = status.ExitStatus()
			}
		} else {
			return err
		}
	}

	return nil
}

func (c *Command) Close() {
	close(c.outChan)
}
