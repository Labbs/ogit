package application

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/utils"
	fiberoapi "github.com/labbs/fiber-oapi"
	"github.com/labbs/ogit/domain"
	"github.com/labbs/ogit/infrastructure/config"
	"github.com/labbs/ogit/infrastructure/helpers/tokenutil"
	"github.com/rs/zerolog"
)

type SessionApp struct {
	Config      config.Config
	Logger      zerolog.Logger
	SessionPers domain.SessionPers
	UserApp     *UserApp
}

func NewSessionApp(config config.Config, logger zerolog.Logger, sessionPers domain.SessionPers, userApp *UserApp) *SessionApp {
	return &SessionApp{
		Config:      config,
		Logger:      logger,
		SessionPers: sessionPers,
		UserApp:     userApp,
	}
}

func (c *SessionApp) Create(session *domain.Session) error {
	logger := c.Logger.With().Str("component", "application.session.create").Logger()

	session.Id = utils.UUIDv4()

	err := c.SessionPers.Create(session)
	if err != nil {
		logger.Error().Err(err).Str("session_id", session.Id).Str("user_id", session.UserId).Msg("failed to create session")
		return err
	}

	return nil
}

func (c *SessionApp) DeleteExpired() error {
	// logger := c.Logger.With().Str("component", "application.session.delete_expired").Logger()

	return nil
}

func (c *SessionApp) ValidateToken(token string) (*fiberoapi.AuthContext, error) {
	logger := c.Logger.With().Str("component", "application.session.validate_token").Logger()

	sessionId, err := tokenutil.GetSessionIdFromToken(token, c.Config)
	if err != nil {
		logger.Error().Err(err).Str("token", token).Msg("failed to get session id from token")
		return nil, fmt.Errorf("invalid token")
	}

	session, err := c.SessionPers.GetById(sessionId)
	if err != nil {
		logger.Error().Err(err).Str("token", token).Msg("failed to get session by token")
		return nil, fmt.Errorf("invalid token")
	}

	if session.ExpiresAt.Before(time.Now()) {
		logger.Warn().Str("token", token).Msg("session has expired")
		return nil, fmt.Errorf("session has expired")
	}

	ctx := &fiberoapi.AuthContext{
		UserID: session.UserId,
	}

	return ctx, nil
}

func (c *SessionApp) HasRole(ctx *fiberoapi.AuthContext, role string) bool {
	logger := c.Logger.With().Str("component", "application.session.has_role").Logger()

	logger.Warn().Msg("not implemented")

	return false
}

func (c *SessionApp) HasScope(ctx *fiberoapi.AuthContext, scope string) bool {
	logger := c.Logger.With().Str("component", "application.session.has_scope").Logger()

	logger.Warn().Msg("not implemented")

	return false
}

func (c *SessionApp) CanAccessResource(ctx *fiberoapi.AuthContext, resourceType, resourceID, action string) (bool, error) {
	logger := c.Logger.With().Str("component", "application.session.can_access_resource").Logger()

	logger.Warn().Msg("not implemented")

	return false, fmt.Errorf("not implemented")
}

func (c *SessionApp) GetUserPermissions(ctx *fiberoapi.AuthContext, resourceType, resourceID string) (*fiberoapi.ResourcePermission, error) {
	logger := c.Logger.With().Str("component", "application.session.get_user_permissions").Logger()

	logger.Warn().Msg("not implemented")

	return nil, fmt.Errorf("not implemented")
}
