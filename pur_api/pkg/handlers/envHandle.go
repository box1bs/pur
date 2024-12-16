package handlers

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func ExtractVars(keys []string) ([]string, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading env vars: %v", err)
	}

	var vars []string
	for _, v := range keys {
		vars = append(vars, os.Getenv(v))
	}

	return vars, nil
}