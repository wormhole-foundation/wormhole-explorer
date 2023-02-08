package healthcheck

import "context"

type Check func(context.Context) error
