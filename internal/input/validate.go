package input

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

type compositeValidationError []error

func (v compositeValidationError) Error() string {
	var b strings.Builder
	b.WriteString("There were some problems with inputs to the command:\n")
	for _, e := range v {
		b.WriteString(fmt.Sprintf("\t> %v\n", e))
	}
	return b.String()
}

func validateGaBetaInputs(ga, beta bool) error {
	if ga && beta {
		return fmt.Errorf("you should provide only one of the -ga and -beta flags")
	}
	if !ga && !beta {
		return fmt.Errorf("you need to provide at least one of the -ga and -beta flags")
	}
	return nil
}

func validateProviderInputs(providerVersion string) error {
	pv := strings.ToLower(providerVersion)
	switch pv {
	case "ga":
		return nil
	case "beta":
		return nil
	default:
		return fmt.Errorf("bad provider version input, please answer 'ga' or 'beta' ")
	}
}

func validateCommitShaInput(commitSha string) error {
	if commitSha == "" {
		return fmt.Errorf("you need to provide a commit SHA to be the basis of the new release")
	}
	return nil
}

func validateVersionInputs(new, old string) error {
	// Assert provided
	if new == "" || old == "" {
		return fmt.Errorf("make sure to provide values for both release_version and prev_release_version flags")
	}

	// Assert inputs are valid format
	err := checkSemVer(new, old)
	if err != nil {
		return err
	}

	// Assert new is a later version than old, and not the same
	if semver.Compare(new, old) != +1 {
		return fmt.Errorf("the new release we're preparing (%s) should be a more recent version number than the provided previous version number (%s)", new, old)
	}

	return nil
}

func checkSemVer(new, old string) error {

	if !semver.IsValid(new) {
		return fmt.Errorf("release_version is not in correct format")
	}
	if !semver.IsValid(old) {
		return fmt.Errorf("prev_release_version is not in correct format")
	}

	return nil
}
