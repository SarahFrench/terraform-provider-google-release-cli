package git

import (
	"bytes"
	"testing"
)

type executeMock struct{}

// Command wraps exec.Command
func (e *executeMock) Command(name string, arg ...string) Cmder {
	return &CmdMock{
		cmd:  &CommandMock{},
		name: name,
		args: arg,
	}
}

type CommandMock struct {
	Dir    string
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
}

type CmdMock struct {
	cmd         *CommandMock
	name        string
	args        []string
	description string
}

func (c *CmdMock) SetDir(dir string) {
	c.cmd.Dir = dir
}

func (c *CmdMock) GetDir() string {
	return c.cmd.Dir
}

func (c *CmdMock) SetStderr(stderr *bytes.Buffer) {
	c.cmd.Stderr = stderr
}

func (c *CmdMock) SetStdout(stdout *bytes.Buffer) {
	c.cmd.Stdout = stdout
}

func (c *CmdMock) Run() error {
	c.description = c.name
	for _, v := range c.args {
		c.description = c.description + " " + v
	}
	return nil
}

func (c *CmdMock) Description() string {
	return ""
}

func NewTestGitInteract(directory, previousRelease, remote string) *GitInteract {
	return &GitInteract{
		Dir:             directory,
		PreviousRelease: previousRelease,
		Remote:          remote,

		exec: &executeMock{},
	}
}

func Test_NewGitInteract(t *testing.T) {
	dir := "dir"
	version := "1.1.1"
	remote := "origin"
	gi := NewGitInteract(dir, version, remote)

	if gi.Dir != dir {
		t.Fatalf("Dir: got %s, want %s", gi.Dir, dir)
	}
	if gi.PreviousRelease != version {
		t.Fatalf("PreviousRelease: got %s, want %s", gi.PreviousRelease, version)
	}
	if gi.Remote != remote {
		t.Fatalf("Remote: got %s, want %s", gi.Remote, remote)
	}
}

func Test_NewGitInteract_commands(t *testing.T) {
	dir := "dir"
	previousReleaseVersion := "1.1.1"
	remote := "origin"

	cases := map[string]struct {
		expectedCommand string
		callFunction    func(*GitInteract) (string, GitCommand, error)
	}{
		"Checkout function should checkout the provided branch name": {
			expectedCommand: "git checkout main",
			callFunction: func(gi *GitInteract) (string, GitCommand, error) {
				gc, err := gi.Checkout("main")
				return "", gc, err
			},
		},
		"GetLastReleaseCommit function should perform merge-base between main and the provided previous version": {
			expectedCommand: "git merge-base main v" + previousReleaseVersion,
			callFunction: func(gi *GitInteract) (string, GitCommand, error) {
				return gi.GetLastReleaseCommit()
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			gi := NewTestGitInteract(dir, previousReleaseVersion, remote)

			_, gc, err := tc.callFunction(gi)
			if err != nil {
				t.Fatal("unexpected error occurred")
			}

			gcMock := gc.cmd.(*CmdMock)
			if gcMock.description != tc.expectedCommand {
				t.Fatalf("wanted %s, got: %s", tc.expectedCommand, gcMock.description)
			}
		})
	}
}
