package input

import (
	"testing"
)

func Test_ValidateVersionInputs(t *testing.T) {
	cases := map[string]struct {
		new       string
		old       string
		expectErr bool
	}{
		"good input": {
			new:       "v2.2.2",
			old:       "v1.1.1",
			expectErr: false,
		},
		"an error is returned if all inputs are empty strings": {
			new:       "",
			old:       "",
			expectErr: true,
		},
		"an error is returned if an input is missing \"v\" prefix": {
			new:       "v2.2.2",
			old:       "1.1.1",
			expectErr: true,
		},
		"an error is returned is both versions are the same": {
			new:       "v1.1.1",
			old:       "v1.1.1",
			expectErr: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			err := validateVersionInputs(tc.new, tc.old)

			if err != nil && !tc.expectErr {
				t.Fatalf("encountered errors when none were expected: %v", err)
			}
			if err == nil && tc.expectErr {
				t.Fatalf("expected errors but none were returned from the function")
			}
		})
	}
}
