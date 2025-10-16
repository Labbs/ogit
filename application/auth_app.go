package application

import (
	"fmt"
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/labbs/ogit/domain"
	"github.com/labbs/ogit/infrastructure/config"
	"github.com/labbs/ogit/infrastructure/helpers/tokenutil"
	"github.com/labbs/ogit/interfaces/http/v1/auth/dtos"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type AuthApp struct {
	Config        config.Config
	Logger        zerolog.Logger
	UserApp       UserApp
	SessionApp    SessionApp
	TokenApp      TokenApp
	RepositoryApp RepositoryApp
}

func NewAuthApp(config config.Config, logger zerolog.Logger, userApp UserApp, sessionApp SessionApp, tokenApp TokenApp, repositoryApp RepositoryApp) *AuthApp {
	return &AuthApp{
		Config:        config,
		Logger:        logger,
		UserApp:       userApp,
		SessionApp:    sessionApp,
		TokenApp:      tokenApp,
		RepositoryApp: repositoryApp,
	}
}

func (c *AuthApp) Authenticate(email, password string, ctx *fiber.Ctx) (*dtos.LoginResponse, error) {
	logger := c.Logger.With().Str("component", "application.auth.authenticate").Logger()

	user, err := c.UserApp.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if !user.Active {
		logger.Warn().Str("email", email).Msg("attempt to authenticate inactive user")
		return nil, fmt.Errorf("user is not active")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		logger.Warn().Str("email", email).Msg("invalid password attempt")
		return nil, fmt.Errorf("invalid credentials")
	}

	session := &domain.Session{
		UserId:    user.Id,
		UserAgent: ctx.Get("User-Agent"),
		IpAddress: ctx.IP(),
		ExpiresAt: time.Now().Add(time.Minute * time.Duration(c.Config.Session.ExpirationMinutes)),
	}

	err = c.SessionApp.Create(session)
	if err != nil {
		logger.Error().Err(err).Str("user_id", user.Id).Msg("failed to create session")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	accessToken, err := tokenutil.CreateAccessToken(user.Id, session.Id, c.Config)
	if err != nil {
		logger.Error().Err(err).Str("user_id", user.Id).Str("session_id", session.Id).Msg("failed to create access token")
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	return &dtos.LoginResponse{
		Token: accessToken,
	}, nil
}

func (c *AuthApp) HTTPGitAuthenticate(name, token, repoSlug, operation string) (*domain.Token, error) {
	logger := c.Logger.With().Str("component", "application.auth.http_git_authenticate").Logger()

	// Récupérer le repository
	repo, err := c.RepositoryApp.GetRepository(nil, &repoSlug)
	if err != nil {
		logger.Warn().Err(err).Str("repo_slug", repoSlug).Msg("failed to get repository by slug")
		return nil, fmt.Errorf("repository not found")
	}

	// Récupérer le token
	authToken, err := c.TokenApp.GetByNameAndToken(name, token)
	if err != nil {
		logger.Warn().Err(err).Str("name", name).Msg("failed to get token by name and token")
		return nil, fmt.Errorf("invalid token")
	}

	// Vérifier les scopes du token
	var requiredScope domain.TokenScope
	switch operation {
	case "git-upload-pack": // git clone, git fetch
		requiredScope = domain.ScopeReadRepository
	case "git-receive-pack": // git push
		requiredScope = domain.ScopeWriteRepository
	default:
		return nil, fmt.Errorf("invalid git operation")
	}

	hasScope := slices.Contains(authToken.Scopes, requiredScope)
	if !hasScope {
		return nil, fmt.Errorf("token does not have required scope")
	}

	// Vérifier les permissions selon le type de token
	switch authToken.Type {
	case domain.TokenUser:
		// Token utilisateur - vérifier les permissions via user + groupes
		if authToken.User.Id == "" || !authToken.User.Active {
			return nil, fmt.Errorf("user is not active")
		}

		// Vérifier l'accès via les permissions repository_member
		gitOperation := "read"
		if operation == "git-receive-pack" {
			gitOperation = "write"
		}

		canAccess, err := c.RepositoryApp.CanUserAccess(repo.Id, authToken.User.Id, gitOperation)
		if err != nil {
			logger.Error().Err(err).Msg("failed to check user access")
			return nil, fmt.Errorf("access check failed")
		}
		if !canAccess {
			return nil, fmt.Errorf("user does not have access to repository")
		}

	case domain.TokenRepo:
		// Token repository - vérifier qu'il appartient au bon repo
		if authToken.RepositoryId == nil || *authToken.RepositoryId != repo.Id {
			return nil, fmt.Errorf("token does not belong to repository")
		}
	default:
		return nil, fmt.Errorf("invalid token type")
	}

	return authToken, nil
}

func (c *AuthApp) Register(username, email, password string) error {
	logger := c.Logger.With().Str("component", "application.auth.register").Logger()

	// check if the email is already in use
	_, err := c.UserApp.GetByEmail(email)
	if err == nil {
		return fmt.Errorf("email is already in use")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Err(err).Str("email", email).Msg("failed to hash password")
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := domain.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Active:   true,
	}

	err = c.UserApp.Create(user)
	if err != nil {
		logger.Error().Err(err).Str("email", email).Msg("failed to create user")
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}
