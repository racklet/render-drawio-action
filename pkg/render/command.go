package render

// This file is partly copied from
// https://github.com/cloud-native-nordics/workshopctl/blob/9ff2a0dd74e4e01ec3f148f712b8a503ad40062c/pkg/util/util.go
// However, I, @luxas, authored this code in the first place, and workshopctl is also Apache 2.0 licensed.
// Some time I'll just make a standalone repo for this.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

func Command(ctx context.Context, command string, args ...string) *ExecUtil {
	return &ExecUtil{
		cmd:    exec.CommandContext(ctx, command, args...),
		outBuf: new(bytes.Buffer),
		ctx:    ctx,
	}
}

func ShellCommand(ctx context.Context, format string, args ...interface{}) *ExecUtil {
	return Command(ctx, "/bin/sh", "-c", fmt.Sprintf(format, args...))
}

type ExecUtil struct {
	cmd    *exec.Cmd
	outBuf *bytes.Buffer
	ctx    context.Context
}

func (e *ExecUtil) Cmd() *exec.Cmd {
	return e.cmd
}

func (e *ExecUtil) WithStdio(stdin io.Reader, stdout, stderr io.Writer) *ExecUtil {
	log := zap.L()
	if stdin != nil {
		log.Debug("Set command stdin")
		e.cmd.Stdin = stdin
	}
	if stdout != nil {
		log.Debug("Set command stdout")
		e.cmd.Stdout = stdout
	}
	if stderr != nil {
		log.Debug("Set command stderr")
		e.cmd.Stderr = stderr
	}
	return e
}

func (e *ExecUtil) WithPwd(pwd string) *ExecUtil {
	zap.L().Debug("Set command", zap.String("pwd", pwd))
	e.cmd.Dir = pwd
	return e
}

func (e *ExecUtil) WithEnv(envVars ...string) *ExecUtil {
	zap.L().Debug("Set command", zap.Strings("env vars", envVars))
	e.cmd.Env = append(e.cmd.Env, envVars...)
	return e
}

func (e *ExecUtil) Run() (output string, exitCode int, cmdErr error) {
	cmdArgs := strings.Join(e.cmd.Args, " ")
	log := zap.S()

	// Always capture stdout output to e.outBuf
	if e.cmd.Stdout != nil {
		e.cmd.Stdout = io.MultiWriter(e.cmd.Stdout, e.outBuf)
	} else {
		e.cmd.Stdout = e.outBuf
	}
	// Always capture stderr output to e.outBuf
	if e.cmd.Stderr != nil {
		e.cmd.Stderr = io.MultiWriter(e.cmd.Stderr, e.outBuf)
	} else {
		e.cmd.Stderr = e.outBuf
	}
	// Run command
	log.Debugf("Running command %q", cmdArgs)
	err := e.cmd.Run()

	// Capture combined output
	output = string(bytes.TrimSpace(e.outBuf.Bytes()))
	if len(output) != 0 {
		log.Debugf("Command %q produced output: %s", cmdArgs, output)
	}

	// Handle the error
	if err != nil {
		exitCodeStr := "'unknown'"
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
			exitCodeStr = fmt.Sprintf("%d", exitCode)
		}

		cmdErr = fmt.Errorf("external command %q exited with code %s, error: %w and output: %s", cmdArgs, exitCodeStr, err, output)
		log.Debugf("Command error: %v", cmdErr)
	}
	return
}
