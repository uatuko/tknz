package auth

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/rs/zerolog"
	"go.tknz.dev/internal/db"
	"go.tknz.dev/internal/srv/common"
)

const (
	// Auth data flags (Ref: https://w3c.github.io/webauthn/#authdata-flags)
	authDataFlagUP   byte = 1 << 0 // User present
	authDataFlagFRU1 byte = 1 << 1 // Reserved for future use
	authDataFlagUV   byte = 1 << 2 // User verified
	authDataFlagBE   byte = 1 << 3 // Backup eligibility
	authDataFlagBS   byte = 1 << 4 // Backup state
	authDataFlagRFU2 byte = 1 << 5 // Reserved for future use
	authDataFlagAT   byte = 1 << 6 // Attested credential data included
	authDataFlagED   byte = 1 << 7 // Extension data included

	// COSE algorithms relevant to WebAuthn (Ref: https://w3c.github.io/webauthn/#sctn-alg-identifier)
	coseAlgES256  coseAlg = -7   // ECDSA w/ SHA-256
	coseAlgESP256 coseAlg = -9   // ECDSA using P-256 curve and SHA-256
	coseAlgES384  coseAlg = -35  // ECDSA w/ SHA-384
	coseAlgESP384 coseAlg = -51  // ECDSA using P-384 curve and SHA-384
	coseAlgES512  coseAlg = -36  // ECDSA w/ SHA-512
	coseAlgESP512 coseAlg = -52  // ECDSA using P-521 curve and SHA-512
	coseAlgEdDSA  coseAlg = -8   // Ed25519
	coseAlgRS256  coseAlg = -257 // RSASSA-PKCS1-v1_5 w/ SHA-256

	// COSE elliptic curves (Ref: https://www.rfc-editor.org/rfc/rfc9053#section-7.1)
	coseCrvP256    coseCrv = 1 // NIST P-256
	coseCrvP384    coseCrv = 2 // NIST P-384
	coseCrvP521    coseCrv = 3 // NIST P-521
	coseCrvEd25519 coseCrv = 6 // Ed25519

	// COSE key types (Ref: https://www.rfc-editor.org/rfc/rfc9053#section-7)
	coseKtyOKP coseKty = 1 // Octet Key Pair
	coseKtyEC2 coseKty = 2 // Elliptic Curve keys w/ x/y-coordinate pair
	coseKtyRSA coseKty = 3 // RSA Key

	clientDataTypeWebauthnCreate string = "webauthn.create"
	clientDataTypeWebauthnGet    string = "webauthn.get"
	credentialTypePublicKey      string = "public-key"

	maxCredentialIdLength      = 1023
	minAuthenticatorDataLength = 32 + 1 + 4
)

type coseAlg int
type coseCrv int
type coseKty int

// Ref: https://w3c.github.io/webauthn/#iface-authenticatorresponse
type authenticatorResponse struct {
	AuthenticatorData db.Base64Url `json:"authenticatorData"`
	ClientDataJSON    db.Base64Url `json:"clientDataJSON"`

	// Attestation
	// Ref: https://w3c.github.io/webauthn/#iface-authenticatorattestationresponse
	AttestationObject  db.Base64Url `json:"attestationObject,omitempty"`
	PublicKey          db.Base64Url `json:"publicKey,omitempty"`
	PublicKeyAlgorithm int          `json:"publicKeyAlgorithm,omitempty"`
	Transports         []string     `json:"transports,omitempty"`

	// Assertion
	// Ref: https://w3c.github.io/webauthn/#iface-authenticatorassertionresponse
	Signature  db.Base64Url `json:"signature,omitempty"`
	UserHandle db.Base64Url `json:"userHandle,omitempty"`
}

// Ref: https://w3c.github.io/webauthn/#sctn-attestation
type attestationObject struct {
	Fmt      string         `json:"fmt"`
	AttStmt  map[string]any `json:"attStmt,omitempty"`
	AuthData []byte         `json:"authData,omitempty"`
}

// Ref: https://w3c.github.io/webauthn/#sctn-attested-credential-data
type attestedCredentialData struct {
	Aaguid              [16]byte             `json:"aaguid"`
	CredentialIdLength  uint16               `json:"credentialIdLength"`
	CredentialId        []byte               `json:"credentialId"`
	CredentialPublicKey *credentialPublicKey `json:"credentialPublicKey,omitempty"`

	credentialPublicKeyRaw []byte
}

// Ref: https://w3c.github.io/webauthn/#sctn-authenticator-data
type authenticatorData struct {
	RpIdHash               [32]byte                `json:"rpIdHash"`
	Flags                  byte                    `json:"flags"`
	SignCount              uint32                  `json:"signCount"`
	AttestedCredentialData *attestedCredentialData `json:"attestedCredentialData,omitempty"`
	Extensions             []byte                  `json:"extensions,omitempty"`
}

