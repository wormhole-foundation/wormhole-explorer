package utils

import (
	"regexp"
	"strings"
)

// FindSubstringBeforeDomains finds the substring before specified domains appear
// and also removes http:// or https:// prefix from the string.
func FindSubstringBeforeDomains(s string, domains []string) string {
	// Regular expression to match the protocol prefix (http:// or https://)
	reProtocol := regexp.MustCompile(`^(http://|https://)`)
	// Remove the protocol prefix
	s = reProtocol.ReplaceAllString(s, "")

	// Create a regular expression pattern for the domains
	// Escape dots since they have a special meaning in regex, and join domains with |
	domainPattern := strings.Join(domains, "|")
	domainPattern = strings.Replace(domainPattern, ".", `\.`, -1)
	reDomain := regexp.MustCompile("(" + domainPattern + ")")

	// Find the index of the first occurrence of any domain
	loc := reDomain.FindStringIndex(s)
	if loc != nil {
		// Return the substring from the start to the first domain occurrence
		return s[:loc[0]]
	}

	// If no domain is found, return the original modified string
	return s
}
