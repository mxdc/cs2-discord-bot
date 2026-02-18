package discord

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

// CountryCodeToEmojiFlag converts ISO country code (FR, DE, ES) to its Unicode flag emoji.
// If code is empty, returns two spaces to preserve monospace alignment.
func CountryCodeToEmojiFlag(countryCode string) string {
	// const fallback = rune(0x1F3F3) // white flag
	const fallback = rune(0x1F310) // Internet globe emoji

	if len(countryCode) != 2 {
		return string(fallback)
	}

	code := strings.ToUpper(countryCode)
	base := rune(0x1F1E6)
	first := base + (rune(code[0]) - 'A')
	second := base + (rune(code[1]) - 'A')

	return string(first) + string(second)
}
