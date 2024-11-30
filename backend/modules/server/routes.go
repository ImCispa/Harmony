package server

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.POST("/", h.Create)
	r.GET("/:id", h.Read)
	r.PATCH("/:id", h.Update)
	r.DELETE("/:id", h.Delete)
	r.GET("/:id/invite", h.Invite)
	r.POST("/:id/join", h.Join)
	r.POST("/:id/leave", h.Leave)
}
