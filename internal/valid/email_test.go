package valid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"john@example.test", true},
		{"john@sub.example.test", true},
		{"john+spam@example.test", true},
		{"john.smith@example.test", true},
		{"john.smith+spam@example.test", true},
		{"_john@example.test", true},
		{"1john@example.test", true},
		{"j1@example.test", true},
		{"js@example.test", true},
		{"_@example.test", true},
		{"t+1@example.test", true},
		{"john-smith@example.test", true},
		{"j-s+1@example.test", true},
		{"j-1@example.test", true},
		{"j.smith-1@example.test", true},
		{"j.s-mail@example.test", true},
		{"_j.smith-1@example.test", true},
		{"long.email-address-with-hyphens@and.subdomains.example.com", true},
		{"john.smith+spam@example", false},           // invalid domain
		{"john+@example.test", false},                // local part, no end with '+'
		{"+john@example.test", false},                // local part, no start with '+'
		{"john.@example.test", false},                // local part, no end with '.'
		{".john@example.test", false},                // local part, no start with '.'
		{"john..smith@example.test", false},          // local part, no two consecutive '.'s
		{"john.+spam@example.test", false},           // local part, no '+' after '.'
		{"john+.spam@example.test", false},           // local part, no '.' after '+'
		{"user.name+tag+sorting@example.com", false}, // local part, no multiple '+'
		{"-@example.test", false},                    // local part, no '-' without wrapping alnum
		{"_-@example.test", false},                   // local part, no '-' without wrapping alnum
		{"_-_@example.test", false},                  // local part, no '-' without wrapping alnum
		{"j_-smith@example.test", false},             // local part, no '-' without wrapping alnum
		{"j-_smith@example.test", false},             // local part, no '-' without wrapping alnum
		{"j-@example.text", false},                   // local part, no '-' without wrapping alnum
		{"j-_+spam@example.text", false},             // local part, no '-' without wrapping alnum
		{"john@", false},                             // no domain
		{"@example.com", false},                      // no local part
		{"word", false},                              // no '@'
	}

	for _, test := range tests {
		t.Run(test.email, func(t *testing.T) {
			assert.Equal(t, test.valid, Email(test.email))
		})
	}
}
