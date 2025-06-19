package domain

type Violation struct {
	File      string
	Line      uint32
	Character uint32
	Message   string
	Type      string // e.g. "Dependency violation"
}
