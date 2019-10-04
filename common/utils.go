package common

import (
	"fmt"
	"github.com/satori/go.uuid"
	"strings"
)

func GetNormalizedDeploymentName(appName string, deploymentID int) string {
	return GetNormalizedApplicationName(appName) + fmt.Sprintf("--%v", deploymentID)
}

func GetNormalizedApplicationName(appName string) string {
	appName = strings.ReplaceAll(appName, ".", "-")
	for _, c := range appName {
		if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') {
			continue
		}
		if c == '-' {
			continue
		}
		return ""
	}
	return appName
}

func ValidApplicationName(appName string) bool {
	return GetNormalizedApplicationName(appName) != ""
}

func GenerateRandomToken() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}

func EscapeForRegexp(v string) string {
	return strings.Replace(v, ".", "\\.", -1)
}

func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
