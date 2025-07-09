package lsp

// ServerOptions represents the initialization options sent by the client
type ServerOptions struct {
	CheckAllOnInitialized bool
}

// ParseInitializationOptions safely parses any type to ServerOptions
func ParseInitializationOptions(initializationOptions any) ServerOptions {
	result := ServerOptions{
		CheckAllOnInitialized: false,
	}

	if initializationOptions == nil {
		return result
	}

	if optionsMap, ok := initializationOptions.(map[string]any); ok {
		if checkAll, ok := optionsMap["checkAllOnInitialized"].(bool); ok {
			result.CheckAllOnInitialized = checkAll
		}
	}

	return result
}
