package domain

type Violation struct {
	File    string
	Line    uint32
	Column  uint32
	Message string
}
