package input

import (
	"errors"
	"fmt"
)

type Provider int

const (
	UNSET Provider = iota
	GA
	BETA
)

var providerToString = map[Provider]string{
	UNSET: "UNSET",
	GA:    "GA",
	BETA:  "BETA",
}

var GA_REPO_NAME = "terraform-provider-google"
var BETA_REPO_NAME = "terraform-provider-google-beta"

type Input struct {
	// CommitSha is the SHA1 hash of the commit we want to use as the basis of the new release
	CommitSha string
	// ReleaseVersion is the new release's semver tag in format v1.2.3
	ReleaseVersion string
	// PreviousReleaseVersion is the latest release's semver tag in format v1.2.3
	PreviousReleaseVersion string
	// Provider records whether we're creating a relase for the GA or Beta version of the provider
	Provider Provider
}

func (i *Input) Validate() error {

	var errs compositeValidationError

	if err := validateCommitShaInput(i.CommitSha); err != nil {
		errs = append(errs, err)
	}
	if err := validateVersionInputs(i.ReleaseVersion, i.PreviousReleaseVersion); err != nil {
		errs = append(errs, err)
	}
	if !(i.Provider == GA || i.Provider == BETA) {
		errs = append(errs, errors.New("provider is not set"))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (i *Input) SetCommit(commit string) error {
	err := validateCommitShaInput(commit)
	if err != nil {
		return err
	}

	i.CommitSha = commit
	return nil
}

func (i *Input) SetProvider(providerVersion string) error {
	err := validateProviderInputs(providerVersion)
	if err != nil {
		return err
	}

	switch providerVersion {
	case "ga":
		i.Provider = GA
	case "beta":
		i.Provider = BETA
	default:
		return fmt.Errorf("unexpected provider version value: %s, expected 'ga' or 'beta'", providerVersion)
	}
	return nil
}

func (i *Input) SetProviderFromFlags(ga, beta bool) error {
	err := validateGaBetaInputs(ga, beta)
	if err != nil {
		return err
	}

	switch {
	case ga:
		i.Provider = GA
	case beta:
		i.Provider = BETA
	}
	return nil
}

func (i *Input) SetReleaseVersions(new, old string) error {

	err := validateVersionInputs(new, old)
	if err != nil {
		return err
	}

	i.PreviousReleaseVersion = old
	i.ReleaseVersion = new

	return nil
}

func (i *Input) GetProviderRepoName() string {
	switch i.Provider {
	case GA:
		return GA_REPO_NAME
	case BETA:
		return BETA_REPO_NAME
	case UNSET:
		return "provider has not been set"
	default:
		return ""
	}
}

// New returns an instance of Input
// This function is intended for use with flag inputs only
func New(ga, beta bool, commitSha, releaseVersion, previousReleaseVersion string) (Input, error) {

	errs := compositeValidationError{}

	if err := validateGaBetaInputs(ga, beta); err != nil {
		errs = append(errs, err)
	}
	if err := validateCommitShaInput(commitSha); err != nil {
		errs = append(errs, err)
	}
	if err := validateVersionInputs(releaseVersion, previousReleaseVersion); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return Input{}, errs
	}

	i := Input{
		CommitSha:              commitSha,
		ReleaseVersion:         releaseVersion,
		PreviousReleaseVersion: previousReleaseVersion,
	}
	switch {
	case ga:
		i.Provider = GA
	case beta:
		i.Provider = BETA
	default:
		return Input{}, fmt.Errorf("unable to determing provider version value from inputs")
	}

	return i, nil
}
