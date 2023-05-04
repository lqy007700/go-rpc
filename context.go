package go_rpc

import "context"

type onewayKey struct {
}

func CtxtWithOneway(ctx context.Context) context.Context {
	return context.WithValue(ctx, onewayKey{}, true)
}

func isOneway(ctx context.Context) bool {
	value := ctx.Value(onewayKey{})
	v, ok := value.(bool)
	return ok && v
}
