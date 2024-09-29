package utils

import "os"

var (
    APIBaseURL = getEnv("API_BASE_URL", "http://localhost:8080/api/v1")
)

func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}
