package parsingv2

import (
	"encoding/base64"
	"os"
	"testing"

	"github.com/nicjohnson145/hlp"
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
		{
			name: "no package manager",
			modFunc: func(x *configv1.Node) {
				x.PackageManager = ""
			},
			err: `package_manager is required to be one of ["apt", "brew"]`,
		},
		{
			name: "wrong package manager",
			modFunc: func(x *configv1.Node) {
				x.PackageManager = "fake-guy"
			},
			err: `package_manager is required to be one of ["apt", "brew"]`,
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
			Brew: &configv1.SystemPackage_Brew{
				Name: "some-brew-package",
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
				x.Brew = nil 
			},
			err: "at least one of ['apt', 'brew'] is required",
		},
		{
			name: "no brew name",
			modFunc: func(x *configv1.SystemPackage) {
				x.Brew.Name = ""
			},
			err: "name is a required field",
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

func TestGolang(t *testing.T) {
	t.Parallel()

	valid := func() *configv1.Golang {
		return &configv1.Golang{
			Version: "some-version",
		}
	}

	testData := []struct {
		name    string
		modFunc func(x *configv1.Golang)
		err     string
	}{
		{
			name:    "valid",
			modFunc: func(x *configv1.Golang) {},
			err:     "",
		},
		{
			name: "no version",
			modFunc: func(x *configv1.Golang) {
				x.Version = ""
			},
			err: "version is a required field",
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validObj := valid()
			tc.modFunc(validObj)

			_, err := parseRole(nil, []*configv1.Seed{{
				Element: &configv1.Seed_Golang{
					Golang: validObj,
				},
			}})
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestGoInstall(t *testing.T) {
	t.Parallel()

	valid := func() *configv1.GoInstall {
		return &configv1.GoInstall{
			Package: "some-package",
		}
	}

	testData := []struct {
		name    string
		modFunc func(x *configv1.GoInstall)
		err     string
	}{
		{
			name:    "valid",
			modFunc: func(x *configv1.GoInstall) {},
			err:     "",
		},
		{
			name: "valid with version",
			modFunc: func(x *configv1.GoInstall) {
				x.Version = hlp.Ptr("some-version")
			},
			err: "",
		},
		{
			name: "no package",
			modFunc: func(x *configv1.GoInstall) {
				x.Package = ""
			},
			err: "package is a required field",
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validObj := valid()
			tc.modFunc(validObj)

			_, err := parseRole(nil, []*configv1.Seed{{
				Element: &configv1.Seed_GoInstall{
					GoInstall: validObj,
				},
			}})
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestUrlDownload(t *testing.T) {
	t.Parallel()

	valid := func() *configv1.UrlDownload {
		return &configv1.UrlDownload{
			Urls: &configv1.UrlDownload_OsGroup{
				Linux: &configv1.UrlDownload_OsGroup_ArchGroup{
					Amd64: hlp.Ptr("some-linux-amd64-url"),
				},
				Mac: &configv1.UrlDownload_OsGroup_ArchGroup{
					Amd64: hlp.Ptr("some-mac-amd64-url"),
				},
			},
		}
	}

	testData := []struct {
		name    string
		modFunc func(x *configv1.UrlDownload)
		err     string
	}{
		{
			name:    "valid",
			modFunc: func(x *configv1.UrlDownload) {},
			err:     "",
		},
		{
			name:    "valid single url",
			modFunc: func(x *configv1.UrlDownload) {
				x.Urls.Mac = nil
			},
			err:     "",
		},
		{
			name:    "no urls",
			modFunc: func(x *configv1.UrlDownload) {
				x.Urls.Mac = nil
				x.Urls.Linux = nil
			},
			err:     "must specify at least one OS/Arch url",
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validObj := valid()
			tc.modFunc(validObj)

			_, err := parseRole(nil, []*configv1.Seed{{
				Element: &configv1.Seed_UrlDownload{
					UrlDownload: validObj,
				},
			}})
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.err)
			}
		})
	}
}
