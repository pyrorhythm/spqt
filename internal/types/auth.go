package types

import "context"

type Authenticator func(ctx context.Context) <-chan AuthEvent
