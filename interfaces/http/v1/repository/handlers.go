package repository

import (
	"github.com/gofiber/fiber/v2"
	fiberoapi "github.com/labbs/fiber-oapi"
	"github.com/labbs/ogit/infrastructure/helpers/mapper"
	"github.com/labbs/ogit/infrastructure/helpers/validator"
	"github.com/labbs/ogit/interfaces/http/v1/repository/dtos"
)

func (ctrl *Controller) GetRepository(ctx *fiber.Ctx, req dtos.GetRepositoryRequest) (*dtos.GetRepositoryResponse, *fiberoapi.ErrorResponse) {
	requestId := ctx.Locals("requestid").(string)
	logger := ctrl.Logger.With().Str("request_id", requestId).Str("component", "http.api.v1.repository.get").Logger()

	var id *string
	var slug *string

	if validator.IsValidUUID(req.Identifier) {
		id = &req.Identifier
	} else {
		slug = &req.Identifier
	}

	repo, err := ctrl.RepositoryApp.GetRepository(id, slug)
	if err != nil {
		logger.Error().Err(err).Str("identifier", req.Identifier).Msg("failed to get repository")
		return nil, &fiberoapi.ErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Details: err.Error(),
			Type:    "GET_REPOSITORY_ERROR",
		}
	}

	var response dtos.GetRepositoryResponse
	mapper.MapStructByFieldNames(repo, &response)

	return &response, nil
}

func (ctrl *Controller) CreateRepository(ctx *fiber.Ctx, req dtos.CreateRepositoryRequest) (*dtos.CreateRepositoryResponse, *fiberoapi.ErrorResponse) {
	requestId := ctx.Locals("requestid").(string)
	logger := ctrl.Logger.With().Str("request_id", requestId).Str("component", "http.api.v1.repository.create").Logger()

	// Get the authenticated user context
	authCtx, err := fiberoapi.GetAuthContext(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return nil, &fiberoapi.ErrorResponse{
			Code:    fiber.StatusUnauthorized,
			Details: "Authentication required",
			Type:    "AUTHENTICATION_REQUIRED",
		}
	}

	repo, err := ctrl.RepositoryApp.Create(req.Name, req.Description, authCtx.UserID, req.IsPrivate)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create repository")
		return nil, &fiberoapi.ErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Details: err.Error(),
			Type:    "CREATE_REPOSITORY_ERROR",
		}
	}

	var response dtos.CreateRepositoryResponse
	mapper.MapStructByFieldNames(repo, &response)

	return &response, nil
}
