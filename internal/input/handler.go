package input

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Handler struct {
	reader *bufio.Reader

	input *Input
}

func NewHandler(input *Input) Handler {

	reader := bufio.NewReader(os.Stdin)

	return Handler{
		reader: reader,
		input:  input,
	}
}

func (h *Handler) WaitForResponse() (string, error) {
	pv, err := h.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	pv = prepareStdinInput(pv)
	return pv, nil
}

func (h *Handler) PromptAndProcessProviderChoiceInput() error {

	fmt.Println("What provider do you want to make a release for (ga/beta)?")

	pv, err := h.WaitForResponse()
	if err != nil {
		return err
	}
	if err := h.input.SetProvider(pv); err != nil {
		return err
	}
	return nil
}

func (h *Handler) PromptAndProcessReleaseVersionChoiceInput(lastRelaseVersion, possibleNextVersion string) error {

	fmt.Printf("The latest release of %s is %s\n", h.input.GetProviderRepoName(), lastRelaseVersion)
	fmt.Printf("Are you planning on making the next minor release, %s? (y/n)\n", possibleNextVersion)

	in, err := h.WaitForResponse()
	if err != nil {
		return err
	}

	switch in {
	case "y":
		if err := h.input.SetReleaseVersions(possibleNextVersion, lastRelaseVersion); err != nil {
			return err
		}
	case "n":
		// The user might be making a patch release, major release, or a backport. Asking for previous version and new version enables all these.
		fmt.Println("Provide the previous release version as a semver string, e.g. v1.2.3:")

		old, err := h.reader.ReadString('\n')
		if err != nil {
			return err
		}
		old = prepareStdinInput(old)

		fmt.Println("Provide the new release version we are prepating as a semver string, e.g. v1.2.3:")
		new, err := h.reader.ReadString('\n')
		if err != nil {
			return err
		}
		new = prepareStdinInput(new)

		if err := h.input.SetReleaseVersions(new, old); err != nil {
			return err
		}
	}
	return errors.New("bad input where y/n was expected, exiting")
}

func (h *Handler) PromptAndProcessCommitChoiceInput() error {

	fmt.Println("What commit do you want to use to cut the release?")

	c, err := h.WaitForResponse()
	if err != nil {
		return err
	}

	if err := h.input.SetCommit(c); err != nil {
		return err
	}
	return nil
}

func prepareStdinInput(in string) string {
	in = strings.TrimSuffix(in, "\n")
	in = strings.ToLower(in)
	return in
}
