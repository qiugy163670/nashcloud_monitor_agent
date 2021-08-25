package cmd

import (
	"bufio"
	"container/list"
	fmt "fmt"
	"os"
	"os/exec"
)

/*
* cmd
 */

type ProcessAgentCheck struct {
	BinPath     string
	commandOpts []string
	running     uint32
	stop        chan struct{}
	stopDone    chan struct{}
	source      string
	telemetry   bool
}

func init() {
	//fmt.Println("cmd_agent is init")
}
func (c *ProcessAgentCheck) ExecCmd4String(cmdStr string) (error, string) {
	str := ""
	select {
	case <-c.stop:
		fmt.Println("Not starting Process Agent check: stop requested")
		c.stopDone <- struct{}{}
		return nil, "nil"
	default:
	}

	cmd := exec.Command(c.BinPath, "-c", cmdStr)
	stdOut, err := cmd.StdoutPipe()

	if err != nil {
		fmt.Println(err)
	}

	go func() {
		in := bufio.NewScanner(stdOut)
		for in.Scan() {
			//fmt.Println(in.Text())
			str += in.Text()
		}
	}()

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return err, "nil"
	}
	go func() {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			//fmt.Println(in.Text())
			str += in.Text()
		}
	}()

	if err = cmd.Start(); err != nil {
		//return retryExitError(err)
	}

	processDone := make(chan error)
	go func() {
		processDone <- cmd.Wait()
	}()

	select {
	case err = <-processDone:
		//return retryExitError(err)
	case <-c.stop:
		err = cmd.Process.Signal(os.Kill)
		if err != nil {
			fmt.Println("unable to stop process-agent check: %s", err)
		}
	}

	// wait for process to exit
	//err = <-processDone
	//c.stopDone <- struct{}{}

	return err, str
}
func (c *ProcessAgentCheck) ExecCmd(cmdStr string) (error, list.List) {
	l := list.New()
	select {
	case <-c.stop:
		fmt.Println("Not starting Process Agent check: stop requested")
		c.stopDone <- struct{}{}
		return nil, *l
	default:
	}

	cmd := exec.Command(c.BinPath, "-c", cmdStr)
	stdOut, err := cmd.StdoutPipe()

	if err != nil {
		fmt.Println(err)
	}

	go func() {
		in := bufio.NewScanner(stdOut)
		for in.Scan() {
			//fmt.Println(in.Text())
			l.PushBack(in.Text())
		}
	}()

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return err, *l
	}
	go func() {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			//fmt.Println(in.Text())
			l.PushBack(in.Text())
		}
	}()

	if err = cmd.Start(); err != nil {
		//return retryExitError(err)
	}

	processDone := make(chan error)
	go func() {
		processDone <- cmd.Wait()
	}()

	select {
	case err = <-processDone:
		//return retryExitError(err)
	case <-c.stop:
		err = cmd.Process.Signal(os.Kill)
		if err != nil {
			fmt.Println("unable to stop process-agent check: %s", err)
		}
	}

	// wait for process to exit
	//err = <-processDone
	//c.stopDone <- struct{}{}

	return err, *l
}
