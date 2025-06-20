package domain

import (
	"regexp"
	"strconv"
	"strings"
)

var lineRegex = regexp.MustCompile(`^([^:]+):(\d+):(\d+)$`)
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
var messageRegex = regexp.MustCompile(`^([^:]+): `)

type CheckResult struct {
	body string
}

func NewCheckResult(body string) *CheckResult {
	return &CheckResult{body: body}
}

// ParsePackwerkOutput parses the output of 'packwerk check' and returns violations.
func (c *CheckResult) Parse() []Violation {
	var violations []Violation
	lines := c.cleanOutputLines()
	for i := 0; i < len(lines); i++ {
		m := lineRegex.FindStringSubmatch(lines[i])
		if m != nil && i+1 < len(lines) {
			line, _ := strconv.Atoi(m[2])
			column, _ := strconv.Atoi(m[3])
			// Collect all lines after the match until the next blank line
			msgLines := []string{}
			for j := i + 1; j < len(lines); j++ {
				if lines[j] == "" {
					break
				}
				msgLines = append(msgLines, lines[j])
			}
			msg := strings.Join(msgLines, " ")
			violationType := ""
			if len(msgLines) > 0 {
				if mm := messageRegex.FindStringSubmatch(msgLines[0]); mm != nil {
					violationType = mm[1]
				}
			}
			violations = append(violations, Violation{
				File:      m[1],
				Line:      uint32(line),
				Character: uint32(column),
				Message:   msg,
				Type:      violationType,
			})
			i += len(msgLines) // skip message lines
		}
	}
	return violations
}

func (c *CheckResult) cleanOutputLines() []string {
	var result []string
	for _, line := range strings.Split(c.body, "\n") {
		line = ansiEscape.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		result = append(result, line)
	}
	return result
}
