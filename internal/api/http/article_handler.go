package handler

import (
	"net/http"

	"github.com/bagasss3/go-article/internal/config"
	"github.com/bagasss3/go-article/internal/helper"
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/bagasss3/go-article/pkg/response"
	"github.com/labstack/echo/v4"
)

type articleHandler struct {
	articleService model.ArticleMethodService
}

func NewArticleHandler(articleService model.ArticleMethodService) *articleHandler {
	return &articleHandler{
		articleService: articleService,
	}
}

func (h *articleHandler) Register(g *echo.Group) {
	api := g.Group("/article")
	{
		api.GET("", h.getAll)
		api.POST("", h.create)
	}
}
func (h *articleHandler) getAll(c echo.Context) error {
	var query model.ArticleQuery

	if err := c.Bind(&query); err != nil {
		return response.ResponseInterfaceError(c, http.StatusBadRequest, err.Error(), config.BadRequest)
	}

	articles, total, err := h.articleService.FindAll(c.Request().Context(), query)
	if err != nil {
		return handleError(c, err)
	}

	return response.ResponseInterfaceTotal(c, 200, articles, "List Article", int(total))
}

func (h *articleHandler) create(c echo.Context) error {
	var req *model.CreateArticleRequest

	if err := c.Bind(&req); err != nil {
		return response.ResponseInterfaceError(c, http.StatusBadRequest, err.Error(), config.BadRequest)
	}

	if err := c.Validate(req); err != nil {
		return response.ResponseInterfaceError(c, http.StatusBadRequest, config.BadRequest, helper.GetValueBetween(err.Error(), "Error:", "tag"))
	}

	result, err := h.articleService.Create(c.Request().Context(), req)
	if err != nil {
		return handleError(c, err)
	}

	return response.ResponseInterface(c, http.StatusCreated, result, "Store Article")
}
