package release_version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/mod/semver"
)

type ReleaseQuery struct {
	client        *http.Client
	owner         string
	repo          string
	latestRelease string
}

type LatestReleaseResp struct {
	TagName string `json:"tag_name"`
}

func New(owner, repo string) ReleaseQuery {
	return ReleaseQuery{
		client: &http.Client{Timeout: 10 * time.Second},
		owner:  owner,
		repo:   repo,
	}
}

func (c *ReleaseQuery) GetLastVersionFromGitHub() (string, error) {
	// return result from previous run, if present
	if c.latestRelease != "" {
		return c.latestRelease, nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", c.owner, c.repo)
	resp, err := c.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("error getting latest release from github.com/%s/%s : %w", c.owner, c.repo, err)
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		return "", fmt.Errorf("got a non-200 response: status '%s', body '%s'", resp.Status, data)
	}

	defer resp.Body.Close()

	data := LatestReleaseResp{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", fmt.Errorf("error parsing response body : %w", err)
	}

	// memo
	c.latestRelease = data.TagName

	return data.TagName, nil
}

func NextMinorVersion(latestVersion string) (string, error) {
	if !semver.IsValid(latestVersion) {
		return "", fmt.Errorf("invalid version provided: %s", latestVersion)
	}

	semverRECaps := `^v(?P<major>\d{1,3})\.(?P<minor>\d{1,3})\.(?P<patch>\d{1,3})$`
	re := regexp.MustCompile(semverRECaps)
	matches := re.FindStringSubmatch(latestVersion)

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	nextMinor := minor + 1

	return fmt.Sprintf("v%d.%d.0", major, nextMinor), nil
}
