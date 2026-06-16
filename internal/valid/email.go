package valid

import "strings"

// Email checks whether the given string is a valid email.
// This is a stricter check compared to [RFC 5322 addr-spec](https://www.rfc-editor.org/rfc/rfc5322#section-3.4.1)
func Email(s string) bool {
	local, domain, ok := strings.Cut(s, "@")
	if !ok {
		return false
	}

	if !rxEmailLocalPart().MatchString(local) {
		// Invalid local part
		return false
	}

	if !rxDomainName().MatchString(domain) {
		// Invalid domain
		return false
	}

	return true
}
