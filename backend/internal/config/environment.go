package config

import "os"

// Environment アプリケーション環境を表す
type Environment string

const (
	Development Environment = "development"
	Test        Environment = "test"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

// GetEnvironment 現在の環境を返す
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

// IsDevelopment 現在の環境が開発環境かチェック
func (e Environment) IsDevelopment() bool {
	return e == Development
}

// IsProduction 現在の環境が本番環境かチェック
func (e Environment) IsProduction() bool {
	return e == Production
}

// IsStaging 現在の環境がステージング環境かチェック
func (e Environment) IsStaging() bool {
	return e == Staging
}

// IsTest 現在の環境がテスト環境かチェック
func (e Environment) IsTest() bool {
	return e == Test
}
