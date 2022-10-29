package util

import "os"

func GetEnvOrDefault(envName string, defaultValue string) string {
	env := os.Getenv(envName)
	if len(env) == 0 {
		return defaultValue
	}
	return env
}
