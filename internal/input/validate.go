package input

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

func validateCommitShaInput(commitSha string) error {
	if commitSha == "" {
		return fmt.Errorf("you need to provide a commit SHA to be the basis of the new release")
	}
	return nil
}

func validateVersionInputs(new, old string) []error {
	validationErrs := compositeValidationError{}
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
