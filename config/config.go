package config

import "os"

type Config struct {
	ServerPort	string
	DbName		string
	DbPassword	string
	DbUsername	string
}

//constructor
func NewConfig() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", ""),
		DbName: getEnv("DB_NAME", ""),
		DbPassword: getEnv("DB_PASSWORD", ""),
		DbUsername: getEnv("DB_USERNAME", ""),
	}
}

//helper function to read an environment variable or return a default value
func getEnv(key, defaultVal string) string {
	if value, exist := os.LookupEnv(key); exist {
		return value
	}

	return defaultVal
}
