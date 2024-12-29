package controller

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/nicjohnson145/plantr/internal/parsingv2"
)

func seedHash(x *parsingv2.Seed) string {
	var parts []string

	switch concrete := x.Element.(type) {
	case *parsingv2.ConfigFile:
		parts = []string{
			"ConfigFile",
			concrete.TemplateContent,
			concrete.Destination,
		}
	case *parsingv2.GithubRelease:
		parts = []string{
			"GithubRelease",
			concrete.Repo,
			concrete.Tag,
		}
	case *parsingv2.SystemPackage:
		parts = []string{
			"SystemPackage",
		}
		if concrete.Apt != nil {
			parts = append(
				parts,
				"APT",
				concrete.Apt.Name,
			)
		}
	default:
		panic(fmt.Sprintf("unhandled seed type %T", concrete))
	}

	return fmt.Sprint(md5.Sum([]byte(strings.Join(parts, "")))) //nolint: gosec // its a hash, it doesnt have to be cryptographically secure
}

