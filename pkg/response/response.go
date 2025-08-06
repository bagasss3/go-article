package response

import (
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/labstack/echo/v4"
)

func ResponseInterface(c echo.Context, statusServer int, res any, msg string) error {
	c.JSON(statusServer, model.JsonResponse{
		RequestId: c.Response().Header().Get(echo.HeaderXRequestID),
		Status:    statusServer,
		Messages:  msg,
		Data:      res,
	})
	return nil
}

func ResponseInterfaceTotal(c echo.Context, statusServer int, res any, msg string, total int) error {
	c.JSON(statusServer, model.JsonResponseTotal{
		RequestId: c.Response().Header().Get(echo.HeaderXRequestID),
		Status:    statusServer,
		Messages:  msg,
		Data:      res,
		Total:     total,
	})
	return nil
}

func ResponseInterfaceError(c echo.Context, statusServer int, res any, msg string) error {
	c.JSON(statusServer, model.JsonResponsError{
		RequestId:        c.Response().Header().Get(echo.HeaderXRequestID),
		StatusCode:       statusServer,
		ErrorCode:        statusServer,
		ErrorMessage:     msg,
		DeveloperMessage: res,
	})
	return nil
}
