package common

import (
	"fmt"
	"os"
)

func GetRequiredEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s environment variable is required", name)
	}

	return value, nil
}
