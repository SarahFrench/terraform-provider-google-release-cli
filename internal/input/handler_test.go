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
