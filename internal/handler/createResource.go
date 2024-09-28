package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/internal/schemas"
)

// CreateResourceHandler handles the creation of a new resource by calling the use case.
// CreateResourceHandler creates a new resource
// @Summary Create a resource
// @Description Create a new resource with the provided JSON body
// @Tags Resources
// @Accept  json
// @Produce  json
// @Param resource body schemas.Resource true "Resource Data"
// @Success 200 {object} schemas.Resource "Successfully created"
// @Failure 400 {object} gin.H "Bad request"
// @Router /resource [post]
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

    err := h.CreateResource.Execute(&request)
    if err != nil {
        SendError(ctx, http.StatusBadRequest, err.Error())
        return
    }

    SendSuccess(ctx, "create-resource", &request)
}
