package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"moviedata.com/pkg/discovery"
)

type serviceName string
type instanceID string

type Registry struct {
	sync.RWMutex
	serviceAddrs map[serviceName]map[instanceID]*serviceInstance
}

type serviceInstance struct {
	hostPort   string
	lastActive time.Time
}

func NewRegistry() *Registry {
	return &Registry{serviceAddrs: map[serviceName]map[instanceID]*serviceInstance{}}
}

func (r *Registry) Register(ctx context.Context, instID instanceID, servName serviceName, hostPort string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[servName]; !ok {
		r.serviceAddrs[servName] = map[instanceID]*serviceInstance{}
	}
	r.serviceAddrs[servName][instID] = &serviceInstance{hostPort: hostPort, lastActive: time.Now()}
	return nil
}

func (r *Registry) Deregister(ctx context.Context, instID instanceID, servName serviceName) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[servName]; !ok {
		return nil
	}
	delete(r.serviceAddrs[servName], instID)
	return nil
}

func (r *Registry) ReportHealthyState(instID instanceID, servName serviceName) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[servName]; !ok {
		return errors.New("service is not registered")
	}
	if _, ok := r.serviceAddrs[servName][instID]; !ok {
		return errors.New("service instance is not registered yet")
	}
	r.serviceAddrs[servName][instID].lastActive = time.Now()
	return nil
}

func (r *Registry) ServiceAddresses(ctx context.Context, servName serviceName) ([]string, error) {
	r.RLock()
	defer r.RUnlock()
	if len(r.serviceAddrs[servName]) == 0 {
		return nil, discovery.ErrNotFound
	}

	var res []string
	for _, i := range r.serviceAddrs[servName] {
		if i.lastActive.Before(time.Now().Add(-5 * time.Second)) {
			continue
		}
		res = append(res, i.hostPort)
	}
	return res, nil
}
