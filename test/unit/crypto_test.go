package unit_test

import (
	"testing"

	"github.com/yourorg/boardgov/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword_And_Check(t *testing.T) {
	hash, err := crypto.HashPassword("SecurePass123!")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, crypto.CheckPassword(hash, "SecurePass123!"))
	assert.False(t, crypto.CheckPassword(hash, "WrongPass"))
}

func TestSHA256File(t *testing.T) {
	data := []byte("hello world")
	hash := crypto.SHA256File(data)
	assert.Len(t, hash, 64)
	assert.Equal(t, hash, crypto.SHA256File(data))
}

func TestAESEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	for i := range key { key[i] = byte(i) }
	plaintext := []byte("top secret board document")
	ciphertext, err := crypto.AESEncrypt(key, plaintext)
	require.NoError(t, err)
	decrypted, err := crypto.AESDecrypt(key, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestRandomHex(t *testing.T) {
	h1, err := crypto.RandomHex(32)
	require.NoError(t, err)
	h2, err := crypto.RandomHex(32)
	require.NoError(t, err)
	assert.Len(t, h1, 64)
	assert.NotEqual(t, h1, h2)
}
