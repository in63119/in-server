package httputil

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"in-server/pkg/apperr"
)

func WriteError(c *gin.Context, err error) {
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		c.JSON(appErr.Status, gin.H{"code": appErr.Code, "message": appErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_SERVER_ERROR", "message": "internal server error"})
}
