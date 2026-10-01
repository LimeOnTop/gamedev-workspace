package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

type ModelController struct {
	model   usecase.Model
	maxSize int64
}

func NewModelController(model usecase.Model, maxSize int64) *ModelController {
	return &ModelController{model: model, maxSize: maxSize}
}

func (c *ModelController) Register(api gin.IRouter, root gin.IRouter) {
	api.PUT("/nodes/:id/model", c.Upload)
	api.DELETE("/nodes/:id/model", c.Delete)
	root.GET("/uploads/models/:name", c.Serve)
}

func (c *ModelController) Upload(ctx *gin.Context) {
	// Leave headroom for multipart framing; the service enforces the exact limit.
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, c.maxSize+1<<20)
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			apperr.Respond(ctx, apperr.ErrTooLarge)
			return
		}
		apperr.Respond(ctx, apperr.Validation("multipart field 'file' is required"))
		return
	}
	defer file.Close()

	model, err := c.model.Upload(ctx.Request.Context(), usecase.UploadModelInput{
		NodeID:   ctx.Param("id"),
		Filename: header.Filename,
		Content:  file,
	})
	if err != nil {
		apperr.Respond(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, toModelResponse(model))
}

func (c *ModelController) Delete(ctx *gin.Context) {
	if err := c.model.Delete(ctx.Request.Context(), ctx.Param("id")); err != nil {
		apperr.Respond(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *ModelController) Serve(ctx *gin.Context) {
	file, err := c.model.Open(ctx.Request.Context(), ctx.Param("name"))
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			ctx.Status(http.StatusNotFound)
			return
		}
		apperr.Respond(ctx, err)
		return
	}
	defer file.Content.Close()

	h := ctx.Writer.Header()
	h.Set("Content-Type", file.ContentType)
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Content-Security-Policy", "sandbox; default-src 'none'")
	h.Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(ctx.Writer, ctx.Request, "", file.ModTime, file.Content)
}
