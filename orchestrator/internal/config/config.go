// Package config reads all configuration from the environment (Twelve-Factor,
// Factor III). A missing required variable is a fatal configuration error,
// never a silently-applied default for infrastructure values.
package config

import (
	"fmt"
	"os"
)

// MissingConfigError is returned by RequireEnv when a required variable is unset or empty.
type MissingConfigError struct {
	VarName string
}

func (e *MissingConfigError) Error() string {
	return fmt.Sprintf("required environment variable %q is not set", e.VarName)
}

// RequireEnv returns the value of the named environment variable, or a
// MissingConfigError if it is unset or empty.
func RequireEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", &MissingConfigError{VarName: name}
	}
	return value, nil
}
