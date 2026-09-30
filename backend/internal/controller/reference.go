package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

type ReferenceController struct {
	reference usecase.Reference
	maxSize   int64
}

func NewReferenceController(reference usecase.Reference, maxSize int64) *ReferenceController {
	return &ReferenceController{reference: reference, maxSize: maxSize}
}

func (c *ReferenceController) Register(api gin.IRouter, root gin.IRouter) {
	api.POST("/nodes/:id/references", c.Upload)
	api.DELETE("/references/:id", c.Delete)
	root.GET("/uploads/:name", c.Serve)
}

func (c *ReferenceController) Upload(ctx *gin.Context) {
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

	ref, err := c.reference.Upload(ctx.Request.Context(), usecase.UploadReferenceInput{
		NodeID:   ctx.Param("id"),
		Filename: header.Filename,
		Content:  file,
	})
	if err != nil {
		apperr.Respond(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, toReferenceResponse(ref))
}

func (c *ReferenceController) Delete(ctx *gin.Context) {
	if err := c.reference.Delete(ctx.Request.Context(), ctx.Param("id")); err != nil {
		apperr.Respond(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *ReferenceController) Serve(ctx *gin.Context) {
	file, err := c.reference.Open(ctx.Request.Context(), ctx.Param("name"))
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
	// Uploaded files are user content: never let them run scripts in our origin.
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Content-Security-Policy", "sandbox; default-src 'none'; img-src 'self'; media-src 'self'; style-src 'unsafe-inline'")
	h.Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(ctx.Writer, ctx.Request, "", file.ModTime, file.Content)
}
