package healthcheck

import (
	"sync"

	"github.com/Vorian-Atreides/scaffolder"
)

type HealthRegistry interface {
	SetStatus(service string, status Status)
	Services() map[string]Status
}

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

type healthRegistry struct {
	mutex    sync.Mutex
	services map[string]Status
}

func NewRegistry() HealthRegistry {
	registry := &healthRegistry{}
	scaffolder.Init(registry)
	return registry
}

func (h *healthRegistry) Default() {
	h.services = make(map[string]Status)
}

func (h *healthRegistry) SetStatus(service string, status Status) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.services[service] = status
}

func (h *healthRegistry) Services() map[string]Status {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	cp := map[string]Status{}
	for key, value := range h.services {
		cp[key] = value
	}
	return cp
}
