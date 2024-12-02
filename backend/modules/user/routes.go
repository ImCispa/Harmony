package user

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/:id", h.Read)
	r.PATCH("/:id", h.Update)
	r.DELETE("/:id", h.Delete)
}

func RegisterRoutesNoAuth(r *gin.RouterGroup, h *Handler) {
	r.POST("/", h.Create)
}
