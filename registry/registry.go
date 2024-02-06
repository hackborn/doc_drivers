package registry

import (
	"fmt"
	"sync"

	"github.com/hackborn/onefunc/lock"
)

// Register a new backend.
func Register(f Factory) error {
	return reg.Register(f)
}

// Open a backend. The backend will register any
// required dependencies (nodes, filesystems, etc.)
func Open(name string) (Factory, error) {
	return reg.Open(name)
}

// Find a backend.
func Find(name string) (Factory, bool) {
	return reg.Find(name)
}

func Names() []string {
	return reg.Names()
}

type registry struct {
	lock      sync.Mutex
	factories map[string]Factory
}

func newRegistry() *registry {
	factories := make(map[string]Factory)
	return &registry{factories: factories}
}

func (r *registry) Register(f Factory) error {
	defer lock.Locker(&r.lock).Unlock()
	if _, ok := r.factories[f.Name]; ok {
		return fmt.Errorf(`Factory "` + f.Name + `" alreeady registered`)
	}
	r.factories[f.Name] = f
	return nil
}

func (r *registry) Open(name string) (Factory, error) {
	defer lock.Locker(&r.lock).Unlock()
	if f, ok := r.factories[name]; ok {
		if f.Open != nil {
			err := f.Open()
			if err != nil {
				return Factory{}, err
			}
		}
		return f, nil
	}
	return Factory{}, fmt.Errorf("No backend available for name \"%v\"", name)
}

func (r *registry) Find(name string) (Factory, bool) {
	defer lock.Locker(&r.lock).Unlock()
	if f, ok := r.factories[name]; ok {
		return f, ok
	}
	return Factory{}, false
}

func (r *registry) Names() []string {
	defer lock.Locker(&r.lock).Unlock()
	n := make([]string, 0, len(r.factories))
	for k, _ := range r.factories {
		n = append(n, k)
	}
	return n
}

var (
	reg *registry = newRegistry()
)
