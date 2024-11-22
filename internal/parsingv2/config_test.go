package parsingv2

import (
	"os"
	"testing"

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
