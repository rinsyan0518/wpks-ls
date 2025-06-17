package inmemory

import (
	"errors"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
)

type ConfigurationRepository struct {
	configuration *domain.Configuration
}

func NewConfigurationRepository() *ConfigurationRepository {
	return &ConfigurationRepository{}
}

func (r *ConfigurationRepository) Save(configuration *domain.Configuration) error {
	r.configuration = configuration
	return nil
}

func (r *ConfigurationRepository) GetConfiguration() (*domain.Configuration, error) {
	if r.configuration == nil {
		return nil, errors.New("configuration not found")
	}
	return r.configuration, nil
}

var _ out.ConfigurationRepository = (*ConfigurationRepository)(nil)
