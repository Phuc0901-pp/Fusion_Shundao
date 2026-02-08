package utils

import (
	"github.com/google/uuid"
)

// ProjectNamespace is the base namespace for FusionSolar project UUIDs
// Generated using uuid.New() once to ensure consistency
var ProjectNamespace = uuid.Must(uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")) // Using DNS namespace as base for now, or could define a custom one

// GenerateUUID generates a deterministic UUID v5 based on the ProjectNamespace and input string (DN)
func GenerateUUID(input string) string {
	if input == "" {
		return ""
	}
	return uuid.NewSHA1(ProjectNamespace, []byte(input)).String()
}
