package controller

import (
	"crypto/md5" //nolint:gosec // its for fingerprinting, it doesnt have to be cryptographically secure
	"fmt"
	"slices"
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
	case *parsingv2.Golang:
		parts = []string{
			"Golang",
			concrete.Version,
		}
	case *parsingv2.GoInstall:
		parts = []string{
			"GoInstall",
			concrete.Package,
		}
		if concrete.Version != nil {
			parts = append(parts, *concrete.Version)
		}
	default:
		panic(fmt.Sprintf("unhandled seed type %T", concrete))
	}

	return fmt.Sprint(md5.Sum([]byte(strings.Join(parts, "")))) //nolint: gosec // its a hash, it doesnt have to be cryptographically secure
}

func sortSeeds(seeds []*parsingv2.Seed) {
	ordering := []string{
		fmt.Sprintf("%T", &parsingv2.ConfigFile{}),
		fmt.Sprintf("%T", &parsingv2.GitRepo{}),
		fmt.Sprintf("%T", &parsingv2.GithubRelease{}),
		fmt.Sprintf("%T", &parsingv2.SystemPackage{}),
		fmt.Sprintf("%T", &parsingv2.Golang{}),
		fmt.Sprintf("%T", &parsingv2.GoInstall{}),
	}
	orderMap := map[string]int{}
	for idx, val := range ordering {
		orderMap[val] = idx
	}

	slices.SortFunc(seeds, func(a *parsingv2.Seed, b *parsingv2.Seed) int {
		aVal := orderMap[fmt.Sprintf("%T", a.Element)]
		bVal := orderMap[fmt.Sprintf("%T", b.Element)]

		if aVal != bVal {
			if aVal < bVal {
				return -1
			} else {
				return 1
			}
		}

		return strings.Compare(seedHash(a), seedHash(b))
	})
}
