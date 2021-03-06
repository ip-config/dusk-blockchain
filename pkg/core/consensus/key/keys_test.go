package key

import (
	"io"
	"testing"

	crypto "github.com/dusk-network/dusk-crypto/hash"
	"github.com/stretchr/testify/assert"
)

// Test that NewRandKeys does not ever return an EOF error
func TestNewRandKeys(t *testing.T) {
	keys, err := NewRandKeys()
	if err == io.EOF {
		t.Fatal("got EOF: NewRandKeys should never give an EOF error")
	}

	assert.NoError(t, err)
	assert.NotNil(t, keys)
}

func TestDeterministicKeyGen(t *testing.T) {
	var firstKeyPair, secondKeyPair Keys
	for {
		seed, err := crypto.RandEntropy(128)
		assert.Nil(t, err)

		firstKeyPair, err = NewKeysFromBytes(seed)
		if err == io.EOF {
			continue
		}
		assert.Nil(t, err)

		secondKeyPair, err = NewKeysFromBytes(seed)
		if err == io.EOF {
			continue
		}
		assert.Nil(t, err)

		break
	}

	assert.Equal(t, firstKeyPair.BLSPubKeyBytes, secondKeyPair.BLSPubKeyBytes)
	assert.Equal(t, firstKeyPair.BLSSecretKey, secondKeyPair.BLSSecretKey)

}
