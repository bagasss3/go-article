package handler

import (
	"net/http"

	"github.com/bagasss3/go-article/internal/config"
	"github.com/bagasss3/go-article/internal/helper"
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/bagasss3/go-article/pkg/response"
	"github.com/labstack/echo/v4"
)

type authorHandler struct {
	authorService model.AuthorMethodService
}

func NewAuthorHandler(authorService model.AuthorMethodService) *authorHandler {
	return &authorHandler{
		authorService: authorService,
	}
}

func (h *authorHandler) Register(g *echo.Group) {
	api := g.Group("/author")
	{
		api.POST("", h.create)
		api.GET("/:id", h.getByID)
	}
}

func (h *authorHandler) create(c echo.Context) error {
	var req *model.CreateAuthorRequest

	if err := c.Bind(&req); err != nil {
		return response.ResponseInterfaceError(c, http.StatusBadRequest, err.Error(), config.BadRequest)
	}

	if err := c.Validate(req); err != nil {
		return response.ResponseInterfaceError(c, http.StatusBadRequest, config.BadRequest, helper.GetValueBetween(err.Error(), "Error:", "tag"))
	}

	result, err := h.authorService.Create(c.Request().Context(), req)
	if err != nil {
		return handleError(c, err)
	}

	return response.ResponseInterface(c, http.StatusCreated, result, "Store Author")
}

func (h *authorHandler) getByID(c echo.Context) error {
	id := c.Param("id")

	result, err := h.authorService.FindByID(c.Request().Context(), id)
	if err != nil {
		return handleError(c, err)
	}

	return response.ResponseInterface(c, http.StatusOK, result, "Find Author By ID")
}
