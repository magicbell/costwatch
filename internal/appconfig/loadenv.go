package appconfig

import (
	"fmt"
	"os"
	"path"

	"github.com/joho/godotenv"
)

// LoadEnvFile loads the environment variables from the file. It uses
// godotenv.Load to load the environment variables from the file,
// which WILL NOT OVERRIDE an env variable that already exists.
func LoadEnvFile(name string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %s", err)
	}

	if err := godotenv.Load(path.Join(cwd, name)); err != nil {
		return fmt.Errorf("godotenv.Load: %w", err)
	}

	return nil
}
