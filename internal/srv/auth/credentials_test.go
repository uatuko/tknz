package auth

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/felk-ai/idaas/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialPublicKeyUnmarshalCBOR(t *testing.T) {
	// EC2: generate ECDSA P-256 key pair
	ec2Key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	xEC2 := ec2Key.PublicKey.X.Bytes()
	yEC2 := ec2Key.PublicKey.Y.Bytes()

	// OKP: generate Ed25519 key pair
	okpPub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	xOKP := []byte(okpPub)

	// RSA: generate RSA key pair
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	nRSA := rsaKey.PublicKey.N.Bytes()
	eRSA := big.NewInt(int64(rsaKey.PublicKey.E)).Bytes()

	t.Run("EC2", func(t *testing.T) {
		data := coseKey(t, map[int]any{
			1:  int(coseKtyEC2),
			3:  int(coseAlgES256),
			-1: int(coseCrvP256),
			-2: xEC2,
			-3: yEC2,
		})
		var k credentialPublicKey
		require.NoError(t, k.UnmarshalCBOR(data))
		assert.Equal(t, coseKtyEC2, k.Kty)
		assert.Equal(t, coseAlgES256, k.Alg)
		assert.Equal(t, coseCrvP256, k.Crv)
		assert.Equal(t, xEC2, k.X)
		assert.Equal(t, yEC2, k.Y)
		assert.Nil(t, k.N)
		assert.Nil(t, k.E)
	})

	t.Run("OKP", func(t *testing.T) {
		data := coseKey(t, map[int]any{
			1:  int(coseKtyOKP),
			3:  int(coseAlgEdDSA),
			-1: int(coseCrvEd25519),
			-2: xOKP,
		})
		var k credentialPublicKey
		require.NoError(t, k.UnmarshalCBOR(data))
		assert.Equal(t, coseKtyOKP, k.Kty)
		assert.Equal(t, coseAlgEdDSA, k.Alg)
		assert.Equal(t, coseCrvEd25519, k.Crv)
		assert.Equal(t, xOKP, k.X)
		assert.Nil(t, k.Y)
		assert.Nil(t, k.N)
		assert.Nil(t, k.E)
	})

	t.Run("RSA", func(t *testing.T) {
		data := coseKey(t, map[int]any{
			1:  int(coseKtyRSA),
			3:  int(coseAlgRS256),
			-1: nRSA,
			-2: eRSA,
		})
		var k credentialPublicKey
		require.NoError(t, k.UnmarshalCBOR(data))
		assert.Equal(t, coseKtyRSA, k.Kty)
		assert.Equal(t, coseAlgRS256, k.Alg)
		assert.Equal(t, nRSA, k.N)
		assert.Equal(t, eRSA, k.E)
		assert.Nil(t, k.X)
		assert.Nil(t, k.Y)
	})

	t.Run("error: missing kty", func(t *testing.T) {
		data := coseKey(t, map[int]any{3: int(coseAlgES256)})
		var k credentialPublicKey
		require.ErrorContains(t, k.UnmarshalCBOR(data), "missing cose key type")
	})

	t.Run("error: missing alg", func(t *testing.T) {
		data := coseKey(t, map[int]any{1: int(coseKtyEC2)})
		var k credentialPublicKey
		require.ErrorContains(t, k.UnmarshalCBOR(data), "missing cose algorithm")
	})

	t.Run("error: unsupported kty", func(t *testing.T) {
		data := coseKey(t, map[int]any{1: 99, 3: int(coseAlgES256)})
		var k credentialPublicKey
		require.ErrorContains(t, k.UnmarshalCBOR(data), "unsupported cose key type")
	})

	t.Run("error: EC2 missing crv", func(t *testing.T) {
		data := coseKey(t, map[int]any{1: int(coseKtyEC2), 3: int(coseAlgES256), -2: xEC2, -3: yEC2})
		var k credentialPublicKey
		require.ErrorContains(t, k.UnmarshalCBOR(data), "missing curve identifier")
	})

	t.Run("error: EC2 missing x", func(t *testing.T) {
		data := coseKey(t, map[int]any{1: int(coseKtyEC2), 3: int(coseAlgES256), -1: int(coseCrvP256), -3: yEC2})
		var k credentialPublicKey
		require.ErrorContains(t, k.UnmarshalCBOR(data), "missing x-coordinate or public key")
	})

	t.Run("error: EC2 missing y", func(t *testing.T) {
		data := coseKey(t, map[int]any{1: int(coseKtyEC2), 3: int(coseAlgES256), -1: int(coseCrvP256), -2: xEC2})
		var k credentialPublicKey
		require.ErrorContains(t, k.UnmarshalCBOR(data), "missing y-coordinate")
	})
}

