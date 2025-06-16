package app

import (
	"regexp"
	"strings"
)

type Violation struct {
	File    string
	Line    int
	Column  int
	Message string
}

var lineRegex = regexp.MustCompile(`^([^:]+):(\d+):(\d+)$`)
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// ParsePackwerkOutput parses the output of 'packwerk check' and returns violations.
func ParsePackwerkOutput(output string) []Violation {
	var violations []Violation
	lines := cleanOutputLines(output)
	for i := 0; i < len(lines); i++ {
		m := lineRegex.FindStringSubmatch(lines[i])
		if m != nil && i+1 < len(lines) {
			violations = append(violations, Violation{
				File:    m[1],
				Line:    atoi(m[2]),
				Column:  atoi(m[3]),
				Message: lines[i+1],
			})
			i++ // skip message line
		}
	}
	return violations
}

func cleanOutputLines(output string) []string {
	var result []string
	for _, line := range strings.Split(output, "\n") {
		line = ansiEscape.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func atoi(s string) int {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}
