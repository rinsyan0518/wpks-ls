package domain

type Violation struct {
	File    string
	Line    int
	Column  int
	Message string
}
