package application

import (
	"fmt"

	"github.com/labbs/ogit/domain"
	"github.com/labbs/ogit/infrastructure/config"
	helperError "github.com/labbs/ogit/infrastructure/helpers/error"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type UserApp struct {
	Config    config.Config
	Logger    zerolog.Logger
	UserPres  domain.UserPers
	GroupPres domain.GroupPers
}

func NewUserApp(config config.Config, logger zerolog.Logger, userPers domain.UserPers, groupPers domain.GroupPers) *UserApp {
	return &UserApp{
		Config:    config,
		Logger:    logger,
		UserPres:  userPers,
		GroupPres: groupPers,
	}
}

func (c *UserApp) GetByEmail(email string) (*domain.User, error) {
	logger := c.Logger.With().Str("component", "application.user.get_by_email").Logger()

	user, err := c.UserPres.GetByEmail(email)
	switch helperError.Catch(err) {
	case gorm.ErrRecordNotFound:
		logger.Warn().Str("email", email).Msg("user not found")
		return nil, fmt.Errorf("user not found")
	case nil:
		// no error
	default:
		logger.Error().Err(err).Str("email", email).Msg("failed to get user by email")
		return nil, err
	}
	return &user, nil
}

func (c *UserApp) Create(user domain.User) error {
	logger := c.Logger.With().Str("component", "application.user.create").Logger()

	err := c.UserPres.Create(&user)
	if helperError.Catch(err) == gorm.ErrDuplicatedKey {
		logger.Warn().Str("email", user.Email).Msg("user with the same email already exists")
		return fmt.Errorf("user with the same email already exists")
	} else if err != nil {
		logger.Error().Err(err).Str("email", user.Email).Msg("failed to create user")
		return err
	}

	return nil
}
