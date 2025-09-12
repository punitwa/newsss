// Package core provides handler registry for loose coupling.
package core

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

// DefaultHandlerRegistry implements HandlerRegistry interface.
type DefaultHandlerRegistry struct {
	handlers map[string]Handler
	mutex    sync.RWMutex
	logger   zerolog.Logger
}

// NewHandlerRegistry creates a new handler registry.
func NewHandlerRegistry(logger zerolog.Logger) HandlerRegistry {
	return &DefaultHandlerRegistry{
		handlers: make(map[string]Handler),
		logger:   logger.With().Str("component", "handler_registry").Logger(),
	}
}

// RegisterHandler registers a handler.
func (r *DefaultHandlerRegistry) RegisterHandler(handler Handler) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	name := handler.GetName()
	if name == "" {
		return fmt.Errorf("handler name cannot be empty")
	}
	
	if _, exists := r.handlers[name]; exists {
		return fmt.Errorf("handler with name '%s' already registered", name)
	}
	
	r.handlers[name] = handler
	
	r.logger.Info().
		Str("handler_name", name).
		Str("base_path", handler.GetBasePath()).
		Msg("Handler registered")
	
	return nil
}

// GetHandler retrieves a handler by name.
func (r *DefaultHandlerRegistry) GetHandler(name string) (Handler, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	handler, exists := r.handlers[name]
	if !exists {
		return nil, fmt.Errorf("handler with name '%s' not found", name)
	}
	
	return handler, nil
}

// GetAllHandlers returns all registered handlers.
func (r *DefaultHandlerRegistry) GetAllHandlers() []Handler {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	handlers := make([]Handler, 0, len(r.handlers))
	for _, handler := range r.handlers {
		handlers = append(handlers, handler)
	}
	
	return handlers
}

// GetHandlersByType returns handlers of a specific type.
func (r *DefaultHandlerRegistry) GetHandlersByType(handlerType string) []Handler {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var handlers []Handler
	for _, handler := range r.handlers {
		// This is a simple type check based on interface assertion
		switch handlerType {
		case "auth":
			if _, ok := handler.(AuthHandler); ok {
				handlers = append(handlers, handler)
			}
		case "news":
			if _, ok := handler.(NewsHandler); ok {
				handlers = append(handlers, handler)
			}
		case "user":
			if _, ok := handler.(UserHandler); ok {
				handlers = append(handlers, handler)
			}
		case "admin":
			if _, ok := handler.(AdminHandler); ok {
				handlers = append(handlers, handler)
			}
		case "health":
			if _, ok := handler.(HealthHandler); ok {
				handlers = append(handlers, handler)
			}
		}
	}
	
	return handlers
}

// GetHandlerCount returns the number of registered handlers.
func (r *DefaultHandlerRegistry) GetHandlerCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	return len(r.handlers)
}

// UnregisterHandler removes a handler from the registry.
func (r *DefaultHandlerRegistry) UnregisterHandler(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.handlers[name]; !exists {
		return fmt.Errorf("handler with name '%s' not found", name)
	}
	
	delete(r.handlers, name)
	
	r.logger.Info().
		Str("handler_name", name).
		Msg("Handler unregistered")
	
	return nil
}

// Clear removes all handlers from the registry.
func (r *DefaultHandlerRegistry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	count := len(r.handlers)
	r.handlers = make(map[string]Handler)
	
	r.logger.Info().
		Int("handlers_cleared", count).
		Msg("All handlers cleared from registry")
}
