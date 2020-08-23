package flux

import (
	"context"

	"github.com/fluxcd/flux/pkg/api/v6"
)

// Mock is a dummy implementation of the Client interface.
type Mock struct {
	Services []v6.ControllerStatus
}

// ListServices returns the contents of the Services variable.
func (m *Mock) ListServices(ctx context.Context, namespace string) ([]v6.ControllerStatus, error) {
	return m.Services, nil
}
