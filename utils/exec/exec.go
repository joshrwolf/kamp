package exec

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// CmdOpts docs todo
type CmdOpts struct {
	Timeout time.Duration
}

// RunCommandExt docs todo
func RunCommandExt(cmd *exec.Cmd, opts CmdOpts) (string, error) {
	logCtx := log.WithFields(log.Fields{"execID": 1})

	args := strings.Join(cmd.Args, " ")
	logCtx.WithFields(log.Fields{"dir": cmd.Dir}).Info(args)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Start()
	if err != nil {
		return "", err
	}

	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	// Start the timer
	timeout := time.Duration(0)
	if opts.Timeout != time.Duration(0) {
		timeout = opts.Timeout
	}

	var timeoutCh <-chan time.Time
	if timeout != 0 {
		timeoutCh = time.NewTimer(timeout).C
	}

	select {
	case <-timeoutCh:
		// Timeout occured
		_ = cmd.Process.Kill()
		output := stdout.String()
		logCtx.WithFields(log.Fields{"duration": time.Since(start)}).Debug(output)
		err = fmt.Errorf("timeout after %v", timeout)
		logCtx.Error(err.Error())
		return strings.TrimSuffix(output, "\n"), err
	case err := <-done:
		// Error occured in the command
		if err != nil {
			// Capture output and log it
			output := stdout.String()
			logCtx.WithFields(log.Fields{"duration": time.Since(start)}).Debug(output)
			err := fmt.Errorf("stderr: %v", stderr.String())
			logCtx.Error(err.Error())
			return strings.TrimSuffix(output, "\n"), err
		}
	}

	output := stdout.String()
	logCtx.WithFields(log.Fields{"duration": time.Since(start)}).Debug(output)

	return strings.TrimSuffix(output, "\n"), nil
}

func RunCommand(name string, opts CmdOpts, args ...string) (string, error) {
	return RunCommandExt(exec.Command(name, args...), opts)
}