func TestPublicKeyCredentialDecode(t *testing.T) {
	b := []byte(`{"authenticatorAttachment":"platform","clientExtensionResults":{},"id":"ZoqmX0VWLDVEOlax0iFMmw","rawId":"ZoqmX0VWLDVEOlax0iFMmw","response":{"attestationObject":"o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YViUSZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NdAAAAAOqbjWZNAR0hPOS2tIy1ddQAEGaKpl9FViw1RDpWsdIhTJulAQIDJiABIVggI9AMkTahPSbtce1rQzHm4CQl4l71KAbaf1kvx6HQgPgiWCAxV_97NEbxARLWIAidv1PSmJKvBLW5wgi23v3b4JKkLA","authenticatorData":"SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NdAAAAAOqbjWZNAR0hPOS2tIy1ddQAEGaKpl9FViw1RDpWsdIhTJulAQIDJiABIVggI9AMkTahPSbtce1rQzHm4CQl4l71KAbaf1kvx6HQgPgiWCAxV_97NEbxARLWIAidv1PSmJKvBLW5wgi23v3b4JKkLA","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoidHRiWnNJNmhnUi02UXo5emk3QVlnQSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6NTE3NCIsImNyb3NzT3JpZ2luIjpmYWxzZSwib3RoZXJfa2V5c19jYW5fYmVfYWRkZWRfaGVyZSI6ImRvIG5vdCBjb21wYXJlIGNsaWVudERhdGFKU09OIGFnYWluc3QgYSB0ZW1wbGF0ZS4gU2VlIGh0dHBzOi8vZ29vLmdsL3lhYlBleCJ9","publicKey":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEI9AMkTahPSbtce1rQzHm4CQl4l71KAbaf1kvx6HQgPgxV_97NEbxARLWIAidv1PSmJKvBLW5wgi23v3b4JKkLA","publicKeyAlgorithm":-7,"transports":["hybrid","internal"]},"type":"public-key"}`)

	var cred publicKeyCredential
	require.NoError(t, json.Unmarshal(b, &cred))

	// Basic assertions for public key credential, no need to test json unmarshal
	assert.Equal(t, "platform", cred.AuthenticatorAttachment)
	assert.Equal(t, "ZoqmX0VWLDVEOlax0iFMmw", cred.Id)
	assert.Equal(t, db.Base64Url{0x66, 0x8a, 0xa6, 0x5f, 0x45, 0x56, 0x2c, 0x35, 0x44, 0x3a, 0x56, 0xb1, 0xd2, 0x21, 0x4c, 0x9b}, cred.RawId)
	assert.Equal(t, "public-key", cred.Type)

	var attObj attestationObject
	require.NoError(t, cborDecoder().Unmarshal(cred.Response.AttestationObject, &attObj))

	assert.Equal(t, "none", attObj.Fmt) // attestation format
	assert.Empty(t, attObj.AttStmt)     // attestation statement

	// Auth data
	authData := decodeAuthenticatorData(attObj.AuthData)
	require.NotNil(t, authData)

	// (AuthData) rp id hash
	assert.Equal(t, [32]byte{0x49, 0x96, 0xd, 0xe5, 0x88, 0xe, 0x8c, 0x68, 0x74, 0x34, 0x17, 0xf, 0x64, 0x76, 0x60, 0x5b, 0x8f, 0xe4, 0xae, 0xb9, 0xa2, 0x86, 0x32, 0xc7, 0x99, 0x5c, 0xf3, 0xba, 0x83, 0x1d, 0x97, 0x63}, authData.RpIdHash)

	assert.Equal(t, byte(0x5d), authData.Flags)
	assert.Equal(t, uint32(0), authData.SignCount)
	assert.Nil(t, authData.Extensions)

	// Attested credentials
	require.NotNil(t, authData.AttestedCredentialData)
	atd := authData.AttestedCredentialData

	// (ATD) aaguid
	assert.Equal(t, [16]uint8{0xea, 0x9b, 0x8d, 0x66, 0x4d, 0x1, 0x1d, 0x21, 0x3c, 0xe4, 0xb6, 0xb4, 0x8c, 0xb5, 0x75, 0xd4}, atd.Aaguid)

	// (ATD) credential id
	assert.Equal(t, uint16(16), atd.CredentialIdLength)
	assert.Equal(t, []byte{0x66, 0x8a, 0xa6, 0x5f, 0x45, 0x56, 0x2c, 0x35, 0x44, 0x3a, 0x56, 0xb1, 0xd2, 0x21, 0x4c, 0x9b}, atd.CredentialId)

	// (ATD) credential public key
	assert.Len(t, atd.credentialPublicKeyRaw, 77)
	require.NotNil(t, atd.CredentialPublicKey)
	cpk := atd.CredentialPublicKey

	assert.Equal(t, coseKtyEC2, cpk.Kty)
	assert.Equal(t, coseAlgES256, cpk.Alg)
	assert.Equal(t, coseCrvP256, cpk.Crv)
	assert.Nil(t, cpk.N)
	assert.Nil(t, cpk.E)

	// (CPK) x-coordinates
	assert.Equal(t, []byte{0x23, 0xd0, 0xc, 0x91, 0x36, 0xa1, 0x3d, 0x26, 0xed, 0x71, 0xed, 0x6b, 0x43, 0x31, 0xe6, 0xe0, 0x24, 0x25, 0xe2, 0x5e, 0xf5, 0x28, 0x6, 0xda, 0x7f, 0x59, 0x2f, 0xc7, 0xa1, 0xd0, 0x80, 0xf8}, cpk.X)

	// (CPK) x-coordinates
	assert.Equal(t, []byte{0x31, 0x57, 0xff, 0x7b, 0x34, 0x46, 0xf1, 0x1, 0x12, 0xd6, 0x20, 0x8, 0x9d, 0xbf, 0x53, 0xd2, 0x98, 0x92, 0xaf, 0x4, 0xb5, 0xb9, 0xc2, 0x8, 0xb6, 0xde, 0xfd, 0xdb, 0xe0, 0x92, 0xa4, 0x2c}, cpk.Y)
}

// coseKey builds a CBOR-encoded COSE_Key map from the given labels.
func coseKey(t *testing.T, fields map[int]any) []byte {
	t.Helper()
	b, err := cborEncoder().Marshal(fields)
	require.NoError(t, err)
	return b
}
