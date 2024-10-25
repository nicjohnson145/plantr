package agent

import (
	"bytes"
	"os/exec"
)

var (
	// Escape hatch for unit testing seeds that call OS commands
	unitTestExecuteFunc func(string, ...string) (string, string, error)
)

func ExecuteOSCommand(bin string, args ...string) (string, string, error) {
	if unitTestExecuteFunc != nil {
		return unitTestExecuteFunc(bin, args...)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
