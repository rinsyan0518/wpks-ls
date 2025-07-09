package lsp

// ServerOptions represents the initialization options sent by the client
type ServerOptions struct {
	CheckAllOnInitialized bool
}

func NewServerOptions() *ServerOptions {
	return &ServerOptions{CheckAllOnInitialized: false}
}

func (o *ServerOptions) Apply(initializationOptions any) {
	if initializationOptions == nil {
		return
	}

	if optionsMap, ok := initializationOptions.(map[string]any); ok {
		if checkAll, ok := optionsMap["checkAllOnInitialized"].(bool); ok {
			o.CheckAllOnInitialized = checkAll
		}
	}
}
