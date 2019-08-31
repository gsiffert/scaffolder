package healthcheck

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Vorian-Atreides/scaffolder"
)

type MergingStrategy func(services map[string]Status) Status

func EveryService(mustBe Status, otherwise Status) MergingStrategy {
	return func(services map[string]Status) Status {
		for _, status := range services {
			if status != mustBe {
				return otherwise
			}
		}
		return mustBe
	}
}

type HTTPHandler struct {
	Interval        time.Duration
	Registry        HealthRegistry
	MergingStrategy MergingStrategy

	mutex sync.Mutex
	value Status
}

func (h *HTTPHandler) Default() {
	h.value = NotReady
	h.Interval = 3 * time.Second
	h.MergingStrategy = EveryService(Ready, NotReady)
}

func WithInterval(interval time.Duration) scaffolder.Option {
	return func(h *HTTPHandler) error {
		h.Interval = interval
		return nil
	}
}

func WithMergingStrategy(strategy MergingStrategy) scaffolder.Option {
	return func(h *HTTPHandler) error {
		h.MergingStrategy = strategy
		return nil
	}
}

func (h *HTTPHandler) Start(ctx context.Context) error {
	go func() {
		for {
			h.setStatus(h.MergingStrategy(h.Registry.Services()))

			select {
			case <-ctx.Done():
				return
			case <-time.After(h.Interval):
			}
		}
	}()
	return nil
}

func (h *HTTPHandler) setStatus(status Status) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.value = status
}

func (h *HTTPHandler) status() Status {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return h.value
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := h.status()

	if status != Ready && status != Healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
