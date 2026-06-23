package jws

import (
	"go.tknz.dev/internal/jose/jwa"
)

const (
	HeaderTypJWT HeaderTyp = "JWT"
)

type HeaderTyp string

type Header struct {
	Alg jwa.Alg   `json:"alg,omitempty"`
	Kid string    `json:"kid,omitempty"`
	Typ HeaderTyp `json:"typ,omitempty"`
}
