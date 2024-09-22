package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/SarahFrench/terraform-provider-google-release-cli/internal/config"
	"github.com/SarahFrench/terraform-provider-google-release-cli/internal/git"
	input_pkg "github.com/SarahFrench/terraform-provider-google-release-cli/internal/input"
)

var changelogExecutable string = "changelog-gen"

func main() {

	// Handle inputs via flags
	// var githubToken string //TODO
	var commitSha string
	var releaseVersion string
	var previousReleaseVersion string
	var ga bool
	var beta bool

	// flag.StringVar(&githubToken, "gh_token", "", "Create a PAT with no permissions, see: https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token")
	flag.StringVar(&commitSha, "commit_sha", "", "The commit from the main branch that will be used for the release")
	flag.StringVar(&releaseVersion, "release_version", "", "The version that we're about to prepare, in format 4.XX.0")
	flag.StringVar(&previousReleaseVersion, "previous_release_version", "", "The previous version that was released, in format 4.XX.0")
	flag.BoolVar(&ga, "ga", false, "Flag to start creating a release for the GA provider")
	flag.BoolVar(&beta, "beta", false, "Flag to start creating a release for the Beta provider")
	flag.Parse()

	input, err := input_pkg.New(ga, beta, commitSha, releaseVersion, previousReleaseVersion)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	// Load in config
	c, err := config.LoadConfigFromFile()
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	// Make sure dependencies present
	_, err = exec.LookPath(changelogExecutable)
	if err != nil {
		log.Fatal("you need to have changelog-gen in your PATH to use this CLI. Ensure it is in your PATH or download it via: go install github.com/paultyng/changelog-gen@master")
		os.Exit(1)
	}

	// Prepare
	providerRepoName, _ := input.GetProviderRepoName()
	gi := git.GitInteract{
		Dir:             providerRepoName,
		PreviousRelease: input.PreviousReleaseVersion,
		Remote:          c.Remote,
	}

	// Ensure we have checked out main
	cmd, err := gi.Checkout("main")
	if err != nil {
		log.Fatal(cmd.ErrorDescription("error when checking out provided commit SHA"))
	}

	// Run commands to create the release branch
	lastRelease, cmd, err := gi.GetLastReleaseCommit()
	if err != nil {
		log.Fatal(cmd.ErrorDescription("error when getting last release's commit"))
	}
	fmt.Println(lastRelease) // TODO remove when last release used in future code

	log.Print("Starting to create and push new release branch")

	// git pull $REMOTE main --tags
	cmd, err = gi.PullTagsMainBranch()
	if err != nil {
		log.Fatal(cmd.ErrorDescription("error when pulling tags"))
	}

	// git checkout $COMMIT_SHA
	cmd, err = gi.Checkout(input.CommitSha)
	if err != nil {
		log.Fatal(cmd.ErrorDescription("error when checking out provided commit SHA"))
	}

	// git checkout -b release-$RELEASE_VERSION && git push -u $REMOTE release-$RELEASE_VERSION
	branchName, cmd, err := gi.CreateAndPushReleaseBranch(input.ReleaseVersion)
	if err != nil {
		log.Fatal(cmd.ErrorDescription("error when creating a new release branch"))
	}

	log.Printf("Release branch %s was created and pushed", branchName)

	log.Printf("https://github.com/hashicorp/%s/edit/release-%s/CHANGELOG.md", providerRepoName, input.ReleaseVersion)

	log.Println("Done!")
}

func assertPathExists(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cannot find anything at the path '%s', is the path correct?", path)
		}
		return fmt.Errorf("unexpected error when locating %s", path)
	}
	return nil
}
