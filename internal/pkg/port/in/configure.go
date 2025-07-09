package in

type Configure interface {
	Configure(rootUri string, rootPath string) error
}
