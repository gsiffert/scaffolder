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
	ch     chan Status
}

func DefaultClient(ctx context.Context) DefaultHealthCheck {
	client := &defaultClient{
		ctx:    ctx,
		status: NotReady,
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
	case d.ch <- status:
	}
}

func (d *defaultClient) run(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d.status = <-d.ch:
			case d.ch <- d.status:
			}
		}
	}()
}
