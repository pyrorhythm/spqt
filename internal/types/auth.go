package types

import "context"

type Authenticator interface {
	Authorize(ctx context.Context) <-chan Event
}
