package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"

	"github.com/SarahFrench/terraform-provider-google-release-cli/internal/config"
	"github.com/SarahFrench/terraform-provider-google-release-cli/internal/git"
	input_pkg "github.com/SarahFrench/terraform-provider-google-release-cli/internal/input"
	"github.com/SarahFrench/terraform-provider-google-release-cli/internal/release_version"
)

var changelogExecutable string = "changelog-gen"

func main() {

	// Handle inputs via flags
	// var githubToken string //TODO
	var commitShaFlag string
	var releaseVersionFlag string
	var previousReleaseVersionFlag string
	var gaFlag bool
	var betaFlag bool

	// flag.StringVar(&githubToken, "gh_token", "", "Create a PAT with no permissions, see: https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token")
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
		fmt.Println("Release version infomation provided by flags:")
		fmt.Printf("\tPrevious release version: %s\n", previousReleaseVersionFlag)
		fmt.Printf("\tNew release version: %s\n", releaseVersionFlag)
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
	fmt.Println()
	if commitShaFlag != "" {
		// Info provided by flags
		fmt.Printf("Release cut commit provided by flag: %s\n", commitShaFlag)
		input.SetCommit(commitShaFlag)
	} else {
		// Need to get info via stdin
		handler.PromptAndProcessCommitChoiceInput()
	}

	// Double check inputs from above
	if err := input.Validate(); err != nil {
		log.Fatal(fmt.Errorf("validation error raised after collecting user inputs: %w", err))
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
	fmt.Println(lastReleaseCommit) // TODO remove when used in future code

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

	log.Printf("https://github.com/%s/%s/edit/release-%s/CHANGELOG.md", c.RemoteOwner, input.GetProviderRepoName(), input.ReleaseVersion)

	log.Println("Done!")
}
