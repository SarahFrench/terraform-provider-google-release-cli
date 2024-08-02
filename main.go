package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
)

var changelogExecutable string = "changelog-gen"

var defaultRemote string = "origin"

func main() {

	// Handle inputs via flags
	var remote string
	var mmRepoPath string
	// var githubToken string //TODO
	var commitSha string
	var releaseVersion string
	var previousReleaseVersion string
	var ga bool
	var beta bool

	flag.StringVar(&remote, "remote", defaultRemote, "REMOTE is the name of the primary repo's remote on your machine. Typically `upstream` or `origin`")
	flag.StringVar(&mmRepoPath, "mm_repo_path", "", "should point to your checked-out copy of the GoogleCloudPlatform/magic-modules repo")
	// flag.StringVar(&githubToken, "gh_token", "", "Create a PAT with no permissions, see: https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token")
	flag.StringVar(&commitSha, "commit_sha", "", "The commit from the main branch that will be used for the release")
	flag.StringVar(&releaseVersion, "release_version", "", "The version that we're about to prepare, in format 4.XX.0")
	flag.StringVar(&previousReleaseVersion, "previous_release_version", "", "The previous version that was released, in format 4.XX.0")
	flag.BoolVar(&ga, "ga", false, "Flag to start creating a release for the GA provider")
	flag.BoolVar(&beta, "beta", false, "Flag to start creating a release for the Beta provider")
	flag.Parse()

	// Validate inputs where possible

	err := assertPathExists(mmRepoPath)
	if err != nil {
		log.Fatal(err)
	}

	err = validateVersionInputs(releaseVersion, previousReleaseVersion)
	if err != nil {
		log.Fatal(err)
	}

	if ga && beta {
		log.Fatal("you should provide only one of the -ga and -beta flags")
	}
	if !ga && !beta {
		log.Fatal("you need to provide at least one of the -ga and -beta flags")
	}

	if commitSha == "" {
		log.Fatal("you need to provide a commit SHA to be the basis of the new release")
	}

	// Make sure dependencies present
	_, err = exec.LookPath(changelogExecutable)
	if err != nil {
		log.Fatal("you need to have changelog-gen in your PATH to use this CLI. Ensure it is in your PATH or download it via: go install github.com/paultyng/changelog-gen@master")
	}

	// Prepare
	HOME := os.Getenv("HOME")
	var providerDir string
	var providerRepo string
	if ga {
		providerRepo = "terraform-provider-google"
		providerDir = HOME + "/go/src/github.com/SarahFrench/terraform-provider-google"
	}
	if beta {
		providerRepo = "terraform-provider-google-beta"
		providerDir = HOME + "/go/src/github.com/SarahFrench/terraform-provider-google-beta"
	}

	// Run commands to create the release branch
	cmd := exec.Command("git", "merge-base", "main", fmt.Sprintf("v%s", previousReleaseVersion))
	cmd.Dir = providerDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("error when running `%s`:\n\t%s", cmd.String(), string(output))
	}
	commitShaLastRelease := string(output)
	fmt.Println(commitShaLastRelease)

	log.Printf("Starting to create and push new release branch %s", fmt.Sprintf("release-%s", releaseVersion))

	// git pull $REMOTE main --tags && [ -n "$COMMIT_SHA" ] && git checkout $COMMIT_SHA && git checkout -b release-$RELEASE_VERSION && git push -u $REMOTE release-$RELEASE_VERSION
	cmd = exec.Command("git", "pull", remote, "--tags")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("error when pulling tags: `%s` :\n\t%s", cmd.String(), string(output))
	}

	cmd = exec.Command("git", "checkout", commitSha)
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("error when checking out provided commit SHA: `%s` :\n\t%s", cmd.String(), string(output))
	}

	cmd = exec.Command("git", "checkout", "-b", fmt.Sprintf("release-%s", releaseVersion))
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("error when creating a new release branch: `%s` :\n\t%s", cmd.String(), string(output))
	}

	cmd = exec.Command("git", "push", "-u", remote, fmt.Sprintf("release-%s", releaseVersion))
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("error when pushing the new release branch: `%s` :\n\t%s", cmd.String(), string(output))
	}

	log.Printf("Release branch %s was created and pushed", fmt.Sprintf("release-%s", releaseVersion))

	log.Printf("https://github.com/hashicorp/%s/edit/release-%s/CHANGELOG.md", providerRepo, releaseVersion)

	log.Println("Done!")
}

func assertPathExists(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cannot find %s, is the path correct?", path)
		}
		return fmt.Errorf("unexpected error when locating %s", path)
	}
	return nil
}

func validateVersionInputs(new, old string) error {

	// Assert inputs are valid format
	err := checkSemVer(new, old)
	if err != nil {
		return err
	}

	// TODO - fix, this regards 0.0.10 as less than 0.0.9
	// Assert new is a later version than old
	// if old > new || old == new {
	// 	return fmt.Errorf("the version we're preparing (%s) should be a more recent version number than the provided previous version number (%s)", new, old)
	// }

	return nil
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
