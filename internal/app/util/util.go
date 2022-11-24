// Package util define utils for app.
package util

import "os"

// GetEnvOrDefault returns env or defaultValue
func GetEnvOrDefault(envName string, defaultValue string) string {
	env := os.Getenv(envName)
	if len(env) == 0 {
		return defaultValue
	}
	return env
}
