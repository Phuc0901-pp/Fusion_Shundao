package auth

import "fusion/internal/platform/config"

// Credentials returns login info from dynamic config (app.json).
// Falls back to empty strings if config not loaded yet.
func Credentials() (username, password, loginURL string) {
	creds := config.App.Credentials
	return creds.Username, creds.Password, creds.LoginURL
}
