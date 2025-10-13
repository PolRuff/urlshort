package service

import "net/url"

// IsValidURL checks if a string is a valid HTTP or HTTPS URL
func IsValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}
