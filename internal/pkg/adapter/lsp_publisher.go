package adapter

import (
	"encoding/json"
	"fmt"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	portout "github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
)

type DiagnosticsPublisherImpl struct{}

func (DiagnosticsPublisherImpl) Publish(uri string, diagnostics []domain.Diagnostic) {
	publish := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "textDocument/publishDiagnostics",
		"params": map[string]interface{}{
			"uri":         uri,
			"diagnostics": diagnostics,
		},
	}
	publishBytes, _ := json.Marshal(publish)
	fmt.Printf("Content-Length: %d\r\n\r\n%s", len(publishBytes), publishBytes)
}

var _ portout.DiagnosticsPublisher = (*DiagnosticsPublisherImpl)(nil)
