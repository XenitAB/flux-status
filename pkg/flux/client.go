package flux

import (
	"context"

	"github.com/fluxcd/flux/pkg/api/v6"
)

// Client is an interface that wraps the basic Flux client methods.
type Client interface {
	ListServices(context.Context, string) ([]v6.ControllerStatus, error)
}
