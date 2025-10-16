package git

import (
	"bytes"
	"context"

	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/gofiber/fiber/v2"
	helpergit "github.com/labbs/ogit/infrastructure/helpers/git"
)

func (ctrl *Controller) InfoRefs(ctx *fiber.Ctx) error {
	logger := ctrl.Logger.With().Str("component", "http.git.inforefs").Logger()

	service := ctx.Query("service")
	if service != "git-upload-pack" && service != "git-receive-pack" {
		logger.Warn().Str("service", service).Msg("Invalid service parameter")
		return ctx.Status(fiber.StatusBadRequest).SendString("invalid service parameter")
	}

	repoPath := helpergit.ExtractRepoPathFromURL(ctx.Path(), "/info/refs")
	if repoPath == "" {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	srv, ep, err := helpergit.GetTransportServer(repoPath, ctrl.StorageConfig)
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
		if err := helpergit.WriteServiceAdvertisement(ctx.Response().BodyWriter(), service); err != nil {
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
		if err := helpergit.WriteServiceAdvertisement(ctx.Response().BodyWriter(), service); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		if err := adv.Encode(ctx.Response().BodyWriter()); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
	}
	return nil
}

func (ctrl *Controller) HandleUploadPack(c *fiber.Ctx) error {
	logger := ctrl.Logger.With().Str("component", "http.git.uploadpack").Logger()

	repoPath := helpergit.ExtractRepoPathFromURL(c.Path(), "/git-upload-pack")
	if repoPath == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	srv, ep, err := helpergit.GetTransportServer(repoPath, ctrl.StorageConfig)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get transport server")
		return c.Status(fiber.StatusInternalServerError).SendString("failed to get transport server")
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

func (ctrl *Controller) HandleReceivePack(c *fiber.Ctx) error {
	logger := ctrl.Logger.With().Str("component", "http.git.receivepack").Logger()

	// Extract repository path from URL
	repoPath := helpergit.ExtractRepoPathFromURL(c.Path(), "/git-receive-pack")
	if repoPath == "" {
		logger.Error().Msg("Repository path not found")
		return c.SendStatus(fiber.StatusNotFound)
	}

	logger.Debug().Str("repoPath", repoPath).Msg("Handling receive-pack request")

	// Get the go-git transport server for this repository
	srv, ep, err := helpergit.GetTransportServer(repoPath, ctrl.StorageConfig)
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
