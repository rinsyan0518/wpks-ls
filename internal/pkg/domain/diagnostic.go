package domain

type Position struct {
	Line      uint32
	Character uint32
}

type Range struct {
	Start Position
	End   Position
}

const (
	SeverityError   = 1
	SeverityWarning = 2
	SeverityInfo    = 3
	SeverityHint    = 4
)

type Diagnostic struct {
	Range    Range
	Severity int32
	Source   string
	Message  string
}
