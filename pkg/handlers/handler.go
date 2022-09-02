package handlers

import "context"
import "sigs.k8s.io/controller-runtime/pkg/client"

type Interface interface {
	Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
}

func GetHandler(h Handler) Interface {
	return handler{h: h}
}

type Handler interface {
	Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
}

type handler struct {
	h Handler
}

func (h handler) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return h.h.Update(ctx, obj, opts...)
}
