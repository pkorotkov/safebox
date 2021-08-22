package utils

import (
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"syscall"
)

func LaunchProgram(name string, args []string, stdin []byte, envs []string) ([]byte, []byte, error) {
	cmd := exec.Command(name, args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewBuffer(stdin)
	}
	cmd.Env = append(os.Environ(), envs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func MakeDirectory(path string, uid, gid uint32) error {
	cmd := exec.Command("mkdir", "-p", path)
	syscall.Umask(0077) // Set umask for this process
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid}
	return cmd.Run()
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func CaseInsensitiveReplace(subject string, search string, replace string) string {
	sre := regexp.MustCompile("(?i)" + search)
	return sre.ReplaceAllString(subject, replace)
}
