package handlers

import (
	"RedColarTest/internal/common"
	"RedColarTest/internal/incident/domain"
	"RedColarTest/internal/incident/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type IncidentHandler struct {
	svc *services.IncidentService
}

func NewIncidentHandler(svc *services.IncidentService) *IncidentHandler {
	return &IncidentHandler{svc: svc}
}

type createIncidentRequest struct {
	Title         string  `json:"title" binding:"required"`
	Description   *string `json:"description"`
	Latitude      float64 `json:"latitude" binding:"required"`
	Longitude     float64 `json:"longitude" binding:"required"`
	DangerRadiusM int     `json:"danger_radius_m"`
	IsActive      *bool   `json:"is_active"`
}

type updateIncidentRequest struct {
	Title         string  `json:"title" binding:"required"`
	Description   *string `json:"description"`
	Latitude      float64 `json:"latitude" binding:"required"`
	Longitude     float64 `json:"longitude" binding:"required"`
	DangerRadiusM int     `json:"danger_radius_m" binding:"required"`
	IsActive      bool    `json:"is_active" binding:"required"`
}

func (h *IncidentHandler) Create(c *gin.Context) {
	var req createIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r := req.DangerRadiusM
	if r <= 0 {
		r = 100
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	out, err := h.svc.Create(c.Request.Context(), domain.Incident{
		Title:         req.Title,
		Description:   req.Description,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		DangerRadiusM: r,
		IsActive:      isActive,
	})
	if err != nil {
		switch err.Code {
		case common.CodeIternalErr:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case common.CodeNotValid:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, out)
}

func (h *IncidentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	onlyActive := false
	if c.Query("only_active") == "true" || c.Query("only_active") == "1" {
		onlyActive = true
	}

	items, total, pageOut, sizeOut, err := h.svc.List(c.Request.Context(), page, pageSize, onlyActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     items,
		"total":     total,
		"page":      pageOut,
		"page_size": sizeOut,
	})
}

func (h *IncidentHandler) GetByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	out, errorDto := h.svc.GetByID(c.Request.Context(), id)
	if errorDto != nil {
		if errorDto.Code == common.CodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errorDto.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}

func (h *IncidentHandler) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	out, errorDto := h.svc.Update(c.Request.Context(), id, domain.Incident{
		Title:         req.Title,
		Description:   req.Description,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		DangerRadiusM: req.DangerRadiusM,
		IsActive:      req.IsActive,
	})
	if errorDto != nil {
		if errorDto.Code == common.CodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errorDto.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}

func (h *IncidentHandler) Deactivate(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	out, errorDto := h.svc.Deactivate(c.Request.Context(), id)
	if errorDto != nil {
		if errorDto.Code == common.CodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorDto.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}

func parseID(s string) (int64, *common.Error) {
	if s == "" {
		return 0, common.NewError(common.CodeNotValid, "id is empty string")
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, common.NewError(common.CodeNotValid, "cannot parse id")
	}
	return id, nil
}
