package apispec

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/gofrs/uuid"
)

// This code is inspired by the openclarity speculator library
// https://github.com/openclarity/speculator

var digitCheck = regexp.MustCompile(`^[0-9]+$`)

func UnifyParameterizedPathIfApplicable(path string) string {
	var parameterizedPathParts []string
	paramCount := 0
	pathParts := strings.Split(path, "/")

	for _, part := range pathParts {
		if isSuspectPathParam(part) {
			paramCount++
			paramName := fmt.Sprintf("param%v", paramCount)
			parameterizedPathParts = append(parameterizedPathParts, "{"+paramName+"}")
		} else {
			parameterizedPathParts = append(parameterizedPathParts, part)
		}
	}
	return strings.Join(parameterizedPathParts, "/")
}

func isSuspectPathParam(pathPart string) bool {
	if isNumber(pathPart) {
		return true
	}
	if isUUID(pathPart) {
		return true
	}
	if isMixed(pathPart) {
		return true
	}
	if isString(pathPart) {
		return true
	}
	return false
}

func isString(pathPart string) bool {
	return strings.Contains(pathPart, "{") &&
		strings.Contains(pathPart, "}")
}

func isNumber(pathPart string) bool {
	return digitCheck.MatchString(pathPart)
}

func isUUID(pathPart string) bool {
	_, err := uuid.FromString(pathPart)
	return err == nil
}

// Check if a path part that is mixed from digits and chars can be considered as
// parameter following hard-coded heuristics. Temporary, we'll consider strings
// as parameters that are at least 8 chars longs and has at least 3 digits.
func isMixed(pathPart string) bool {
	const maxLen = 8
	const minDigitsLen = 2

	if len(pathPart) < maxLen {
		return false
	}

	return countDigitsInString(pathPart) > minDigitsLen
}

func countDigitsInString(s string) int {
	count := 0
	for _, c := range s {
		if unicode.IsNumber(c) {
			count++
		}
	}
	return count
}
