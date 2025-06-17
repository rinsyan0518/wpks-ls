package domain

import "strings"

type Configuration struct {
	RootUri  string
	RootPath string
}

func NewConfiguration(rootUri string, rootPath string) *Configuration {
	return &Configuration{RootUri: rootUri, RootPath: rootPath}
}

func (c *Configuration) StripRootUri(uri string) string {
	return strings.TrimPrefix(uri, c.RootUri)
}
