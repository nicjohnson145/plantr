package token

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenLoop(t *testing.T) {
	token := Token{
		NodeID: "some-id",
	}
	key := []byte(`some-big-uuid`)

	tokenStr, err := GenerateJWT(key, token)
	require.NoError(t, err)

	outToken, err := ParseJWT(tokenStr, key)
	require.NoError(t, err)

	require.Equal(t, &token, outToken)
}
