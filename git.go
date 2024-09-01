package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type CommonCommitFinder struct {
	Dir             string
	PreviousRelease string

	description          string
	commitShaLastRelease string

	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func NewCommonCommitFinder(directory, previousRelease string) *CommonCommitFinder {
	return &CommonCommitFinder{
		Dir:             directory,
		PreviousRelease: previousRelease,

		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}
}

func (c *CommonCommitFinder) GetLastReleaseCommit() (string, error) {
	c.wipeStreams()

	cmd := exec.Command("git", "merge-base", "main", fmt.Sprintf("v%s", c.PreviousRelease))
	cmd.Dir = c.Dir
	cmd.Stderr = c.stderr
	cmd.Stdout = c.stdout

	c.description = cmd.String()

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	c.commitShaLastRelease = strings.ReplaceAll(c.stdout.String(), "\n", "")
	return c.commitShaLastRelease, nil
}

func (c *CommonCommitFinder) errorDescription(summary string, err error) string {
	return fmt.Sprintf("%s:\n\tCommand: `%s`\n\tError: %s\n\tStdErr: %s", summary, c.description, err, c.stderr.String())
}

// wipeStreams ensures that the buffers for stderr/stdout are empty before running a new command
func (c *CommonCommitFinder) wipeStreams() {
	c.stderr = &bytes.Buffer{}
	c.stdout = &bytes.Buffer{}
}
