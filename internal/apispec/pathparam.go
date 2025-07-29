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

// UnifyParameterizedPathIfApplicable normalizes a path by replacing dynamic segments with {paramN}.
// If isSpec = true, also treats existing {param} segments in OpenAPI specs as parameters.
func UnifyParameterizedPathIfApplicable(path string, isSpec bool) string {
	if path == "" {
		return ""
	}
	if path == "/" {
		return "/"
	}

	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	var parameterizedPathParts []string
	paramCount := 0

	for _, part := range pathParts {
		if isSpec && strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			parameterizedPathParts = append(parameterizedPathParts, part)
			continue
		}

		if isSuspectPathParam(part) {
			paramCount++
			paramName := fmt.Sprintf("param%v", paramCount)
			parameterizedPathParts = append(parameterizedPathParts, "{"+paramName+"}")
		} else {
			parameterizedPathParts = append(parameterizedPathParts, part)
		}
	}
	return "/" + strings.Join(parameterizedPathParts, "/")
}

func isSuspectPathParam(part string) bool {
	return isNumber(part) || isUUID(part) || isMixed(part)
}

func isNumber(s string) bool {
	return digitCheck.MatchString(s)
}

func isUUID(s string) bool {
	_, err := uuid.FromString(s)
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
