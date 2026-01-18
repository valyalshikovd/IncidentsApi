package location

import (
	"RedColarTest/internal/common"
	location "RedColarTest/internal/locations/services"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc                *location.Service
	statsWindowMinutes int
}

func NewLocationHandler(svc *location.Service, statsWindowMinutes int) *Handler {
	return &Handler{svc: svc, statsWindowMinutes: statsWindowMinutes}
}

type LocationCheckRequest struct {
	UserID    string  `json:"user_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

func (h *Handler) LocationCheckHandler(ctx *gin.Context) {
	var req LocationCheckRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.svc.CheckLocation(ctx.Request.Context(), req.UserID, req.Latitude, req.Longitude)
	if err != nil {
		switch err.Code {
		case common.CodeNotValid:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Text, "code": err.Code})
		case common.CodeNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Text, "code": err.Code})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Text, "code": err.Code})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"dangerous": res.Dangerous,
		"incidents": res.Incidents,
	})

	go func() {
		bg, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := h.svc.RecordCheck(bg, req.UserID, req.Latitude, req.Longitude, res.Incidents); err != nil {
			log.Println("record check failed:", err)
		}
	}()
}

func (h *Handler) StatsHandler(ctx *gin.Context) {
	count, err := h.svc.Stats(ctx.Request.Context(), h.statsWindowMinutes)
	if err != nil {
		if err.Code == common.CodeNotValid {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Text})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Text, "code": err.Code})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"user_count": count})
}
