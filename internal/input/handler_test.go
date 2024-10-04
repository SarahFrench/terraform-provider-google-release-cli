package input

import (
	"bufio"
	"bytes"
	"testing"
)

func Test_Handler_PromptAndProcessProviderChoiceInput(t *testing.T) {

	cases := map[string]struct {
		userInput              string
		expectedChosenProvider Provider
		expectError            bool
	}{
		"choosing GA provider: ga": {
			userInput:              "ga\n",
			expectedChosenProvider: GA,
		},
		"choosing GA provider: GA": {
			userInput:              "GA\n",
			expectedChosenProvider: GA,
		},
		"choosing Beta provider: beta": {
			userInput:              "beta\n",
			expectedChosenProvider: BETA,
		},
		"choosing Beta provider: BETA": {
			userInput:              "BETA\n",
			expectedChosenProvider: BETA,
		},
		"responding incorrectly": {
			userInput:   "foobar\n",
			expectError: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			stdin := bytes.Buffer{}
			stdin.Write([]byte(tc.userInput))

			input := Input{}
			handler := NewHandler(&input)

			r := bufio.NewReader(&stdin)
			handler.reader = r

			err := handler.PromptAndProcessProviderChoiceInput()
			if err != nil && !tc.expectError {
				t.Fatal(err.Error())
			}
			if err == nil && tc.expectError {
				t.Fatal("expected error but got none")
			}

			if input.Provider != tc.expectedChosenProvider {
				t.Fatalf("wanted %v, got %v", providerToString[tc.expectedChosenProvider], providerToString[input.Provider])
			}
		})
	}

}

func Test_Handler_PromptAndProcessCommitChoiceInput(t *testing.T) {

	cases := map[string]struct {
		userInput      string
		expectError    bool
		expectedCommit string
	}{
		"provide empty commit SHA": {
			userInput:   "\n",
			expectError: true,
		},
		"provide any string input": {
			userInput:      "abcdefghijklmnopqrstuvwxyz\n",
			expectedCommit: "abcdefghijklmnopqrstuvwxyz",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			stdin := bytes.Buffer{}
			stdin.Write([]byte(tc.userInput))

			input := Input{}
			handler := NewHandler(&input)

			r := bufio.NewReader(&stdin)
			handler.reader = r

			err := handler.PromptAndProcessCommitChoiceInput()
			if err != nil && !tc.expectError {
				t.Fatal(err.Error())
			}
			if err == nil && tc.expectError {
				t.Fatal("expected error but got none")
			}

			if input.CommitSha != tc.expectedCommit {
				t.Fatalf("wanted %v, got %v", tc.expectedCommit, input.CommitSha)
			}
		})
	}
}

func Test_Handler_PromptAndProcessReleaseVersionChoiceInput(t *testing.T) {

	suggestedLastVersion := "v3.1.4"
	suggestedNextVersion := "v3.2.4"

	cases := map[string]struct {
		expectError                    bool
		firstPromptAnswer              string
		secondPromptAnswer             string
		thirdPromptAnswer              string
		expectedPreviousReleaseVersion string
		expectedReleaseVersion         string
	}{
		"accepting suggested versions": {
			firstPromptAnswer:              "y\n",
			expectedPreviousReleaseVersion: suggestedLastVersion,
			expectedReleaseVersion:         suggestedNextVersion,
		},
		"not accepting suggested versions, use provided values": {
			firstPromptAnswer:              "n\n",
			secondPromptAnswer:             "v9.9.0\n",
			thirdPromptAnswer:              "v9.10.0\n",
			expectedPreviousReleaseVersion: "v9.9.0",
			expectedReleaseVersion:         "v9.10.0",
		},
		"error on bad y/n input": {
			firstPromptAnswer: "foobar\n",
			expectError:       true,
		},
		"not accepting suggested versions, get error on bad second input value": {
			firstPromptAnswer:  "n\n",
			secondPromptAnswer: "9.9.0\n", // no v
			expectError:        true,
		},
		"not accepting suggested versions, get error on bad third input value": {
			firstPromptAnswer:  "n\n",
			secondPromptAnswer: "v9.9.0\n",
			thirdPromptAnswer:  "9.10.0\n", // no v
			expectError:        true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			stdin := bytes.Buffer{}

			userInput := tc.firstPromptAnswer + tc.secondPromptAnswer + tc.thirdPromptAnswer
			stdin.Write([]byte(userInput))

			input := Input{}
			handler := NewHandler(&input)

			r := bufio.NewReader(&stdin)
			handler.reader = r

			err := handler.PromptAndProcessReleaseVersionChoiceInput(suggestedLastVersion, suggestedNextVersion)
			if err != nil && !tc.expectError {
				t.Fatal(err.Error())
			}
			if err == nil && tc.expectError {
				t.Fatal("expected error but got none")
			}

			if input.PreviousReleaseVersion != tc.expectedPreviousReleaseVersion {
				t.Fatalf("wanted %s, got %s", tc.expectedPreviousReleaseVersion, input.PreviousReleaseVersion)
			}
			if input.ReleaseVersion != tc.expectedReleaseVersion {
				t.Fatalf("wanted %s, got %s", tc.expectedReleaseVersion, input.ReleaseVersion)
			}
		})
	}
}
