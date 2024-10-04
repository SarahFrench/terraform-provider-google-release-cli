package changelog

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/SarahFrench/terraform-provider-google-release-cli/internal/config"
	"github.com/SarahFrench/terraform-provider-google-release-cli/internal/input"
)

type ChangeLogRun struct {
	Input                    input.Input
	Config                   *config.Config
	LastReleaseCommit        string
	LastCommitCurrentRelease string

	Dir    string
	StdErr *bytes.Buffer
	StdOut *bytes.Buffer
}

func (cl *ChangeLogRun) GenerateChangelog() error {

	changelogCmd := exec.Command("changelog-gen",
		"-repo", cl.Input.GetProviderRepoName(),
		"-branch", "main",
		"-owner", cl.Config.RemoteOwner,
		"-changelog", fmt.Sprintf("%s/.ci/changelog.tmpl", cl.Config.MagicModulesPath),
		"-releasenote", fmt.Sprintf("%s/.ci/release-note.tmpl", cl.Config.MagicModulesPath),
		"-no-note-label", "\"changelog: no-release-note\"",
		cl.LastReleaseCommit,
		cl.LastCommitCurrentRelease,
	)
	changelogCmd.Dir = cl.Dir

	cl.StdErr = &bytes.Buffer{}
	changelogCmd.Stderr = cl.StdErr

	cl.StdOut = &bytes.Buffer{}
	changelogCmd.Stdout = cl.StdOut

	err := changelogCmd.Run()
	return err
}

func (cl *ChangeLogRun) String() string {
	return cl.StdOut.String()
}
