package config

import "os"

// Environment represents the application environment
type Environment string

const (
	Development Environment = "development"
	Test        Environment = "test"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

// GetEnvironment returns the current environment
func GetEnvironment() Environment {
	env := os.Getenv("ENVIRONMENT")
	switch env {
	case "production":
		return Production
	case "staging":
		return Staging
	case "test":
		return Test
	default:
		return Development
	}
}

// IsDevelopment checks if the current environment is development
func (e Environment) IsDevelopment() bool {
	return e == Development
}

// IsProduction checks if the current environment is production
func (e Environment) IsProduction() bool {
	return e == Production
}

// IsStaging checks if the current environment is staging
func (e Environment) IsStaging() bool {
	return e == Staging
}

// IsTest checks if the current environment is test
func (e Environment) IsTest() bool {
	return e == Test
}
