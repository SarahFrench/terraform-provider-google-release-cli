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
			new:       "2.2.2",
			old:       "1.1.1",
			expectErr: false,
		},
		"an error is returned if all inputs are empty strings": {
			new:       "",
			old:       "",
			expectErr: true,
		},
		"an error is returned if an input is incorrectly prepended by a v character": {
			new:       "v2.2.2",
			old:       "1.1.1",
			expectErr: true,
		},
		"an error is returned is both versions are the same": {
			new:       "1.1.1",
			old:       "1.1.1",
			expectErr: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			errors := validateVersionInputs(tc.new, tc.old)

			if len(errors) > 0 && !tc.expectErr {
				t.Fatalf("encountered errors when none were expected: %v", errors)
			}
			if len(errors) == 0 && tc.expectErr {
				t.Fatalf("expected errors but none were returned from the function")
			}
		})
	}
}
