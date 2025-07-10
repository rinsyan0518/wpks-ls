package in

type CreateWorkspace interface {
	Create(rootUri string, rootPath string) error
}
