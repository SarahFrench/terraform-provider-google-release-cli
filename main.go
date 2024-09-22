package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

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
	reader := bufio.NewReader(os.Stdin)

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
		fmt.Println("What provider do you want to make a release for (ga/beta)?")

		pv, err := reader.ReadString('\n') // blocks until the delimiter is entered
		if err != nil {
			log.Fatal(err.Error())
		}
		pv = prepareStdinInput(pv)
		if err := input.SetProvider(pv); err != nil {
			log.Fatal(err.Error())
		}
	}

	// RELEASE VERSION CHOICE
	rq := release_version.New(c.RemoteOwner, input.GetProviderRepoName())
	latestVersion, err := rq.GetLastVersionFromGitHub()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println()
	if releaseVersionFlag != "" || previousReleaseVersionFlag != "" {
		// Info provided by flags
		fmt.Println("Release version infomation provided by flags:")
		fmt.Printf("\tPrevious release version: %s\n", previousReleaseVersionFlag)
		fmt.Printf("\tNew release version: %s\n", releaseVersionFlag)
	} else {
		// Need to get info via stdin
		nextVersion, err := release_version.NextMinorVersion(latestVersion)
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Printf("The latest release of %s is %s\n", input.GetProviderRepoName(), latestVersion)
		fmt.Printf("Are you planning on making the next minor release, %s? (y/n)\n", nextVersion)

		in, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err.Error())
		}
		in = prepareStdinInput(in)
		switch in {
		case "y":
			if err := input.SetReleaseVersions(nextVersion, latestVersion); err != nil {
				log.Fatal(err.Error())
			}
		case "n":
			fmt.Println("Provide the previous release version as a semver string, e.g. v1.2.3:")
			old, err := reader.ReadString('\n') // blocks until the delimiter is entered
			if err != nil {
				log.Fatal(err.Error())
			}
			old = prepareStdinInput(old)

			fmt.Println("Provide the new release version we are prepating as a semver string, e.g. v1.2.3:")
			new, err := reader.ReadString('\n') // blocks until the delimiter is entered
			if err != nil {
				log.Fatal(err.Error())
			}
			new = prepareStdinInput(new)

			if err := input.SetReleaseVersions(new, old); err != nil {
				log.Fatal(err.Error())
			}
		default:
			log.Fatal("bad input where y/n was expected, exiting")
		}
	}

	// 'COMMIT TO CUT RELEASE ON' CHOICE
	fmt.Println()
	if commitShaFlag != "" {
		// Info provided by flags
		fmt.Printf("Release cut commit provided by flag: %s\n", commitShaFlag)
		input.SetCommit(commitShaFlag)
	} else {
		// Need to get info via stdin
		fmt.Println("What commit do you want to use to cut the release?")

		c, err := reader.ReadString('\n') // blocks until the delimiter is entered
		if err != nil {
			log.Fatal(err.Error())
		}
		c = prepareStdinInput(c)
		if err := input.SetCommit(c); err != nil {
			log.Fatal(err.Error())
		}
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

func prepareStdinInput(in string) string {
	in = strings.TrimSuffix(in, "\n")
	in = strings.ToLower(in)
	return in
}
