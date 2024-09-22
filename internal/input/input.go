package input

import "errors"

type Provider int

const (
	UNSET Provider = iota
	GA
	BETA
)

type Input struct {
	// CommitSha is the SHA1 hash of the commit we want to use as the basis of the new release
	CommitSha string
	// ReleaseVersion is the new release's semver tag in format v1.2.3
	ReleaseVersion string
	// PreviousReleaseVersion is the latest release's semver tag in format v1.2.3
	PreviousReleaseVersion string
	Provider               Provider
}

func (i *Input) GetProviderRepoName() (string, error) {
	switch i.Provider {
	case GA:
		return "terraform-provider-google", nil
	case BETA:
		return "terraform-provider-google-beta", nil
	default:
		return "", errors.New("error determining which repo is being generated")
	}
}

func New(ga, beta bool, commitSha, releaseVersion, previousReleaseVersion string) (Input, error) {

	errs := compositeValidationError{}

	if err := validateGaBetaInputs(ga, beta); err != nil {
		errs = append(errs, err)
	}
	if err := validateCommitShaInput(commitSha); err != nil {
		errs = append(errs, err)
	}
	if err := validateVersionInputs(releaseVersion, previousReleaseVersion); err != nil {
		errs = append(errs, err...)
	}

	if len(errs) > 0 {
		return Input{}, errs
	}

	i := Input{
		CommitSha:              commitSha,
		ReleaseVersion:         releaseVersion,
		PreviousReleaseVersion: previousReleaseVersion,
	}
	if ga {
		i.Provider = GA
	}
	if beta {
		i.Provider = BETA
	}

	return i, nil
}
