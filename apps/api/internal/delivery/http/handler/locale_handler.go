package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type LocaleHandler struct {
	supportedLocales []string
}

func NewLocaleHandler(supportedLocales []string) *LocaleHandler {
	return &LocaleHandler{supportedLocales: supportedLocales}
}

func (h *LocaleHandler) List(ginCtx *gin.Context) {
	ginCtx.JSON(http.StatusOK, h.supportedLocales)
}
