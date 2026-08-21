package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const (
	helperProcessEnv = "XKCDPASS_TEST_PROCESS"
	helperVersionEnv = "XKCDPASS_TEST_VERSION"
	testVersion      = "v1.2.3-test"
)

func TestExecutable(t *testing.T) {
	stdout, stderr, exitCode := runExecutable(t, "--words", "3", "--separator", "/")

	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", exitCode, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty output", stderr)
	}
	if strings.Count(stdout, "\n") != 1 || !strings.HasSuffix(stdout, "\n") {
		t.Errorf("stdout = %q, want one newline-terminated passphrase", stdout)
	}
	if output := strings.TrimSuffix(stdout, "\n"); strings.Count(output, "/") != 2 {
		t.Errorf("stdout = %q, want three slash-separated words", stdout)
	}
}

func TestExecutableHelp(t *testing.T) {
	stdout, stderr, exitCode := runExecutable(t, "--help")

	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", exitCode, stderr)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty output", stdout)
	}
	if !strings.Contains(stderr, "Usage:\n  xkcdpass [--words N] [--separator SEP]") {
		t.Errorf("stderr = %q, want usage text", stderr)
	}
}

func TestExecutableVersion(t *testing.T) {
	stdout, stderr, exitCode := runExecutable(t, "--version")

	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", exitCode, stderr)
	}
	if stdout != testVersion+"\n" {
		t.Errorf("stdout = %q, want %q", stdout, testVersion+"\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty output", stderr)
	}
}

func TestExecutableInvalidInput(t *testing.T) {
	stdout, stderr, exitCode := runExecutable(t, "--words", "0")

	if exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr=%q", exitCode, stderr)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty output", stdout)
	}
	if !strings.Contains(stderr, "xkcdpass: --words must be at least 1") {
		t.Errorf("stderr = %q, want invalid-input message", stderr)
	}
}

func TestDevelopmentVersion(t *testing.T) {
	if version != "dev" {
		t.Fatalf("version = %q, want %q", version, "dev")
	}
}

func runExecutable(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	commandArgs := append([]string{"-test.run=^TestExecutableProcess$", "--"}, args...)
	command := exec.Command(os.Args[0], commandArgs...)
	command.Env = append(os.Environ(), helperProcessEnv+"=1", helperVersionEnv+"="+testVersion)

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	command.Stdout = &stdoutBuffer
	command.Stderr = &stderrBuffer

	err := command.Run()
	if err == nil {
		return stdoutBuffer.String(), stderrBuffer.String(), 0
	}

	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("run executable: %v", err)
	}

	return stdoutBuffer.String(), stderrBuffer.String(), exitError.ExitCode()
}

func TestExecutableProcess(t *testing.T) {
	if os.Getenv(helperProcessEnv) != "1" {
		return
	}

	for i, arg := range os.Args {
		if arg != "--" {
			continue
		}

		version = os.Getenv(helperVersionEnv)
		os.Args = append([]string{"xkcdpass"}, os.Args[i+1:]...)
		main()
	}

	os.Exit(125)
}
