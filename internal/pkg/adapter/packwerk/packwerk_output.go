package packwerk

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
var packwerkFileLineOutputRegex = regexp.MustCompile(`^([^:]+):(\d+):(\d+)$`)
var packwerkMessageRegex = regexp.MustCompile(`^([^:]+): `)

type PackwerkOutput struct {
	body string
}

func NewPackwerkOutput(body string) *PackwerkOutput {
	return &PackwerkOutput{body: body}
}

// ParsePackwerkOutput parses the output of 'packwerk check' and returns violations.
func (p *PackwerkOutput) Parse() []domain.Violation {
	var violations []domain.Violation
	lines := p.cleanOutputLines()
	for i := 0; i < len(lines); i++ {
		m := packwerkFileLineOutputRegex.FindStringSubmatch(lines[i])
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
				if mm := packwerkMessageRegex.FindStringSubmatch(msgLines[0]); mm != nil {
					violationType = mm[1]
				}
			}
			violations = append(violations, domain.Violation{
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

func (p *PackwerkOutput) cleanOutputLines() []string {
	var result []string
	for _, line := range strings.Split(p.body, "\n") {
		line = ansiEscape.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		result = append(result, line)
	}
	return result
}