// Ref: https://w3c.github.io/webauthn/#client-data
type clientData struct {
	Challenge   string `json:"challenge"`
	CrossOrigin bool   `json:"crossOrigin"`
	Origin      string `json:"origin"`
	TopOrigin   string `json:"topOrigin,omitempty"`
	Type        string `json:"type"`
}

// credentialPublicKey represents a COSE_Key (RFC 9053) as used in WebAuthn.
// Supports EC2 (kty=2), OKP (kty=1), and RSA (kty=3) key types.
//
// EC2/OKP and RSA share CBOR labels -1 and -2 with different semantics, so
// this type uses a custom CBOR decoder (UnmarshalCBOR) instead of struct tags.
type credentialPublicKey struct {
	Kty coseKty // Key type
	Alg coseAlg // Algorithm
	Crv coseCrv // Curve (EC2/OKP)
	X   []byte  // X coordinate (EC2) or public key bytes (OKP)
	Y   []byte  // Y coordinate (EC2 only)
	N   []byte  // Modulus (RSA)
	E   []byte  // Exponent (RSA)
}

// MarshalCBOR implements cbor.Marshaler for credentialPublicKey.
// It encodes the key as a COSE_Key map (RFC 9052 § Section 7).
func (k *credentialPublicKey) MarshalCBOR() ([]byte, error) {
	m := make(map[int]any, 6)
	m[1] = k.Kty // kty
	m[3] = k.Alg // alg

	switch k.Kty {
	case coseKtyOKP, coseKtyEC2:
		m[-1] = k.Crv // crv
		m[-2] = k.X   // x
		if k.Kty == coseKtyEC2 {
			m[-3] = k.Y // y
		}
	case coseKtyRSA:
		m[-1] = k.N // n
		m[-2] = k.E // e
	default:
		return nil, fmt.Errorf("unsupported cose key type: %d", k.Kty)
	}

	return cborEncoder().Marshal(m)
}

