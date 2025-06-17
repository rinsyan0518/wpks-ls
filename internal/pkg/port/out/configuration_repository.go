package out

import "github.com/rinsyan0518/wpks-ls/internal/pkg/domain"

type ConfigurationRepository interface {
	Save(configuration *domain.Configuration) error
	GetConfiguration() (*domain.Configuration, error)
}
