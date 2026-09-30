package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

type AssetController struct {
	asset usecase.Asset
}

func NewAssetController(asset usecase.Asset) *AssetController {
	return &AssetController{asset: asset}
}

func (c *AssetController) Register(r gin.IRouter) {
	r.GET("/asset-categories", c.Categories)
	r.GET("/assets", c.List)
}

func (c *AssetController) Categories(ctx *gin.Context) {
	categories, err := c.asset.Categories(ctx.Request.Context())
	if err != nil {
		apperr.Respond(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, toAssetCategoriesResponse(categories))
}

func (c *AssetController) List(ctx *gin.Context) {
	assets, err := c.asset.List(ctx.Request.Context(), ctx.Query("category"))
	if err != nil {
		apperr.Respond(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, toAssetsResponse(assets))
}
