package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

type NodeController struct {
	node usecase.Node
}

func NewNodeController(node usecase.Node) *NodeController {
	return &NodeController{node: node}
}

func (c *NodeController) Register(r gin.IRouter) {
	r.GET("/tree", c.Tree)
	r.GET("/search", c.Search)
	r.POST("/nodes", c.Create)
	r.GET("/nodes/:id", c.Get)
	r.PATCH("/nodes/:id", c.Update)
	r.POST("/nodes/:id/move", c.Move)
	r.DELETE("/nodes/:id", c.Delete)
}

type createNodeRequest struct {
	ParentID        *string                 `json:"parent_id"`
	Kind            string                  `json:"kind" binding:"required"`
	Name            string                  `json:"name" binding:"required"`
	Description     string                  `json:"description"`
	Mechanics       string                  `json:"mechanics"`
	Characteristics []entity.Characteristic `json:"characteristics"`
}

type updateNodeRequest struct {
	Name            *string                  `json:"name"`
	Description     *string                  `json:"description"`
	Mechanics       *string                  `json:"mechanics"`
	Characteristics *[]entity.Characteristic `json:"characteristics"`
}

type moveNodeRequest struct {
	ParentID *string `json:"parent_id"`
}

func (c *NodeController) Tree(ctx *gin.Context) {
	tree, err := c.node.Tree(ctx.Request.Context())
	if err != nil {
		apperr.Respond(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, toTreeResponse(tree))
}

func (c *NodeController) Search(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	hits, err := c.node.Search(ctx.Request.Context(), ctx.Query("q"), limit)
	if err != nil {
		apperr.Respond(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, toSearchResponse(hits))
}

func (c *NodeController) Get(ctx *gin.Context) {
	node, err := c.node.GetByID(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		apperr.Respond(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, toNodeResponse(node))
}

func (c *NodeController) Create(ctx *gin.Context) {
	var req createNodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apperr.Bind(ctx, err)
		return
	}
	node, err := c.node.Create(ctx.Request.Context(), usecase.CreateNodeInput{
		ParentID:        req.ParentID,
		Kind:            req.Kind,
		Name:            req.Name,
		Description:     req.Description,
		Mechanics:       req.Mechanics,
		Characteristics: req.Characteristics,
	})
	if err != nil {
		apperr.Respond(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, toNodeResponse(node))
}

func (c *NodeController) Update(ctx *gin.Context) {
	var req updateNodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apperr.Bind(ctx, err)
		return
	}
	node, err := c.node.Update(ctx.Request.Context(), ctx.Param("id"), usecase.UpdateNodeInput{
		Name:            req.Name,
		Description:     req.Description,
		Mechanics:       req.Mechanics,
		Characteristics: req.Characteristics,
	})
	if err != nil {
		apperr.Respond(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, toNodeResponse(node))
}

func (c *NodeController) Move(ctx *gin.Context) {
	var req moveNodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apperr.Bind(ctx, err)
		return
	}
	node, err := c.node.Move(ctx.Request.Context(), ctx.Param("id"), req.ParentID)
	if err != nil {
		apperr.Respond(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, toNodeResponse(node))
}

func (c *NodeController) Delete(ctx *gin.Context) {
	if err := c.node.Delete(ctx.Request.Context(), ctx.Param("id")); err != nil {
		apperr.Respond(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
