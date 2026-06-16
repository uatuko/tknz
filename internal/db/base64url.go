package db

import "encoding/base64"

type Base64Url []byte

func (b Base64Url) MarshalJSON() ([]byte, error) {
	if b == nil {
		return []byte("null"), nil
	}
	enc := base64.RawURLEncoding.EncodeToString(b)
	return []byte(`"` + enc + `"`), nil
}

func (b *Base64Url) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*b = nil
		return nil
	}
	if len(data) < 3 || data[0] != '"' || data[len(data)-1] != '"' {
		return nil
	}
	dec, err := base64.RawURLEncoding.DecodeString(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*b = dec
	return nil
}
