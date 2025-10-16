package application

import (
	"fmt"
	"time"

	"github.com/labbs/ogit/domain"
	"github.com/labbs/ogit/infrastructure/config"
	helperError "github.com/labbs/ogit/infrastructure/helpers/error"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type TokenApp struct {
	Config    config.Config
	Logger    zerolog.Logger
	TokenPers domain.TokenPers
}

func NewTokenApp(config config.Config, logger zerolog.Logger, tokenPers domain.TokenPers) *TokenApp {
	return &TokenApp{
		Config:    config,
		Logger:    logger,
		TokenPers: tokenPers,
	}
}

// GetByNameAndToken retrieves a token by its name and token string, ensuring it is active and not expired.
func (c *TokenApp) GetByNameAndToken(name, tokenStr string) (*domain.Token, error) {
	logger := c.Logger.With().Str("component", "application.token.get_by_name_and_token").Logger()

	logger.Debug().Str("name", name).Msg("fetching token by name and token")

	token, err := c.TokenPers.GetByNameAndToken(name, tokenStr)
	switch helperError.Catch(err) {
	case gorm.ErrRecordNotFound:
		logger.Warn().Str("name", name).Msg("token not found")
		return nil, fmt.Errorf("token not found")
	case nil:
		// No error, continue processing
	default:
		logger.Error().Err(err).Str("name", name).Msg("failed to fetch token by name and token")
		return nil, fmt.Errorf("failed to fetch token: %w", err)
	}

	if !token.Active {
		logger.Warn().Str("name", name).Msg("token inactive")
		return nil, fmt.Errorf("token inactive")
	}

	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		logger.Warn().Str("name", name).Msg("token expired")
		return nil, fmt.Errorf("token expired")
	}

	return token, nil
}

func (c *TokenApp) Create(token *domain.Token) error {
	logger := c.Logger.With().Str("component", "application.token.create").Logger()

	err := c.TokenPers.Create(token)
	if helperError.Catch(err) == gorm.ErrDuplicatedKey {
		logger.Warn().Str("token_name", token.Description).Msg("token with the same name already exists")
		return fmt.Errorf("token with the same name already exists")
	} else if err != nil {
		logger.Error().Err(err).Str("token_id", token.Id).Str("token_name", token.Description).Msg("failed to create token")
		return err
	}

	return nil
}
