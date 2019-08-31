package healthcheck

import (
	"context"
)

type DefaultHealthCheck interface {
	HealthCheck
	SetStatus(status Status)
}

type defaultClient struct {
	ctx    context.Context
	status Status
	inner  chan Status
	ch     chan Status
}

func DefaultClient(ctx context.Context) DefaultHealthCheck {
	client := &defaultClient{
		ctx:    ctx,
		status: NotReady,
		inner:  make(chan Status),
		ch:     make(chan Status),
	}
	go client.run(ctx)
	return client
}

func (d *defaultClient) Status() <-chan Status {
	return d.ch
}

func (d *defaultClient) SetStatus(status Status) {
	select {
	case <-d.ctx.Done():
		return
	case d.inner <- status:
	}
}

func (d *defaultClient) run(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d.status = <-d.inner:
			case d.ch <- d.status:
			}
		}
	}()
}
