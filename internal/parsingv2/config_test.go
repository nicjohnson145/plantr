package parsingv2

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFS(t *testing.T) {
	t.Run("smokes", func(t *testing.T) {
		_, err := ParseFS(os.DirFS("./testdata/basic"))
		require.NoError(t, err)
	})
}
