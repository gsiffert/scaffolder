package healthcheck

import (
	"context"
	"fmt"
	"net/http"
)

type MergingRule func(services map[string]Status) Status

func EveryService(mustBe Status, otherwise Status) MergingRule {
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
	HC          *HealthChecker
	MergingRule func(services map[string]Status) Status

	status Status
}

func (h *HTTPHandler) Default() {
	h.status = NotReady
	h.MergingRule = EveryService(Ready, NotReady)
}

func (h *HTTPHandler) Start(ctx context.Context) error {
	go func() {
		for {
			var services map[string]Status
			select {
			case <-ctx.Done():
				return
			case services = <-h.HC.Status():
			}
			h.status = h.MergingRule(services)
		}
	}()
	return nil
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.status != Ready && h.status != Healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
