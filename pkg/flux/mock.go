package flux

import (
	"context"

	"github.com/fluxcd/flux/pkg/api/v6"
)

type Mock struct {
	Services []v6.ControllerStatus
}

func (m *Mock) ListServices(ctx context.Context, namespace string) ([]v6.ControllerStatus, error) {
	return m.Services, nil
}
