package healthcheck

import (
	"context"
	"time"

	"github.com/Vorian-Atreides/scaffolder"
)

type Status uint8

const (
	NotReady Status = iota
	NotHealthy
	Healthy
	Ready
)

func (s Status) String() string {
	switch s {
	case NotReady:
		return "Not ready"
	case NotHealthy:
		return "Not healthy"
	case Healthy:
		return "Healthy"
	case Ready:
		return "Ready"
	}
	return "Unknown"
}

type HealthResponse struct {
	ComponentName string
	Status        Status
}

type HealthCheck interface {
	Status() <-chan Status
}

type HealthChecker struct {
	reporter            chan map[string]Status
	unresponsiveTimeout time.Duration
	healthCheckInterval time.Duration
	reportInterval      time.Duration

	Containers   []scaffolder.Container `scaffolder:"containers"`
	workerCancel func()
	done         chan struct{}
}

func (h *HealthChecker) Default() {
	h.healthCheckInterval = 300 * time.Millisecond
	h.unresponsiveTimeout = 1 * time.Second
	h.reportInterval = 5 * time.Second
	h.reporter = make(chan map[string]Status)
	h.done = make(chan struct{})
}

func (h *HealthChecker) Start(ctx context.Context) error {
	go h.run(ctx)
	return nil
}

func (h *HealthChecker) Stop(ctx context.Context) error {
	if h.workerCancel != nil {
		h.workerCancel()
	}
	select {
	case <-ctx.Done():
	case <-h.done:
	}
	return nil
}

func (h *HealthChecker) monitor(ctx context.Context, hc HealthCheck, name string) <-chan HealthResponse {
	out := make(chan HealthResponse)

	go func() {
		defer close(out)
		response := HealthResponse{ComponentName: name, Status: NotReady}
		for {
			select {
			case <-ctx.Done():
				return
			case response.Status = <-hc.Status():
			case <-time.After(h.unresponsiveTimeout):
				response.Status = NotReady
			}

			select {
			case <-ctx.Done():
				return
			case out <- response:
			}
			time.Sleep(h.healthCheckInterval)
		}
	}()
	return out
}

func (h *HealthChecker) Status() <-chan map[string]Status {
	return h.reporter
}

func (h *HealthChecker) run(ctx context.Context) {
	defer close(h.done)
	defer close(h.reporter)

	// Start to monitor the components.
	var childCtx context.Context
	childCtx, h.workerCancel = context.WithCancel(ctx)
	defer h.workerCancel()
	var componentStatus []<-chan HealthResponse
	for _, container := range h.Containers {
		component := container.Component()
		if hc, ok := component.(HealthCheck); ok {
			status := h.monitor(childCtx, hc, container.Name())
			componentStatus = append(componentStatus, status)
		}
	}

	// Fan in the monitoring from every component.
	fromComponents := make(chan HealthResponse)
	defer close(fromComponents)
	for _, status := range componentStatus {
		go func(status <-chan HealthResponse) {
			for response := range status {
				select {
				case <-ctx.Done():
					return
				case fromComponents <- response:
				}
			}
		}(status)
	}

	// Batch the health status and report them to the consumer.
	status := map[string]Status{}
	ticker := time.NewTicker(h.reportInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case response := <-fromComponents:
			status[response.ComponentName] = response.Status
		case <-ticker.C:
			select {
			case <-ctx.Done():
				return
			case h.reporter <- status:
			}
		}
	}
}
