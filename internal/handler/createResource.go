package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/internal/schemas"
)

// CreateResourceHandler handles the creation of a new resource by calling the use case.
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
