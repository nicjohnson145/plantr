package controller

import (
	"fmt"
	"testing"

	"github.com/nicjohnson145/hlp"
	"github.com/nicjohnson145/plantr/internal/parsingv2"
	"github.com/stretchr/testify/require"
)

func TestGoInstallsLast(t *testing.T) {
	seeds := []*parsingv2.Seed{
		{
			Element: &parsingv2.GoInstall{
				Package: "abc",
			},
		},
		{
			Element: &parsingv2.Golang{
				Version: "123",
			},
		},
		{
			Element: &parsingv2.SystemPackage{
				Apt: &parsingv2.SystemPackageApt{
					Name: "pkg",
				},
			},
		},
	}
	sortSeeds(seeds)

	types := hlp.Map(seeds, func(item *parsingv2.Seed, _ int) string {
		return fmt.Sprintf("%T", item.Element)
	})

	require.Equal(
		t,
		[]string{
			"*parsingv2.SystemPackage",
			"*parsingv2.Golang",
			"*parsingv2.GoInstall",
		},
		types,
	)
}
