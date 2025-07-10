package domain

import "strings"

type Workspace struct {
	RootUri  string
	RootPath string
}

func NewWorkspace(rootUri string, rootPath string) *Workspace {
	return &Workspace{RootUri: rootUri, RootPath: rootPath}
}

func (w *Workspace) StripRootUri(uri string) string {
	return strings.TrimPrefix(uri, w.RootUri+"/")
}

func (w *Workspace) BuildFileUri(filePath string) string {
	return w.RootUri + "/" + filePath
}
