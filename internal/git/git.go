package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GitInteract contains all the user-supplied information for interacting with git while preparing the releas
type GitInteract struct {
	Dir             string
	PreviousRelease string
	Remote          string
}

// GitCommand describes a git command that has been executed
type GitCommand struct {
	cmd    *exec.Cmd
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	runErr error
}

func NewGitInteract(directory, previousRelease string) *GitInteract {
	return &GitInteract{
		Dir:             directory,
		PreviousRelease: previousRelease,
	}
}

func (c *GitInteract) GetLastReleaseCommit() (string, GitCommand, error) {

	// Get the common commit between the last release and the new release we're preparing
	gc := GitCommand{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}
	gc.cmd = exec.Command("git", "merge-base", "main", c.PreviousRelease)
	gc.cmd.Dir = c.Dir
	gc.cmd.Stderr = gc.stderr
	gc.cmd.Stdout = gc.stdout

	if err := gc.cmd.Run(); err != nil {
		gc.runErr = err
		return "", gc, err
	}

	commit := strings.ReplaceAll(gc.stdout.String(), "\n", "")
	return commit, gc, nil
}

func (c *GitInteract) PullTagsMainBranch() (GitCommand, error) {
	gc := GitCommand{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}
	gc.cmd = exec.Command("git", "pull", c.Remote, "--tags")
	gc.cmd.Dir = c.Dir
	gc.cmd.Stderr = gc.stderr
	gc.cmd.Stdout = gc.stdout

	if err := gc.cmd.Run(); err != nil {
		gc.runErr = err
		return gc, err
	}

	return gc, nil
}

func (c *GitInteract) Checkout(ref string) (GitCommand, error) {
	gc := GitCommand{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}
	gc.cmd = exec.Command("git", "checkout", ref)
	gc.cmd.Dir = c.Dir
	gc.cmd.Stderr = gc.stderr
	gc.cmd.Stdout = gc.stdout

	if err := gc.cmd.Run(); err != nil {
		gc.runErr = err
		return gc, err
	}

	return gc, nil
}

func (c *GitInteract) CreateAndPushReleaseBranch(releaseVersion string) (string, GitCommand, error) {
	gc := GitCommand{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}

	branchName := fmt.Sprintf("release-%s", releaseVersion)

	// Create branch locally
	gc.cmd = exec.Command("git", "checkout", "-b", branchName)
	gc.cmd.Dir = c.Dir
	gc.cmd.Stderr = gc.stderr
	gc.cmd.Stdout = gc.stdout

	if err := gc.cmd.Run(); err != nil {
		gc.runErr = err
		return "", gc, err
	}

	// Push branch
	gc = GitCommand{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}
	gc.cmd = exec.Command("git", "push", "-u", c.Remote, branchName)
	gc.cmd.Dir = c.Dir
	gc.cmd.Stderr = gc.stderr
	gc.cmd.Stdout = gc.stdout

	if err := gc.cmd.Run(); err != nil {
		gc.runErr = err
		return "", gc, err
	}

	return branchName, gc, nil
}

// ErrorDescription returns a formatted string describing how a CLI command has failed
// The output includes:
//   - a user-supplied summary for the error
//   - the command
//   - the directory
//   - the error returned from (c *exec.Cmd).Run()
//   - the stderr from the command
func (gc *GitCommand) ErrorDescription(summary string) string {
	return fmt.Sprintf("%s:\n\tCommand: `%s`\n\tDirectory: %s\n\tError: %s\n\tStdErr: %s", summary, gc.cmd.String(), gc.cmd.Dir, gc.runErr, gc.stderr.String())
}
