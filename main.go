package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/SarahFrench/terraform-provider-google-release-cli/internal/changelog"
	"github.com/SarahFrench/terraform-provider-google-release-cli/internal/config"
	"github.com/SarahFrench/terraform-provider-google-release-cli/internal/git"
	input_pkg "github.com/SarahFrench/terraform-provider-google-release-cli/internal/input"
	"github.com/SarahFrench/terraform-provider-google-release-cli/internal/release_version"
)

var changelogExecutable string = "changelog-gen"

func main() {

	// Handle inputs via flags
	var githubToken string
	var commitShaFlag string
	var releaseVersionFlag string
	var previousReleaseVersionFlag string
	var gaFlag bool
	var betaFlag bool

	flag.StringVar(&githubToken, "gh_token", "", "Create a PAT with no permissions, see: https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token")
	flag.StringVar(&commitShaFlag, "commit_sha", "", "The commit from the main branch that will be used for the release")
	flag.StringVar(&releaseVersionFlag, "release_version", "", "The version that we're about to prepare, in format v4.XX.0")
	flag.StringVar(&previousReleaseVersionFlag, "prev_release_version", "", "The previous version that was released, in format v4.XX.0")
	flag.BoolVar(&gaFlag, "ga", false, "Flag to start creating a release for the GA provider")
	flag.BoolVar(&betaFlag, "beta", false, "Flag to start creating a release for the Beta provider")
	flag.Parse()

	// Load in config
	c, err := config.LoadConfigFromFile()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Make sure dependencies present
	_, err = exec.LookPath(changelogExecutable)
	if err != nil {
		log.Fatal("you need to have changelog-gen in your PATH to use this CLI. Ensure it is in your PATH or download it via: go install github.com/paultyng/changelog-gen@master")
	}

	// Ready to collect input
	input := input_pkg.Input{}
	handler := input_pkg.NewHandler(&input)

	// PROVIDER CHOICE
	fmt.Println()
	if gaFlag || betaFlag {
		// Info provided by flags
		fmt.Println("Provider choice set via flag:")
		err := input.SetProviderFromFlags(gaFlag, betaFlag)
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Printf("\tMaking a release for %s\n", input.GetProviderRepoName())
	} else {
		// Need to get info via stdin
		err = handler.PromptAndProcessProviderChoiceInput()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	// RELEASE VERSION CHOICE
	if releaseVersionFlag != "" || previousReleaseVersionFlag != "" {
		// Info provided by flags
		log.Println("Release version infomation provided by flags:")
		log.Printf("\tPrevious release version: %s\n", previousReleaseVersionFlag)
		log.Printf("\tNew release version: %s\n", releaseVersionFlag)
		err := input.SetReleaseVersions(releaseVersionFlag, previousReleaseVersionFlag)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {

		// Prepare info about the last release and proposed new minor release versions.
		rq := release_version.New(c.RemoteOwner, input.GetProviderRepoName())
		latestVersion, err := rq.GetLastVersionFromGitHub()
		if err != nil {
			log.Fatal(err.Error())
		}
		proposedNextVersion, err := release_version.NextMinorVersion(latestVersion)
		if err != nil {
			log.Fatal(err.Error())
		}

		// Need to get info via stdin
		handler.PromptAndProcessReleaseVersionChoiceInput(latestVersion, proposedNextVersion)
	}

	// 'COMMIT TO CUT RELEASE ON' CHOICE
	if commitShaFlag != "" {
		// Info provided by flags
		log.Printf("Release cut commit provided by flag: %s\n", commitShaFlag)
		input.SetCommit(commitShaFlag)
	} else {
		// Need to get info via stdin
		handler.PromptAndProcessCommitChoiceInput()
	}

	// Double check inputs from above
	if err := input.Validate(); err != nil {
		log.Fatal(fmt.Errorf("validation error raised after collecting user inputs: %w", err))
	}
	if githubToken == "" && c.GitHubToken == "" {
		log.Fatal("no GitHub token provided: either add one to your config file or supply using a -gh_token flag")
	}

	// Prepare
	dir := c.GetProviderDirectoryPath(input.GetProviderRepoName())
	gi := git.GitInteract{
		Dir:             dir,
		PreviousRelease: input.PreviousReleaseVersion,
		Remote:          c.Remote,
	}

	// Ensure we have checked out main
	cmd, err := gi.Checkout("main")
	if err != nil {
		log.Fatal(cmd.ErrorDescription("error when checking out provided commit SHA"))
	}

	// Run commands to create the release branch
	lastReleaseCommit, cmd, err := gi.GetLastReleaseCommit()
	if err != nil {
		log.Fatal(cmd.ErrorDescription("error when getting last release's commit"))
	}

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

	// This should be the same as input.CommitSha, but the release process includes running
	// git rev-list -n 1 HEAD
	lastCommitCurrentRelease, cmd, err := gi.GetLastCommitOfCurrentRelease(branchName)
	if err != nil {
		log.Fatal(cmd.ErrorDescription("error when getting last commit of current release"))
	}

	log.Printf("Release branch %s was created and pushed", branchName)

	log.Println("Creating CHANGELOG entry")

	token := githubToken
	if token == "" {
		// Flag takes precedence over config
		token = c.GitHubToken
	}
	os.Setenv("GITHUB_TOKEN", token)
	defer os.Setenv("GITHUB_TOKEN", "")

	// changelog-gen -repo $REPO_NAME -branch main -owner hashicorp -changelog ${MM_REPO}/.ci/changelog.tmpl -releasenote ${MM_REPO}/.ci/release-note.tmpl -no-note-label "changelog: no-release-note" $COMMIT_SHA_OF_LAST_RELEASE $COMMIT_SHA_OF_LAST_COMMIT_IN_CURRENT_RELEASE
	cl := changelog.ChangeLogRun{
		Input:                    input,
		Config:                   c,
		LastReleaseCommit:        lastReleaseCommit,
		LastCommitCurrentRelease: lastCommitCurrentRelease,

		Dir: dir,
	}
	cl.GenerateChangelog()
	output := cl.String()

	fmt.Print("\n---\n")
	fmt.Printf("\n\033[32m" + output)
	fmt.Print("\n---\n")

	log.Printf("Copy the CHANGELOG above into : https://github.com/%s/%s/edit/release-%s/CHANGELOG.md", c.RemoteOwner, input.GetProviderRepoName(), input.ReleaseVersion)

}
