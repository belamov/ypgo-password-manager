package proto

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ClientAuthInterceptor is a client interceptor for authentication
type ClientAuthInterceptor struct {
	accessToken string
}

// NewClientAuthInterceptor  returns a new auth interceptor
func NewClientAuthInterceptor(accessToken string) *ClientAuthInterceptor {
	return &ClientAuthInterceptor{
		accessToken: accessToken,
	}
}

// Unary returns a client interceptor to authenticate unary RPC
func (interceptor *ClientAuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return invoker(interceptor.attachToken(ctx), method, req, reply, cc, opts...)
	}
}

// Stream returns a client interceptor to authenticate stream RPC
func (interceptor *ClientAuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		return streamer(interceptor.attachToken(ctx), desc, cc, method, opts...)
	}
}

func (interceptor *ClientAuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", interceptor.accessToken)
}
