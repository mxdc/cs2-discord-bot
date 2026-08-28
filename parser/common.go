package parser

import (
	"fmt"
	"strings"
)

// CountryCodeToFlag converts ISO country codes to flag emojis
func CountryCodeToFlag(countryCode string) string {
	if countryCode == "" || len(countryCode) != 2 {
		return ":earth_americas:"
	}

	// Convert to lowercase for Discord
	return fmt.Sprintf(":flag_%s:", strings.ToLower(countryCode))
}
