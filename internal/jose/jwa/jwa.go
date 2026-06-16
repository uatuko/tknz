package jwa

const (
	AlgNone  Alg = "none"
	AlgES256 Alg = "ES256"
	AlgRS256 Alg = "RS256"

	EcCrvP256 EcCrv = "P-256"
)

type Alg string
type EcCrv string
