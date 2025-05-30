package controller

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGithub_ParseUrl(t *testing.T) {
	t.Parallel()

	testData := []struct {
		name  string
		url   string
		owner string
		repo  string
	}{
		{
			name:  "ssh",
			url:   "git@github.com:nicjohnson145/plantr.git",
			owner: "nicjohnson145",
			repo:  "plantr",
		},
		{
			name:  "https",
			url:   "https://github.com/nicjohnson145/plantr.git",
			owner: "nicjohnson145",
			repo:  "plantr",
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			owner, repo, err := (&GithubGitClient{}).parseUrl(tc.url)
			require.NoError(t, err)
			require.Equal(t, tc.owner, owner)
			require.Equal(t, tc.repo, repo)
		})
	}
}
