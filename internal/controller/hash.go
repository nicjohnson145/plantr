package controller

import (
	"crypto/md5" //nolint:gosec // its for fingerprinting, it doesnt have to be cryptographically secure
	"fmt"
	"sort"
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
	case *parsingv2.GitRepo:
		parts = []string{
			"GitRepo",
			concrete.URL,
			concrete.Location,
		}
		if concrete.Tag != nil {
			parts = append(parts, *concrete.Tag)
		}
		if concrete.Commit != nil {
			parts = append(parts, *concrete.Commit)
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

func sortSeeds(seeds []*parsingv2.Seed) {
	sort.Slice(seeds, func(i, j int) bool {
		return strings.Compare(seedHash(seeds[i]), seedHash(seeds[j])) < 0
	})
}
