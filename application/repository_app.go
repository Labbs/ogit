package application

import (
	"fmt"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/gosimple/slug"
	"github.com/labbs/ogit/domain"
	"github.com/labbs/ogit/infrastructure/config"
	helperError "github.com/labbs/ogit/infrastructure/helpers/error"
	"github.com/labbs/ogit/infrastructure/storage"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type RepositoryApp struct {
	Config               config.Config
	Logger               zerolog.Logger
	RepoPres             domain.RepositoryPers
	RepositoryMemberPers domain.RepositoryMemberPers
	SSHKeyPres           domain.SSHKeyPers
	StorageConfig        storage.Config
}

func NewRepositoryApp(config config.Config, logger zerolog.Logger, repoPers domain.RepositoryPers, repositoryMemberPers domain.RepositoryMemberPers, sshKeyPers domain.SSHKeyPers, storageConfig storage.Config) *RepositoryApp {
	return &RepositoryApp{
		Config:        config,
		Logger:        logger,
		RepoPres:      repoPers,
		SSHKeyPres:    sshKeyPers,
		StorageConfig: storageConfig,
	}
}

func (c *RepositoryApp) Create(name, description, creator string, isPrivate bool) (*domain.Repository, error) {
	logger := c.Logger.With().Str("component", "application.repository.create").Logger()

	repo := &domain.Repository{
		Id:          utils.UUIDv4(),
		Name:        name,
		Description: description,
		IsPrivate:   isPrivate,
		Slug:        slug.Make(name),
	}

	err := c.RepoPres.Create(repo)
	if helperError.Catch(err) == gorm.ErrDuplicatedKey {
		logger.Warn().Str("repo_name", name).Msg("repository with the same name already exists")
		return nil, fmt.Errorf("repository with the same name already exists")
	} else if err != nil {
		logger.Error().Err(err).Str("repo_id", repo.Id).Str("repo_name", repo.Name).Msg("failed to create repository")
		return nil, err
	}

	err = c.StorageConfig.CreateRepository(repo.Slug)
	if err != nil {
		logger.Error().Err(err).Str("repo_id", repo.Id).Str("repo_name", repo.Name).Msg("failed to create repository storage")
		// Attempt to clean up the database entry if storage creation fails
		c.RepoPres.Delete(*repo)
		return nil, fmt.Errorf("failed to create repository storage: %w", err)
	}

	member := &domain.RepositoryMember{
		Id:           utils.UUIDv4(),
		RepositoryId: repo.Id,
		MemberType:   domain.MemberTypeUser,
		UserId:       &creator,
		AccessType:   domain.AccessTypeOwner,
	}

	err = c.RepositoryMemberPers.Create(member)
	if helperError.Catch(err) == gorm.ErrDuplicatedKey {
		logger.Warn().Str("repo_id", repo.Id).Str("user_id", creator).Msg("repository member already exists")
		return nil, fmt.Errorf("repository member already exists")
	} else if err != nil {
		logger.Error().Err(err).Str("repo_id", repo.Id).Str("user_id", creator).Msg("failed to create repository member")
		c.RepoPres.Delete(*repo)
		c.StorageConfig.DeleteRepository(repo.Slug)
		return nil, err
	}

	return repo, nil
}

func (c *RepositoryApp) GetRepository(id, slug *string) (*domain.Repository, error) {
	logger := c.Logger.With().Str("component", "application.repository.get").Logger()

	if id == nil && slug == nil {
		logger.Error().Msg("either id or slug must be provided")
		return nil, fmt.Errorf("either id or slug must be provided")
	}

	repo := domain.Repository{
		Id:   *id,
		Slug: *slug,
	}

	repository, err := c.RepoPres.GetRepository(repo)
	switch helperError.Catch(err) {
	case nil:
	case gorm.ErrRecordNotFound:
		logger.Warn().Str("repo_id", repo.Id).Str("repo_slug", repo.Slug).Msg("repository not found")
		return nil, fmt.Errorf("repository not found")
	default:
		logger.Error().Err(err).Str("repo_id", repo.Id).Str("repo_slug", repo.Slug).Msg("failed to get repository")
		return nil, err
	}

	return repository, nil
}

func (c *RepositoryApp) GetUserEffectiveRole(repositoryId, userId string) (domain.AccessType, error) {
	logger := c.Logger.With().Str("component", "application.repository.get_user_effective_role").Logger()

	// Récupérer tous les membres du repository (users + groups) avec les relations
	members, err := c.RepositoryMemberPers.GetByRepositoryWithRelations(repositoryId)
	if err != nil {
		logger.Error().Err(err).Str("repository_id", repositoryId).Msg("failed to get repository members")
		return domain.AccessTypeNone, err
	}

	var effectiveRole domain.AccessType = domain.AccessTypeNone

	for _, member := range members {
		switch member.MemberType {
		case domain.MemberTypeUser:
			// Permission directe utilisateur
			if member.UserId != nil && *member.UserId == userId {
				effectiveRole = domain.MaxAccessType(effectiveRole, member.AccessType)
				logger.Debug().
					Str("user_id", userId).
					Str("direct_role", string(member.AccessType)).
					Str("current_effective", string(effectiveRole)).
					Msg("Found direct user permission")
			}

		case domain.MemberTypeGroup:
			// Permission via groupe - vérifier si l'user est dans le groupe
			if member.GroupId != nil {
				// Vérifier si l'utilisateur est membre du groupe
				for _, groupMember := range member.Group.Members {
					if groupMember.Id == userId {
						effectiveRole = domain.MaxAccessType(effectiveRole, member.AccessType)
						logger.Debug().
							Str("user_id", userId).
							Str("group_id", *member.GroupId).
							Str("group_role", string(member.AccessType)).
							Str("current_effective", string(effectiveRole)).
							Msg("Found group permission")
						break
					}
				}
			}
		}
	}

	logger.Debug().
		Str("user_id", userId).
		Str("repository_id", repositoryId).
		Str("effective_role", string(effectiveRole)).
		Msg("Calculated effective role")

	return effectiveRole, nil
}

func (c *RepositoryApp) CanUserAccess(repositoryId, userId, operation string) (bool, error) {
	// Récupérer le repository pour vérifier s'il est privé
	repo, err := c.GetRepository(&repositoryId, nil)
	if err != nil {
		return false, err
	}

	// Si le repo est public et l'opération est de lecture, autoriser
	if !repo.IsPrivate && operation == "read" {
		return true, nil
	}

	// Sinon, vérifier les permissions
	effectiveRole, err := c.GetUserEffectiveRole(repositoryId, userId)
	if err != nil {
		return false, err
	}

	switch operation {
	case "read":
		return effectiveRole.CanRead(), nil
	case "write":
		return effectiveRole.CanWrite(), nil
	case "admin":
		return effectiveRole.CanAdmin(), nil
	default:
		return false, fmt.Errorf("unknown operation: %s", operation)
	}
}
