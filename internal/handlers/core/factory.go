// Package core provides handler factory for dependency injection.
package core

import (
	"fmt"
	
	"github.com/rs/zerolog"
)

// DefaultHandlerFactory implements HandlerFactory interface.
type DefaultHandlerFactory struct {
	config HandlerConfig
	logger zerolog.Logger
}

// NewHandlerFactory creates a new handler factory.
func NewHandlerFactory(config HandlerConfig, logger zerolog.Logger) HandlerFactory {
	return &DefaultHandlerFactory{
		config: config,
		logger: logger.With().Str("component", "handler_factory").Logger(),
	}
}

// CreateAuthHandler creates an authentication handler.
func (f *DefaultHandlerFactory) CreateAuthHandler(deps *HandlerDependencies) AuthHandler {
	// This will be implemented by importing the actual auth handler
	// For now, return nil to avoid circular imports
	f.logger.Info().Msg("Creating auth handler")
	return nil // Will be implemented in the actual auth package
}

// CreateNewsHandler creates a news handler.
func (f *DefaultHandlerFactory) CreateNewsHandler(deps *HandlerDependencies) NewsHandler {
	f.logger.Info().Msg("Creating news handler")
	return nil // Will be implemented in the actual news package
}

// CreateUserHandler creates a user handler.
func (f *DefaultHandlerFactory) CreateUserHandler(deps *HandlerDependencies) UserHandler {
	f.logger.Info().Msg("Creating user handler")
	return nil // Will be implemented in the actual user package
}

// CreateAdminHandler creates an admin handler.
func (f *DefaultHandlerFactory) CreateAdminHandler(deps *HandlerDependencies) AdminHandler {
	f.logger.Info().Msg("Creating admin handler")
	return nil // Will be implemented in the actual admin package
}

// CreateHealthHandler creates a health handler.
func (f *DefaultHandlerFactory) CreateHealthHandler(deps *HandlerDependencies) HealthHandler {
	f.logger.Info().Msg("Creating health handler")
	return nil // Will be implemented in the actual health package
}

// HandlerBuilder provides a fluent interface for building handlers.
type HandlerBuilder struct {
	deps   *HandlerDependencies
	config HandlerConfig
	logger zerolog.Logger
}

// NewHandlerBuilder creates a new handler builder.
func NewHandlerBuilder(deps *HandlerDependencies) *HandlerBuilder {
	return &HandlerBuilder{
		deps:   deps,
		config: DefaultHandlerConfig(),
		logger: deps.Logger.With().Str("component", "handler_builder").Logger(),
	}
}

// WithConfig sets the handler configuration.
func (b *HandlerBuilder) WithConfig(config HandlerConfig) *HandlerBuilder {
	b.config = config
	return b
}

// WithLogger sets the logger.
func (b *HandlerBuilder) WithLogger(logger zerolog.Logger) *HandlerBuilder {
	b.logger = logger
	return b
}

// BuildAll builds all standard handlers and registers them.
func (b *HandlerBuilder) BuildAll(registry HandlerRegistry) error {
	// factory := NewHandlerFactory(b.config, b.logger)
	
	// Create and register handlers
	handlers := []Handler{
		// These will be created by their respective packages
		// factory.CreateAuthHandler(b.deps),
		// factory.CreateNewsHandler(b.deps),
		// factory.CreateUserHandler(b.deps),
		// factory.CreateAdminHandler(b.deps),
		// factory.CreateHealthHandler(b.deps),
	}
	
	for _, handler := range handlers {
		if handler != nil {
			if err := registry.RegisterHandler(handler); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// ValidateDependencies validates that all required dependencies are present.
func (f *DefaultHandlerFactory) ValidateDependencies(deps *HandlerDependencies) error {
	if deps == nil {
		return fmt.Errorf("handler dependencies cannot be nil")
	}
	
	if deps.Config == nil {
		return fmt.Errorf("config is required")
	}
	
	if deps.NewsService == nil {
		return fmt.Errorf("news service is required")
	}
	
	if deps.UserService == nil {
		return fmt.Errorf("user service is required")
	}
	
	if deps.SearchService == nil {
		return fmt.Errorf("search service is required")
	}
	
	if deps.TrendingService == nil {
		return fmt.Errorf("trending service is required")
	}
	
	if deps.ResponseWriter == nil {
		return fmt.Errorf("response writer is required")
	}
	
	if deps.Validator == nil {
		return fmt.Errorf("validator is required")
	}
	
	if deps.ContextManager == nil {
		return fmt.Errorf("context manager is required")
	}
	
	return nil
}
