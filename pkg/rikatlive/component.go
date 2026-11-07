package rikatlive

import (
	"fmt"
	"sync"
)

// Component defines the essential contract for a RikatLive reactive server component.
type Component interface {
	Render() string
}

// Mounter allows a component to receive initialization properties upon creation.
type Mounter interface {
	Mount(props map[string]any) error
}

// Updater allows a component to respond whenever a field is updated from the client.
type Updater interface {
	Updated(field string, value any) error
}

// DynamicComponent allows non-struct components (such as interpreted .ngpk scripts)
// to manage their own reactive state, action calls, and rendering.
type DynamicComponent interface {
	Component
	GetState() map[string]any
	SetState(state map[string]any) error
	CallAction(action string, args []any) error
}

// ComponentLoader resolves and instantiates a Component dynamically by name.
type ComponentLoader func(name string) (Component, error)

var (
	registryMu    sync.RWMutex
	registry      = make(map[string]func() Component)
	customLoaders []ComponentLoader
)

// Register registers a component constructor function under a unique name.
func Register(name string, factory func() Component) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[name] = factory
}

// RegisterLoader registers a dynamic component loader (e.g. for .ngpk components).
func RegisterLoader(loader ComponentLoader) {
	registryMu.Lock()
	defer registryMu.Unlock()
	customLoaders = append(customLoaders, loader)
}

// Create instantiates a fresh Component instance by registered name or via dynamic loaders.
func Create(name string) (Component, error) {
	registryMu.RLock()
	factory, exists := registry[name]
	loaders := make([]ComponentLoader, len(customLoaders))
	copy(loaders, customLoaders)
	registryMu.RUnlock()

	if exists && factory != nil {
		return factory(), nil
	}

	for _, loader := range loaders {
		comp, err := loader(name)
		if err == nil && comp != nil {
			return comp, nil
		}
	}

	return nil, fmt.Errorf("rikatlive: component '%s' is not registered or found", name)
}

