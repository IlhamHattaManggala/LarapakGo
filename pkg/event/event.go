package event

import (
	"fmt"
	"sync"
)

// ListenerFunc defines the callback signature for event listeners.
type ListenerFunc func(eventName string, payload any) error

// Dispatcher manages event listeners and dispatches events synchronously or asynchronously.
type Dispatcher struct {
	listeners map[string][]ListenerFunc
	mu        sync.RWMutex
}

var (
	globalDispatcher *Dispatcher
	dispatcherOnce   sync.Once
)

// GetDispatcher returns the global singleton event dispatcher.
func GetDispatcher() *Dispatcher {
	dispatcherOnce.Do(func() {
		globalDispatcher = New()
	})
	return globalDispatcher
}

// New creates a new Dispatcher instance.
func New() *Dispatcher {
	return &Dispatcher{
		listeners: make(map[string][]ListenerFunc),
	}
}

// Listen registers a listener callback for the given event name.
func (d *Dispatcher) Listen(eventName string, fn ListenerFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.listeners[eventName] = append(d.listeners[eventName], fn)
}

// Rungokna is an alias for Listen in Ngapak.
func (d *Dispatcher) Rungokna(eventName string, fn ListenerFunc) {
	d.Listen(eventName, fn)
}

// HasListeners checks if an event has at least one registered listener.
func (d *Dispatcher) HasListeners(eventName string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.listeners[eventName]) > 0
}

// Dispatch fires an event synchronously, invoking all registered listeners in registration order.
// Returns a slice of errors encountered, if any.
func (d *Dispatcher) Dispatch(eventName string, payload any) []error {
	d.mu.RLock()
	listeners := make([]ListenerFunc, len(d.listeners[eventName]))
	copy(listeners, d.listeners[eventName])
	d.mu.RUnlock()

	var errors []error
	for _, fn := range listeners {
		if err := fn(eventName, payload); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// Kirim is an alias for Dispatch in Ngapak.
func (d *Dispatcher) Kirim(eventName string, payload any) []error {
	return d.Dispatch(eventName, payload)
}

// DispatchAsync fires an event asynchronously in background goroutines.
func (d *Dispatcher) DispatchAsync(eventName string, payload any) {
	d.mu.RLock()
	listeners := make([]ListenerFunc, len(d.listeners[eventName]))
	copy(listeners, d.listeners[eventName])
	d.mu.RUnlock()

	for _, fn := range listeners {
		go func(listener ListenerFunc) {
			defer func() {
				if r := recover(); r != nil {
					// recover silently to prevent process crash
					_ = fmt.Sprintf("Event listener panic: %v", r)
				}
			}()
			_ = listener(eventName, payload)
		}(fn)
	}
}

// KirimAsync is an alias for DispatchAsync in Ngapak.
func (d *Dispatcher) KirimAsync(eventName string, payload any) {
	d.DispatchAsync(eventName, payload)
}

// Forget removes all listeners for the given event name.
func (d *Dispatcher) Forget(eventName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.listeners, eventName)
}

// Lalekna is an alias for Forget in Ngapak.
func (d *Dispatcher) Lalekna(eventName string) {
	d.Forget(eventName)
}

// Global convenience functions
func Listen(eventName string, fn ListenerFunc) {
	GetDispatcher().Listen(eventName, fn)
}

func Dispatch(eventName string, payload any) []error {
	return GetDispatcher().Dispatch(eventName, payload)
}

func DispatchAsync(eventName string, payload any) {
	GetDispatcher().DispatchAsync(eventName, payload)
}
