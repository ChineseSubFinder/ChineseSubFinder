package cmdutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func ExecAndGetStdoutJson(cmd *exec.Cmd, v interface{}) error {
	b, err := ExecAndGetStdoutBytes(cmd)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, v)
	if err != nil {
		return err
	}
	return nil
}

func ExecAndGetStdoutBytes(cmd *exec.Cmd) ([]byte, error) {
	b := new(bytes.Buffer)
	if err := ExecAndWriteStdout(cmd, b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func ExecAndWriteStdout(cmd *exec.Cmd, w io.Writer) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error opening stdout of command: %v", err)
	}
	defer stdout.Close()
	log.Debugf("Executing: %v %v", cmd.Path, cmd.Args)
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %v", err)
	}
	if _, err := io.Copy(w, stdout); err != nil {
		// Ask the process to exit
		cmd.Process.Signal(syscall.SIGKILL)
		cmd.Process.Wait()
		return fmt.Errorf("error copying stdout to buffer: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed %v", err)
	}
	return nil
}