// UnmarshalCBOR implements cbor.Unmarshaler for credentialPublicKey.
// It decodes COSE_Key maps (RFC 9052 § Section 7), disambiguating key-type-specific
// parameters (-1, -2, -3) based on the kty field.
//
// Ref: https://w3c.github.io/webauthn/#sctn-encoded-credPubKey-examples
func (k *credentialPublicKey) UnmarshalCBOR(data []byte) error {
	dec := cborDecoder()

	var raw map[int]cbor.RawMessage
	if err := dec.Unmarshal(data, &raw); err != nil {
		return err
	}

	if v, ok := raw[1]; ok { // kty
		if err := dec.Unmarshal(v, &k.Kty); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("missing cose key type")
	}

	if v, ok := raw[3]; ok { // alg
		if err := dec.Unmarshal(v, &k.Alg); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("missing cose algorithm")
	}

	switch k.Kty {
	case coseKtyOKP, coseKtyEC2:
		if v, ok := raw[-1]; ok { // crv
			if err := dec.Unmarshal(v, &k.Crv); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("missing curve identifier")
		}
		if v, ok := raw[-2]; ok { // x
			if err := dec.Unmarshal(v, &k.X); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("missing x-coordinate or public key")
		}

		if k.Kty == coseKtyEC2 {
			if v, ok := raw[-3]; ok { // y
				if err := dec.Unmarshal(v, &k.Y); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("missing y-coordinate")
			}
		}

	case coseKtyRSA:
		if v, ok := raw[-1]; ok { // n
			if err := dec.Unmarshal(v, &k.N); err != nil {
				return err
			}
		}
		if v, ok := raw[-2]; ok { // e
			if err := dec.Unmarshal(v, &k.E); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unsupported cose key type: %d", k.Kty)
	}

	return nil
}

// Ref: https://w3c.github.io/webauthn/#iface-pkcredential
type publicKeyCredential struct {
	AuthenticatorAttachment string                `json:"authenticatorAttachment"`
	Id                      string                `json:"id"`
	RawId                   db.Base64Url          `json:"rawId"`
	Response                authenticatorResponse `json:"response"`
	Type                    string                `json:"type"`
}

func credentialsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(r.Context(), w, ErrMethodNotAllowed)
		return
	}

	stateId := r.FormValue(stateKey)
	log := zerolog.Ctx(r.Context()).With().Str(stateKey, stateId).Logger()
	if stateId == "" {
		log.Debug().Msg("empty auth state")
		writeError(r.Context(), w, ErrInvalidRequest)
		return
	}

	credential := r.FormValue(credentialKey)
	if credential == "" {
		log.Debug().Msg("empty passkey credential")
		writeError(r.Context(), w, ErrInvalidRequest)
		return
	}

	state, err := db.RetrieveAuthState(r.Context(), stateId)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(r.Context(), w, ErrAccessDenied)
			return
		}

		log.Error().Err(err).Msg("failed to retrieve auth state")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if state.Expired() {
		log.Debug().Msg("expired state")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	// We only support passkey credentials (WebAuthn § 7.1)
	// Ref: https://w3c.github.io/webauthn/#sctn-registering-a-new-credential (Level 3)

	var cred publicKeyCredential
	if err = json.Unmarshal([]byte(credential), &cred); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal public key credential")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if cred.Type != credentialTypePublicKey {
		log.Warn().Str("type", cred.Type).Msg("invalid credential type")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	// Steps 6-11: client data
	var clientData clientData
	if err = json.Unmarshal(cred.Response.ClientDataJSON, &clientData); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal public key credential client data")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if clientData.Type != clientDataTypeWebauthnCreate {
		log.Warn().Str("type", clientData.Type).Msg("invalid client data type")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if clientData.Challenge != state.WebAuthnChallenge() {
		log.Warn().Str("challenge", clientData.Challenge).Msg("client data challenge mismatch")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if !strings.HasPrefix(common.AuthBaseUrl(), clientData.Origin) {
		log.Warn().Str("origin", clientData.Origin).Msg("unexpected client data origin")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if clientData.CrossOrigin {
		log.Warn().Str("topOrigin", clientData.TopOrigin).Msg("client data from cross origin")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	// Step 12: clientDataJSON hash (ignore)

	// Step 13: CBOR decoding on attestationObject field
	var attObj attestationObject
	if err = cborDecoder().Unmarshal(cred.Response.AttestationObject, &attObj); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal public key credential attestation object")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	authData := decodeAuthenticatorData(attObj.AuthData)
	if authData == nil {
		log.Warn().Msg("failed to decode authenticator data")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	// Step 14: verify rpIdHash
	rpIdHash := sha256.Sum256([]byte(state.RpId()))
	if rpIdHash != authData.RpIdHash {
		log.Warn().Msg("rp id hash mismatch")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	// Step 15: require UP flag
	if (authData.Flags & authDataFlagUV) != authDataFlagUV {
		log.Warn().Msg("user not verified")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	// Steps 16 - 19: auth data flags (ignore)

	// Step 20: verify the credential public key algorithm is supported
	if authData.AttestedCredentialData == nil {
		log.Warn().Msg("missing attested credential data")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	cpk := authData.AttestedCredentialData.CredentialPublicKey
	if cpk == nil {
		log.Warn().Msg("missing credential public key")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	switch cpk.Alg {
	case coseAlgES256, coseAlgEdDSA, coseAlgRS256:
		// Supported
	default:
		log.Warn().Int("alg", int(cpk.Alg)).Msg("unsupported credential algorithm")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	// Steps 21-24: attestation (ignore)
	// Step 25: verify credentialId length (done while decoding auth data)

	// Step 26: verify that the credentialId is not yet registered for any user
	found, err := db.FindIdnSrcBySub(r.Context(), cred.Id)
	if err != nil {
		log.Error().Err(err).Msg("failed to lookup credential id")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}
	if found {
		log.Warn().Str("credential_id", cred.Id).Msg("credential already exists")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	// Steps 27-29: credential record
	idn, err := db.NewIdn(r.Context(), state.AppId(), state.Login(), db.IdnAttrs{
		Email: state.Login(),
	})
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			log.Warn().Err(err).Msg("failed to create identity")
			writeError(r.Context(), w, ErrAccessDenied)
			return
		}

		log.Error().Err(err).Msg("failed to create identity")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	_, err = db.NewIdnSrcWithCr(r.Context(), idn.Id(), state.ProviderId(), cred.Id, db.IdnSrcAttrsCr{
		AttestationObject:         cred.Response.AttestationObject,
		AttestationClientDataJSON: cred.Response.ClientDataJSON,
		RpId:                      state.RpId(),

		BackupEligible: (authData.Flags & authDataFlagBE) == authDataFlagBE,
		BackupState:    (authData.Flags & authDataFlagBS) == authDataFlagBS,
		PublicKey:      authData.AttestedCredentialData.credentialPublicKeyRaw,
		SignCount:      authData.SignCount,
		Transports:     cred.Response.Transports,
		UvInitialized:  (authData.Flags & authDataFlagUV) == authDataFlagUV,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create passkey credential record")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	// Redirect with code so RP can choose to sign-in the user
	if err = state.Expire(r.Context()); err != nil {
		log.Error().Err(err).Msg("failed to expire auth state")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	code, err := db.NewAuthCode(r.Context(), 5*time.Minute, db.AuthCodeAttrs{
		ClientId:    state.ClientId(),
		ProviderId:  state.ProviderId(),
		RedirectUri: state.RedirectUri(),
		Sub:         idn.Id(),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create auth code")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	u, err := url.Parse(state.RedirectUri())
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse redirect uri")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	q := u.Query()
	q.Set(codeKey, code.Id())
	u.RawQuery = q.Encode()

	w.Header().Add("Location", u.String())
	w.WriteHeader(http.StatusFound)
}

func cborDecoder() cbor.DecMode {
	dec, err := cbor.DecOptions{
		DupMapKey:       cbor.DupMapKeyEnforcedAPF,
		MaxNestedLevels: 4,
		IndefLength:     cbor.IndefLengthForbidden,
		TagsMd:          cbor.TagsForbidden,
	}.DecMode()

	if err != nil {
		panic(err)
	}

	return dec
}

func cborEncoder() cbor.EncMode {
	enc, _ := cbor.CTAP2EncOptions().EncMode()
	return enc
}

func decodeAuthenticatorData(data []byte) *authenticatorData {
	if len(data) < minAuthenticatorDataLength {
		// Not enough bytes
		return nil
	}

	authData := authenticatorData{
		RpIdHash:  [32]byte(data[:32]),
		Flags:     data[32],
		SignCount: binary.BigEndian.Uint32(data[33:37]),
	}

	l := minAuthenticatorDataLength
	if (authData.Flags & authDataFlagAT) == authDataFlagAT {
		l += 16 + 2

		if len(data) < l {
			// Not enough bytes to decode attested credential data
			return nil
		}

		atd := attestedCredentialData{
			Aaguid:             [16]byte(data[37:53]),
			CredentialIdLength: binary.BigEndian.Uint16(data[53:55]),
		}

		credIdLen := atd.CredentialIdLength
		if credIdLen > maxCredentialIdLength {
			// Credential id length too long
			return nil
		}

		l += int(credIdLen)
		if len(data) < l {
			// Not enough bytes to decode credential id
			return nil
		}

		atd.CredentialId = data[55:l]

		r := bytes.NewReader(data[l:])
		if err := cborDecoder().NewDecoder(r).Decode(&atd.CredentialPublicKey); err != nil {
			return nil
		}

		atd.credentialPublicKeyRaw = data[l : len(data)-r.Len()]
		l += len(data[l:]) - r.Len()

		authData.AttestedCredentialData = &atd
	}

	if (authData.Flags & authDataFlagED) == authDataFlagED {
		if len(data) <= l {
			// Missing extension data
			return nil
		}

		authData.Extensions = data[l:]
	}

	if len(data) != l {
		// Letfover bytes afte decoding
		return nil
	}

	return &authData
}

// verifyAssertionSignature verifies an assertion signature.
// See WebAuthn § 6.3.3, step 11 for more info on how the signature is constructed.
// Ref: https://w3c.github.io/webauthn/#fig-signature (Level 3)
func verifyAssertionSignature(cpk *credentialPublicKey, sig []byte, authData []byte, hash []byte) error {
	data := slices.Concat(authData, hash)

	switch cpk.Alg {
	case coseAlgES256:
		pub := &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(cpk.X),
			Y:     new(big.Int).SetBytes(cpk.Y),
		}
		h := sha256.Sum256(data)
		var s struct{ R, S *big.Int }
		if _, err := asn1.Unmarshal(sig, &s); err != nil {
			return fmt.Errorf("invalid ecdsa signature: %w", err)
		}
		if !ecdsa.Verify(pub, h[:], s.R, s.S) {
			return fmt.Errorf("ecdsa signature mismatch")
		}

	case coseAlgEdDSA:
		if len(cpk.X) != ed25519.PublicKeySize {
			return fmt.Errorf("invalid ed25519 public key length: %d", len(cpk.X))
		}
		if !ed25519.Verify(ed25519.PublicKey(cpk.X), data, sig) {
			return fmt.Errorf("eddsa signature mismatch")
		}

	case coseAlgRS256:
		pub := &rsa.PublicKey{
			N: new(big.Int).SetBytes(cpk.N),
			E: int(new(big.Int).SetBytes(cpk.E).Int64()),
		}
		h := sha256.Sum256(data)
		if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], sig); err != nil {
			return fmt.Errorf("rsa signature mismatch: %w", err)
		}

	default:
		return fmt.Errorf("unsupported key algorithm: %d", cpk.Alg)
	}

	return nil
}
