package release_version

import (
	"net/http"
	"testing"
	"time"

	"golang.org/x/mod/semver"
)

func TestGetLastVersionFromGitHub(t *testing.T) {
	t.Run("can GET the latest version of the terraform-provider-google repo", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}
		c := ReleaseQuery{
			client: client,
		}
		ver, err := c.GetLastVersionFromGitHub("terraform-provider-google")
		if err != nil {
			t.Fatalf("unexpected error(s) encountered: %v", err)
		}

		if !semver.IsValid(ver) {
			t.Fatalf("expected a valid semver returned for the latest version, got: %s", ver)
		}
	})
}

func TestNextMinorVersion(t *testing.T) {
	t.Run("can suggest the next minor version as the next version to release", func(t *testing.T) {
		ver, err := NextMinorVersion("v1.2.3")
		if err != nil {
			t.Fatalf("unexpected error(s) encountered: %v", err)
		}

		if ver != "v1.3.0" {
			t.Fatalf("expected v1.3.0, got: %s", ver)
		}
	})
}
