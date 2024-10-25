package encryption

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptLoop(t *testing.T) {
	public, private, err := GenerateKeyPair(nil)
	require.NoError(t, err)

	value := "some-secret-value"

	encrypted, err := EncryptValue(value, public)
	require.NoError(t, err)

	output, err := DecryptValue(encrypted, private)
	require.NoError(t, err)

	require.Equal(t, value, output)
}
