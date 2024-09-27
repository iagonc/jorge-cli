package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/internal/schemas"
)

// CreateResourceHandler é responsável por chamar o use case para criar o recurso
func (h *Handler) CreateResourceHandler(ctx *gin.Context) {
    ValidateEmptyRequest(ctx)
    if ctx.IsAborted() {
        return
    }

    var request schemas.Resource
    ValidateRequiredFields(ctx, &request)
    if ctx.IsAborted() {
        return
    }

    err := h.CreateResourceUseCase.Execute(&request)
    if err != nil {
        SendError(ctx, http.StatusBadRequest, err.Error())
        return
    }

    SendSuccess(ctx, "create-resource", &request)
}

