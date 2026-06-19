package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func ginWriteError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

func ginWriteErr(c *gin.Context, err error) {
	switch {
	case pkgerrors.Is(err, pkgerrors.ErrConflict):
		ginWriteError(c, http.StatusConflict, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrUnauthorized):
		ginWriteError(c, http.StatusUnauthorized, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrForbidden):
		ginWriteError(c, http.StatusForbidden, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrNotFound):
		ginWriteError(c, http.StatusNotFound, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrBadRequest):
		ginWriteError(c, http.StatusBadRequest, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrValidation):
		ginWriteError(c, http.StatusUnprocessableEntity, err.Error())
	default:
		ginWriteError(c, http.StatusInternalServerError, "internal server error")
	}
}
