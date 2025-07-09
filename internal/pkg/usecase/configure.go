package usecase

import (
	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/in"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/port/out"
)

type Configure struct {
	configurationRepository out.ConfigurationRepository
}

func NewConfigure(configurationRepository out.ConfigurationRepository) *Configure {
	return &Configure{configurationRepository: configurationRepository}
}

func (c *Configure) Configure(rootUri string, rootPath string) error {
	configuration := domain.NewConfiguration(rootUri, rootPath)
	return c.configurationRepository.Save(configuration)
}

var _ in.Configure = (*Configure)(nil)
