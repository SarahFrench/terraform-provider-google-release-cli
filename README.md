# terraform-provider-google Release CLI
A CLI to help making releases of the Terraform providers for Google Cloud

## Configuring the tool

You need to create a file called `.tpg-cli-config.json` in your HOME directory:

```bash
touch $HOME/.tpg-cli-config.json
```

In that file you need to paste a JSON like below, which contains:
- magicModulesPath : the (absolute) path to where you have cloned the https://github.com/GoogleCloudPlatform/magic-modules repository
- googlePath : the (absolute) path to where you have cloned the https://github.com/hashicorp/terraform-provider-google repository
- googleBetaPath : the (absolute) path to where you have cloned the https://github.com/hashicorp/terraform-provider-google-beta repository
- remote : in your cloned copies of terraform-provider-google(-beta), the name of the "remote"  that corresponds to the official repo. If you're unsure, `cd` into those repos and run `git remote`.


```bash
vi  $HOME/.tpg-cli-config.json
```

```json
{
    "magicModulesPath": "/Users/Foobar/go/src/github.com/Foobar/magic-modules",
    "googlePath": "/Users/Foobar/go/src/github.com/Foobar/terraform-provider-google",
    "googleBetaPath": "/Users/Foobar/go/src/github.com/Foobar/terraform-provider-google-beta",
    "remote": "origin",
    "githubToken": "<PAT token with no permissions>"
}
```


## Using the CLI

The CLI can be run using flags or can interactively ask for input values.

You can run the CLI from any directory and you don't need to worry about checking out a given branch before starting to cut the release branch.


### Interactive mode

In interactive mode you'll be asked for:
- GA vs Beta provider choice
- Whether you want to make the next minor release, or supply your own last/next versions
- (if supplying your own last/next versions)
   - Prompt for previous release version
   - Prompt for the new release version
- The commit to cut the release from

For example:

```

> What provider do you want to make a release for (ga/beta)?
ga

> The latest release of terraform-provider-google is v6.5.0
  Are you planning on making the next minor release, v6.6.0? (y/n)
n

> Provide the previous release version as a semver string, e.g. v1.2.3:
v9.9.9

> Provide the new release version we are prepating as a semver string, e.g. v1.2.3:
v9.10.0

> What commit do you want to use to cut the release?
33db873052ab34b92b5f6512bd874730a0f83164

Starting to create and push new release branch

Release branch release-9.9.10 was created and pushed

Copy the CHANGELOG below into : https://github.com/hashicorp/terraform-provider-google/edit/release-v9.9.10/CHANGELOG.md

<print out of changelog-gen tool to terminal>
```


### Using flags

| Flag                  | Usage                                                                                                                                         |
|-----------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| -ga                   | Flag to select creating a release for the GA provider. Cannot be used with -beta.                                                             |
| -beta                 | Flag to select creating a release for the Beta provider. Cannot be used with -ga.                                                             |
| -gh_token             | Set the value as a PAT with no permissions, see: https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token" |
| -commit_sha           | The commit from the main branch that will be used for the release.                                                                            |
| -release_version      | The version that we're about to prepare, in format v4.XX.0.                                                                                   |
| -prev_release_version | The previous version that was released, in format v4.XX.0.                                                                                    |


### Using a combination of flags and interactive prompts

It's possible to use a combination of flags and interactive prompts, and the tool will print to the terminal to let you know which is used.


## This CLI replaces the need to run bash commands when releasing a new version of the Google provider.

The Google provider's [release process is documented here]([https://github.com/hashicorp/terraform-provider-google/wiki/Release-Process](https://github.com/hashicorp/terraform-provider-google/wiki/Release-Process#on-wednesday)) as a large amount of bash:

```bash
# Fixed values- consider setting them in `.bash_profile` or `bashrc`
# REMOTE is the name of the primary repo's remote on your machine. Typically `upstream` or `origin`
REMOTE=upstream
# MM_REPO should point to your checked-out copy of the GoogleCloudPlatform/magic-modules repo
MM_REPO="path/to/magic-modules"
# https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token, no permissions
export GITHUB_TOKEN=

# Fill these in each time
# COMMIT_SHA is build.vcs.number in TeamCity
COMMIT_SHA= 
RELEASE_VERSION=4.XX.0
PREVIOUS_RELEASE_VERSION=4.XX.0

COMMIT_SHA_OF_LAST_RELEASE=`git merge-base main v${PREVIOUS_RELEASE_VERSION}`
REPO_NAME=$(basename $(git rev-parse --show-toplevel))
# use [ -n "$COMMIT_SHA" ] to make sure COMMIT_SHA is set, `git checkout` is a valid command on its own
git pull $REMOTE main --tags && [ -n "$COMMIT_SHA" ] && git checkout $COMMIT_SHA && git checkout -b release-$RELEASE_VERSION && git push -u $REMOTE release-$RELEASE_VERSION
COMMIT_SHA_OF_LAST_COMMIT_IN_CURRENT_RELEASE=`git rev-list -n 1 HEAD`
go install github.com/paultyng/changelog-gen@master
changelog-gen -repo $REPO_NAME -branch main -owner hashicorp -changelog ${MM_REPO}/.ci/changelog.tmpl -releasenote ${MM_REPO}/.ci/release-note.tmpl -no-note-label "changelog: no-release-note" $COMMIT_SHA_OF_LAST_RELEASE $COMMIT_SHA_OF_LAST_COMMIT_IN_CURRENT_RELEASE
open https://github.com/hashicorp/$REPO_NAME/edit/release-$RELEASE_VERSION/CHANGELOG.md
```

This bash needs to be run inside the repo of the provider that you're cutting the release for, but the CLI tool can be run from any directory.
