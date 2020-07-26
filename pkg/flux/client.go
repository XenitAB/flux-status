package flux

import (
	"context"

	"github.com/fluxcd/flux/pkg/api/v6"
)

type Client interface {
	ListServices(context.Context, string) ([]v6.ControllerStatus, error)
}
