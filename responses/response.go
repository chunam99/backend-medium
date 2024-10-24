package responses

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
	HTTPCode int         `json:"http_code"`
}

func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success:  true,
		Message:  message,
		Data:     data,
		HTTPCode: http.StatusOK,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string, err string) {
	c.JSON(statusCode, APIResponse{
		Success:  false,
		Message:  message,
		Error:    err,
		HTTPCode: statusCode,
	})
}
