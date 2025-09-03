// Package controller provides HTTP handlers for Git Smart HTTP protocol operations.
package controller

import (
	"bytes"
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/labbs/git-server-s3/pkg/common"
	"github.com/labbs/git-server-s3/pkg/storage"
	"github.com/rs/zerolog"

	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
)

// GitController handles Git Smart HTTP protocol requests.
// It implements the server-side of the Git Smart HTTP transport protocol,
// supporting both upload-pack (clone/fetch) and receive-pack (push) operations.
type GitController struct {
	Logger  zerolog.Logger               // Logger for request logging and error reporting
	Storage storage.GitRepositoryStorage // Storage backend for Git repository operations
}

// InfoRefs handles GET requests to /{repo}/info/refs endpoint.
// This is the initial request in Git Smart HTTP protocol that advertises
// repository references (branches, tags) to the client.
//
// Query parameters:
//   - service: either "git-upload-pack" (for clone/fetch) or "git-receive-pack" (for push)
//
// Response: Git protocol formatted reference advertisement
func (gc *GitController) InfoRefs(ctx *fiber.Ctx) error {
	logger := gc.Logger.With().Str("event", "InfoRefs").Logger()

	logger.Debug().Str("repo", ctx.Params("repo")).Send()

	service := ctx.Query("service")
	if service != "git-upload-pack" && service != "git-receive-pack" {
		logger.Warn().Str("service", service).Msg("Invalid service parameter")
		return ctx.Status(fiber.StatusBadRequest).SendString("invalid service parameter")
	}

	// Extract repository path from the URL
	repoPath := common.ExtractRepoPathFromURL(ctx.Path(), "/info/refs")
	if repoPath == "" {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	// Get the go-git transport server for this repository
	srv, ep, err := common.GetTransportServer(repoPath, gc.Storage)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get transport server")
		return ctx.Status(fiber.StatusInternalServerError).SendString("failed to get transport server")
	}

	ctx.Set("Cache-Control", "no-cache")
	switch service {
	case "git-upload-pack":
		ctx.Set("Content-Type", "application/x-git-upload-pack-advertisement")
		sess, err := srv.NewUploadPackSession(ep, nil)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		adv, err := sess.AdvertisedReferences()
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		if err := common.WriteServiceAdvertisement(ctx.Response().BodyWriter(), service); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		if err := adv.Encode(ctx.Response().BodyWriter()); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
	case "git-receive-pack":
		ctx.Set("Content-Type", "application/x-git-receive-pack-advertisement")
		sess, err := srv.NewReceivePackSession(ep, nil)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		adv, err := sess.AdvertisedReferences()
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		if err := common.WriteServiceAdvertisement(ctx.Response().BodyWriter(), service); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		if err := adv.Encode(ctx.Response().BodyWriter()); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
	}
	return nil
}

// HandleUploadPack handles POST requests to /{repo}/git-upload-pack endpoint.
// This handles the actual data transfer for clone and fetch operations.
// It processes the client's wants/haves and sends back the requested pack data.
//
// Request body: Git pack protocol request (binary)
// Response: Git pack protocol response with requested objects
func (gc *GitController) HandleUploadPack(c *fiber.Ctx) error {
	logger := gc.Logger.With().Str("event", "HandleUploadPack").Logger()

	// Extract repository path from URL
	repoPath := common.ExtractRepoPathFromURL(c.Path(), "/git-upload-pack")
	if repoPath == "" {
		logger.Error().Msg("Repository path not found")
		return c.SendStatus(fiber.StatusNotFound)
	}

	logger.Debug().Str("repoPath", repoPath).Msg("Handling upload-pack request")

	// Get the go-git transport server for this repository
	srv, ep, err := common.GetTransportServer(repoPath, gc.Storage)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get transport server")
		return err
	}

	// Create an upload pack session
	sess, err := srv.NewUploadPackSession(ep, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create upload pack session")
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// Decode the upload pack request from the client
	req := packp.NewUploadPackRequest()
	if err := req.Decode(bytes.NewReader(c.Body())); err != nil {
		logger.Error().Err(err).Msg("Failed to decode upload pack request")
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	logger.Debug().Msg("Calling UploadPack")
	resp, err := sess.UploadPack(context.Background(), req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to execute upload pack")
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer resp.Close()

	c.Set("Content-Type", "application/x-git-upload-pack-result")
	logger.Debug().Msg("Encoding response")
	if err := resp.Encode(c.Response().BodyWriter()); err != nil {
		logger.Error().Err(err).Msg("Failed to encode upload pack response")
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	logger.Debug().Msg("Upload pack completed successfully")
	return nil
}

// HandleReceivePack handles POST requests to /{repo}/git-receive-pack endpoint.
// This handles the actual data transfer for push operations.
// It processes the client's reference updates and pack data, then sends back a status report.
//
// Request body: Git pack protocol request with reference updates and pack data (binary)
// Response: Git pack protocol status report indicating success/failure of each reference update
func (gc *GitController) HandleReceivePack(c *fiber.Ctx) error {
	logger := gc.Logger.With().Str("event", "HandleReceivePack").Logger()

	// Extract repository path from URL
	repoPath := common.ExtractRepoPathFromURL(c.Path(), "/git-receive-pack")
	if repoPath == "" {
		logger.Error().Msg("Repository path not found")
		return c.SendStatus(fiber.StatusNotFound)
	}

	logger.Debug().Str("repoPath", repoPath).Msg("Handling receive-pack request")

	// Get the go-git transport server for this repository
	srv, ep, err := common.GetTransportServer(repoPath, gc.Storage)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get transport server")
		return err
	}

	// Create a receive pack session
	sess, err := srv.NewReceivePackSession(ep, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create receive pack session")
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// Decode the reference update request from the client
	req := packp.NewReferenceUpdateRequest()
	if err := req.Decode(bytes.NewReader(c.Body())); err != nil {
		logger.Error().Err(err).Msg("Failed to decode receive pack request")
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// Process the receive pack request and generate a status report
	report, err := sess.ReceivePack(context.Background(), req)
	c.Set("Content-Type", "application/x-git-receive-pack-result")
	if err != nil {
		logger.Error().Err(err).Msg("Receive pack failed")
		// Even if there was an error, we still need to send the report
		_ = report.Encode(c.Response().BodyWriter())
		return nil
	}

	// Encode and send the status report back to the client
	if err := report.Encode(c.Response().BodyWriter()); err != nil {
		logger.Error().Err(err).Msg("Failed to encode receive pack report")
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	logger.Debug().Msg("Receive pack completed successfully")
	return nil
}
