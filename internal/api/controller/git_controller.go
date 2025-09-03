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

type GitController struct {
	Logger  zerolog.Logger
	Storage storage.GitRepositoryStorage
}

func (gc *GitController) InfoRefs(ctx *fiber.Ctx) error {
	logger := gc.Logger.With().Str("event", "InfoRefs").Logger()

	logger.Debug().Str("repo", ctx.Params("repo")).Send()

	service := ctx.Query("service")
	if service != "git-upload-pack" && service != "git-receive-pack" {
		logger.Warn().Str("service", service).Msg("Invalid service parameter")
		return ctx.Status(fiber.StatusBadRequest).SendString("invalid service parameter")
	}

	repoPath := common.ExtractRepoPathFromURL(ctx.Path(), "/info/refs")
	if repoPath == "" {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

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

func (gc *GitController) HandleUploadPack(c *fiber.Ctx) error {
	logger := gc.Logger.With().Str("event", "HandleUploadPack").Logger()

	repoPath := common.ExtractRepoPathFromURL(c.Path(), "/git-upload-pack")
	if repoPath == "" {
		logger.Error().Msg("Repository path not found")
		return c.SendStatus(fiber.StatusNotFound)
	}

	logger.Debug().Str("repoPath", repoPath).Msg("Handling upload-pack request")

	srv, ep, err := common.GetTransportServer(repoPath, gc.Storage)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get transport server")
		return err
	}

	sess, err := srv.NewUploadPackSession(ep, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create upload pack session")
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

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

func (gc *GitController) HandleReceivePack(c *fiber.Ctx) error {
	repoPath := common.ExtractRepoPathFromURL(c.Path(), "/git-receive-pack")
	if repoPath == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	srv, ep, err := common.GetTransportServer(repoPath, gc.Storage)
	if err != nil {
		return err
	}

	sess, err := srv.NewReceivePackSession(ep, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	req := packp.NewReferenceUpdateRequest()
	if err := req.Decode(bytes.NewReader(c.Body())); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	report, err := sess.ReceivePack(context.Background(), req)
	c.Set("Content-Type", "application/x-git-receive-pack-result")
	if err != nil {
		_ = report.Encode(c.Response().BodyWriter())
		return nil
	}
	if err := report.Encode(c.Response().BodyWriter()); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return nil
}
