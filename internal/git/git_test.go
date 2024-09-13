package git

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

type executeMock struct {
	testArg string
}

// Command wraps exec.Command
func (e *executeMock) Command(name string, arg ...string) Cmder {
	return &CmdMock{
		cmd:  &CommandMock{},
		name: name,
		args: arg,

		testArg: e.testArg,
	}
}

type CommandMock struct {
	Dir    string
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer

	testArg string
}

type CmdMock struct {
	cmd         *CommandMock
	name        string
	args        []string
	description string

	testArg string
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

	// This switch allows us to force the commands to fail with specific errors
	// or just generic errors to assert an error is returned
	switch c.testArg {
	case "CreateAndPushReleaseBranch fails on checkout":
		return errors.New("forced error from test mock")
	}

	return nil
}

func (c *CmdMock) Description() string {
	return ""
}

func NewTestGitInteract(directory, previousRelease, remote, testArg string) *GitInteract {
	return &GitInteract{
		Dir:             directory,
		PreviousRelease: previousRelease,
		Remote:          remote,

		exec: &executeMock{
			testArg: testArg,
		},
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
	newReleaseVersion := "1.2.0"
	remote := "origin"

	cases := map[string]struct {
		expectedCommand string
		expectedOutput  string
		callFunction    func(*GitInteract) (string, GitCommand, error)
		testArg         string // argument used to control failures of the run command
		expectError     bool
	}{
		"Checkout: function should checkout the provided branch name": {
			expectedCommand: "git checkout main",
			callFunction: func(gi *GitInteract) (string, GitCommand, error) {
				gc, err := gi.Checkout("main")
				return "", gc, err
			},
		},
		"GetLastReleaseCommit: function should perform merge-base between main and the provided previous version": {
			expectedCommand: "git merge-base main v" + previousReleaseVersion,
			callFunction: func(gi *GitInteract) (string, GitCommand, error) {
				return gi.GetLastReleaseCommit()
			},
		},
		"PullTagsMainBranch: function should perform 'git pull <remote> --tags'": {
			expectedCommand: "git pull " + remote + " --tags",
			callFunction: func(gi *GitInteract) (string, GitCommand, error) {
				gc, err := gi.PullTagsMainBranch()
				return "", gc, err
			},
		},
		"CreateAndPushReleaseBranch: function should finish by performing 'git checkout <string>'": {
			expectedCommand: "git push -u " + remote + " release-" + newReleaseVersion,
			expectedOutput:  "release-" + newReleaseVersion,
			callFunction: func(gi *GitInteract) (string, GitCommand, error) {
				return gi.CreateAndPushReleaseBranch(newReleaseVersion)
			},
		},
		"CreateAndPushReleaseBranch: function stops after first command if it errors": {
			testArg:         "CreateAndPushReleaseBranch fails on checkout",
			expectError:     true,
			expectedCommand: "git checkout -b " + fmt.Sprintf("release-%s", newReleaseVersion),
			callFunction: func(gi *GitInteract) (string, GitCommand, error) {
				return gi.CreateAndPushReleaseBranch(newReleaseVersion)
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			gi := NewTestGitInteract(dir, previousReleaseVersion, remote, tc.testArg)

			output, gc, err := tc.callFunction(gi)
			if err != nil {
				if tc.expectError == false {
					t.Fatal("unexpected error occurred")
				}
				// fall through on error, as we force these as part of testing
			}

			if output != tc.expectedOutput {
				t.Fatalf("wanted output %s, got: %s", tc.expectedOutput, output)
			}

			gcMock := gc.cmd.(*CmdMock)
			if gcMock.description != tc.expectedCommand {
				t.Fatalf("wanted %s, got: %s", tc.expectedCommand, gcMock.description)
			}
		})
	}
}
