// Package service events: helper for constructing EventContext from request context.
package service

import (
	"context"

	"avex-backend/internal/modules/orders/port"
)

func (s *Service) eventContext(ctx context.Context, actor port.ActorContext) port.EventContext {
	return port.EventContext{
		Actor: actor,
		Metadata: port.EventMetadata{
			OccurredAt: s.deps.Clock.Now(),
		},
	}
}
