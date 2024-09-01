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

	exec Execer
}

// Execer is used to mock out exec.Command
type Execer interface {
	Command(name string, arg ...string) Cmder
}

type execute struct{}

// Command wraps exec.Command
func (e *execute) Command(name string, arg ...string) Cmder {
	c := exec.Command(name, arg...)
	return &Cmd{
		cmd: c,
	}
}

// Cmder interface is used to allow mocking of *exec.Cmd
type Cmder interface {
	Run() error
	Description() string

	SetDir(string)
	GetDir() string

	SetStderr(*bytes.Buffer)
	SetStdout(*bytes.Buffer)
}

type Cmd struct {
	cmd *exec.Cmd
}

func (c *Cmd) SetDir(dir string) {
	c.cmd.Dir = dir
}

func (c *Cmd) GetDir() string {
	return c.cmd.Dir
}

func (c *Cmd) SetStderr(stderr *bytes.Buffer) {
	c.cmd.Stderr = stderr
}

func (c *Cmd) SetStdout(stdout *bytes.Buffer) {
	c.cmd.Stdout = stdout
}

func (c *Cmd) Run() error {
	return c.cmd.Run()
}

func (c *Cmd) Description() string {
	return c.cmd.String()
}

// GitCommand describes a git command that has been executed
type GitCommand struct {
	cmd    Cmder // Typically *exec.Cmd
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	runErr error
}

func NewGitInteract(directory, previousRelease, remote string) *GitInteract {
	return &GitInteract{
		Dir:             directory,
		PreviousRelease: previousRelease,
		Remote:          remote,

		exec: &execute{},
	}
}

func (c *GitInteract) GetLastReleaseCommit() (string, GitCommand, error) {
	gc := GitCommand{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}
	gc.cmd = c.exec.Command("git", "merge-base", "main", fmt.Sprintf("v%s", c.PreviousRelease))
	gc.cmd.SetDir(c.Dir)
	gc.cmd.SetStderr(gc.stderr)
	gc.cmd.SetStdout(gc.stdout)

	err := gc.cmd.Run()
	if err != nil {
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
	gc.cmd = c.exec.Command("git", "pull", c.Remote, "--tags")
	gc.cmd.SetDir(c.Dir)
	gc.cmd.SetStderr(gc.stderr)
	gc.cmd.SetStdout(gc.stdout)

	err := gc.cmd.Run()
	if err != nil {
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
	gc.cmd = c.exec.Command("git", "checkout", ref)
	gc.cmd.SetDir(c.Dir)
	gc.cmd.SetStderr(gc.stderr)
	gc.cmd.SetStdout(gc.stdout)

	err := gc.cmd.Run()
	if err != nil {
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
	gc.cmd = c.exec.Command("git", "checkout", "-b", branchName)
	gc.cmd.SetDir(c.Dir)
	gc.cmd.SetStderr(gc.stderr)
	gc.cmd.SetStdout(gc.stdout)

	err := gc.cmd.Run()
	if err != nil {
		gc.runErr = err
		return "", gc, err
	}

	// Push branch
	gc = GitCommand{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}
	gc.cmd = c.exec.Command("git", "push", "-u", c.Remote, branchName)
	gc.cmd.SetDir(c.Dir)
	gc.cmd.SetStderr(gc.stderr)
	gc.cmd.SetStdout(gc.stdout)

	err = gc.cmd.Run()
	if err != nil {
		gc.runErr = err
		return "", gc, err
	}

	return branchName, gc, nil
}

// errorDescription returns a formatted string describing how a CLI command has failed
// The output includes:
//   - a user-supplied summary for the error
//   - the command
//   - the directory
//   - the error returned from (c *exec.Cmd).Run()
//   - the stderr from the command
func (gc *GitCommand) ErrorDescription(summary string) string {
	return fmt.Sprintf("%s:\n\tCommand: `%s`\n\tDirectory: %s\n\tError: %s\n\tStdErr: %s", summary, gc.cmd.GetDir(), gc.cmd.Description(), gc.runErr, gc.stderr.String())
}
