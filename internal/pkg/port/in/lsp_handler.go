package portin

type LSPHandler interface {
	HandleInitialize(params map[string]interface{}, id interface{})
	HandleDidOpen(params map[string]interface{})
	HandleDidChange(params map[string]interface{})
}
