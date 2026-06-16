package db

import (
	"github.com/google/uuid"
)

func uuidv7() string {
	id, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	b, err := id.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return encoding.EncodeToString(b)
}
