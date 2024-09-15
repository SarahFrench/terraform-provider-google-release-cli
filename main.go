package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/SarahFrench/terraform-provider-google-release-cli/internal/config"
	"github.com/SarahFrench/terraform-provider-google-release-cli/internal/git"
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

	// Validate inputs where possible
	validationErrs := []error{} // Collect errors and report them all later

	// Load in config
	c, errs := config.LoadConfigFromFile()
	if len(errs) > 0 {
		validationErrs = append(validationErrs, errs...)
	}

	errs = validateVersionInputs(releaseVersion, previousReleaseVersion)
	if len(errs) != 0 {
		validationErrs = append(validationErrs, errs...)
	}

	if ga && beta {
		validationErrs = append(validationErrs, fmt.Errorf("you should provide only one of the -ga and -beta flags"))
	}
	if !ga && !beta {
		validationErrs = append(validationErrs, fmt.Errorf("you need to provide at least one of the -ga and -beta flags"))
	}

	if commitSha == "" {
		validationErrs = append(validationErrs, fmt.Errorf("you need to provide a commit SHA to be the basis of the new release"))
	}

	// Make sure dependencies present
	_, err := exec.LookPath(changelogExecutable)
	if err != nil {
		validationErrs = append(validationErrs, fmt.Errorf("you need to have changelog-gen in your PATH to use this CLI. Ensure it is in your PATH or download it via: go install github.com/paultyng/changelog-gen@master"))
	}

	if len(validationErrs) > 0 {
		fmt.Println("There were some problems with inputs to the command:")
		for _, e := range validationErrs {
			fmt.Printf("\t> %v\n", e)
		}
		os.Exit(1)
	}

	// Prepare
	var providerDir string
	var providerRepo string

	switch {
	case ga:
		providerRepo = "terraform-provider-google"
		providerDir = c.GooglePath
	case beta:
		providerRepo = "terraform-provider-google-beta"
		providerDir = c.GoogleBetaPath
	default:
		log.Fatal("error determining which repo is being generated")
	}

	gi := git.GitInteract{
		Dir:             providerDir,
		PreviousRelease: previousReleaseVersion,
		Remote:          c.Remote,
	}

	// Ensure we have checkout out main
	cmd, err := gi.Checkout("main")
	if err != nil {
		log.Fatal(cmd.ErrorDescription("error when checking out provided commit SHA"))
	}

	// Run commands to create the release branch
	lastRelease, cmd, err := gi.GetLastReleaseCommit()
	if err != nil {
		log.Fatal(cmd.ErrorDescription("error when getting last release's commit"))
	}
	fmt.Println(lastRelease)

	log.Print("Starting to create and push new release branch")

	// git pull $REMOTE main --tags
	cmd, err = gi.PullTagsMainBranch()
	if err != nil {
		log.Fatal(cmd.ErrorDescription("error when pulling tags"))
	}

	// git checkout $COMMIT_SHA
	cmd, err = gi.Checkout(commitSha)
	if err != nil {
		log.Fatal(cmd.ErrorDescription("error when checking out provided commit SHA"))
	}

	// git checkout -b release-$RELEASE_VERSION && git push -u $REMOTE release-$RELEASE_VERSION
	branchName, cmd, err := gi.CreateAndPushReleaseBranch(releaseVersion)
	if err != nil {
		log.Fatal(cmd.ErrorDescription("error when creating a new release branch"))
	}

	log.Printf("Release branch %s was created and pushed", branchName)

	log.Printf("https://github.com/hashicorp/%s/edit/release-%s/CHANGELOG.md", providerRepo, releaseVersion)

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

func validateVersionInputs(new, old string) []error {

	validationErrs := []error{}
	// Assert provided
	if new == "" || old == "" {
		validationErrs = append(validationErrs, fmt.Errorf("make sure to provide values for both release_version and previous_release_version flags"))
	}

	// Assert inputs are valid format
	err := checkSemVer(new, old)
	if err != nil {
		validationErrs = append(validationErrs, err)
	}

	// Assert not the same
	if new == old {
		validationErrs = append(validationErrs, fmt.Errorf("the version we're preparing (%s) should be a more recent version number than the provided previous version number (%s)", new, old))
	}

	// Assert new is a later version than old
	semverRECaps := `^(?P<major>\d{1,3})\.(?P<minor>\d{1,3})\.(?P<patch>\d{1,3})$`
	re := regexp.MustCompile(semverRECaps)
	newMatches := re.FindStringSubmatch(new)
	oldMatches := re.FindStringSubmatch(old)

	// Skip 0 element as it contains the whole input
	for i := 1; i < len(newMatches); i++ {
		n, err := strconv.Atoi(newMatches[i])
		if err != nil {
			validationErrs = append(validationErrs, fmt.Errorf("error converting %s to integer", newMatches[i]))
		}
		o, err := strconv.Atoi(oldMatches[i])
		if err != nil {
			validationErrs = append(validationErrs, fmt.Errorf("error converting %s to integer", oldMatches[i]))
		}
		if o > n {
			validationErrs = append(validationErrs, fmt.Errorf("the version we're preparing (%s) should be a more recent version number than the provided previous version number (%s)", new, old))
		}
	}

	return validationErrs
}

func checkSemVer(new, old string) error {
	semverRe := regexp.MustCompile(`^\d{1}\.\d{1,2}\.\d{1,2}$`)

	if !semverRe.MatchString(new) {
		return fmt.Errorf("release_version is not in correct format")
	}
	if !semverRe.MatchString(old) {
		return fmt.Errorf("previous_release_version is not in correct format")
	}

	return nil
}
