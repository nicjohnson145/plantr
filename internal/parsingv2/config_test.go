package parsingv2

import (
	"encoding/base64"
	"os"
	"testing"

	configv1 "github.com/nicjohnson145/plantr/gen/plantr/config/v1"
	"github.com/psanford/memfs"
	"github.com/stretchr/testify/require"
)

func TestParseFS(t *testing.T) {
	t.Parallel()

	t.Run("smokes", func(t *testing.T) {
		t.Parallel()
		_, err := ParseFS(os.DirFS("./testdata/basic"))
		require.NoError(t, err)
	})
}

func TestConfigFile(t *testing.T) {
	t.Parallel()

	fsys := memfs.New()
	require.NoError(t, fsys.WriteFile("some-config", []byte(`some config content`), 0664))

	valid := func() *configv1.ConfigFile {
		return &configv1.ConfigFile{
			Path:        "some-config",
			Destination: "def",
		}
	}

	testData := []struct {
		name    string
		modFunc func(c *configv1.ConfigFile)
		err     string
	}{
		{
			name:    "valid",
			modFunc: func(c *configv1.ConfigFile) {},
			err:     "",
		},
		{
			name: "no path",
			modFunc: func(c *configv1.ConfigFile) {
				c.Path = ""
			},
			err: "path is a required field",
		},
		{
			name: "no destination",
			modFunc: func(c *configv1.ConfigFile) {
				c.Destination = ""
			},
			err: "destination is a required field",
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validConf := valid()
			tc.modFunc(validConf)

			_, err := parseSeed_configFile(fsys, validConf)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestNode(t *testing.T) {
	t.Parallel()

	valid := func() *configv1.Node {
		return &configv1.Node{
			Id:             "some-id",
			PublicKeyB64:   base64.StdEncoding.EncodeToString([]byte(`some-key`)),
			UserHome:       "user-home",
			Os:             "linux",
			Arch:           "amd64",
			PackageManager: "apt",
		}
	}

	testData := []struct {
		name    string
		modFunc func(x *configv1.Node)
		err     string
	}{
		{
			name:    "valid",
			modFunc: func(x *configv1.Node) {},
			err:     "",
		},
		{
			name: "no id",
			modFunc: func(x *configv1.Node) {
				x.Id = ""
			},
			err: "id is a required field",
		},
		{
			name: "no public key",
			modFunc: func(x *configv1.Node) {
				x.PublicKeyB64 = ""
			},
			err: "public_key_b64 is a required field",
		},
		{
			name: "no user home",
			modFunc: func(x *configv1.Node) {
				x.UserHome = ""
			},
			err: "user_home is a required field",
		},
		{
			name: "no os",
			modFunc: func(x *configv1.Node) {
				x.Os = ""
			},
			err: `os is required to be one of ["linux", "darwin"]`,
		},
		{
			name: "invalid os",
			modFunc: func(x *configv1.Node) {
				x.Os = "not-an-os"
			},
			err: `os is required to be one of ["linux", "darwin"]`,
		},
		{
			name: "no arch",
			modFunc: func(x *configv1.Node) {
				x.Arch = ""
			},
			err: `arch is required to be one of ["amd64", "arm64"]`,
		},
		{
			name: "invalid arch",
			modFunc: func(x *configv1.Node) {
				x.Arch = "fake-arch"
			},
			err: `arch is required to be one of ["amd64", "arm64"]`,
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validObj := valid()
			tc.modFunc(validObj)

			_, err := parseNode(validObj)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestGithubRelease(t *testing.T) {
	t.Parallel()

	valid := func() *configv1.GithubRelease {
		return &configv1.GithubRelease{
			Repo: "some/repo",
			Tag:  "v1.0.0",
		}
	}

	testData := []struct {
		name    string
		modFunc func(x *configv1.GithubRelease)
		err     string
	}{
		{
			name:    "valid",
			modFunc: func(x *configv1.GithubRelease) {},
			err:     "",
		},
		{
			name: "no repo",
			modFunc: func(x *configv1.GithubRelease) {
				x.Repo = ""
			},
			err: "repo is a required field",
		},
		{
			name: "no tag",
			modFunc: func(x *configv1.GithubRelease) {
				x.Tag = ""
			},
			err: "tag is a required field",
		},
		{
			name: "invalid regex",
			modFunc: func(x *configv1.GithubRelease) {
				x.AssetPatterns = &configv1.GithubRelease_AssetPattern{
					Linux: &configv1.GithubRelease_AssetPattern_ArchPattern{
						Amd64: `[a-z`,
					},
				}
			},
			err: "invalid regex: error parsing regex for linux/amd64",
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validObj := valid()
			tc.modFunc(validObj)

			_, err := parseSeed_githubRelease(validObj)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestSystemPackage(t *testing.T) {
	t.Parallel()

	valid := func() *configv1.SystemPackage {
		return &configv1.SystemPackage{
			Apt: &configv1.SystemPackage_Apt{
				Name: "some-apt-package",
			},
		}
	}

	testData := []struct {
		name    string
		modFunc func(x *configv1.SystemPackage)
		err     string
	}{
		{
			name:    "valid",
			modFunc: func(x *configv1.SystemPackage) {},
			err:     "",
		},
		{
			name: "no apt name",
			modFunc: func(x *configv1.SystemPackage) {
				x.Apt.Name = ""
			},
			err: "name is a required field",
		},
		{
			name: "no top level keys",
			modFunc: func(x *configv1.SystemPackage) {
				x.Apt = nil
			},
			err: "at least one of ['apt'] is required",
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validObj := valid()
			tc.modFunc(validObj)

			_, err := parseSeed_systemPackage(validObj)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestGitRepo(t *testing.T) {
	t.Parallel()

	valid := func() *configv1.GitRepo {
		return &configv1.GitRepo{
			Url:      "some-url",
			Location: "some-location",
			Ref: &configv1.GitRepo_Tag{
				Tag: "some-tag",
			},
		}
	}

	testData := []struct {
		name    string
		modFunc func(x *configv1.GitRepo)
		err     string
	}{
		{
			name:    "valid",
			modFunc: func(x *configv1.GitRepo) {},
			err:     "",
		},
		{
			name: "valid commit",
			modFunc: func(x *configv1.GitRepo) {
				x.Ref = &configv1.GitRepo_Commit{
					Commit: "some-commit",
				}
			},
			err: "",
		},
		{
			name: "no url",
			modFunc: func(x *configv1.GitRepo) {
				x.Url = ""
			},
			err: "url is a required field",
		},
		{
			name: "no location",
			modFunc: func(x *configv1.GitRepo) {
				x.Location = ""
			},
			err: "location is a required field",
		},
		{
			name: "no ref",
			modFunc: func(x *configv1.GitRepo) {
				x.Ref = nil
			},
			err: "ref: exactly one field is required in oneof",
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validObj := valid()
			tc.modFunc(validObj)

			_, err := parseSeed_gitRepo(validObj)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.err)
			}
		})
	}
}
