package registry

import (
	"fmt"
	"sync"

	"github.com/hackborn/onefunc/lock"
)

func Register(f Factory) error {
	return reg.Register(f)
}

func Find(name string) (Factory, bool) {
	return reg.Find(name)
}

func DriverNames() []string {
	return reg.DriverNames()
}

type registry struct {
	l         sync.Locker
	factories map[string]Factory
}

func newRegistry() *registry {
	lock := &sync.Mutex{}
	factories := make(map[string]Factory)
	return &registry{l: lock, factories: factories}
}

func (r *registry) Register(f Factory) error {
	defer lock.Locker(reg.l).Unlock()
	if _, ok := r.factories[f.Name]; ok {
		return fmt.Errorf(`Factory name "` + f.Name + `" alreeady in use`)
	}
	r.factories[f.Name] = f
	return nil
}

func (r *registry) Find(name string) (Factory, bool) {
	defer lock.Locker(reg.l).Unlock()
	if f, ok := r.factories[name]; ok {
		return f, ok
	}
	return Factory{}, false
}

func (r *registry) DriverNames() []string {
	defer lock.Locker(reg.l).Unlock()
	n := make([]string, 0, len(r.factories))
	for k, _ := range r.factories {
		n = append(n, k)
	}
	return n
}

var (
	reg *registry = newRegistry()
)
