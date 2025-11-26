package logaudit

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LogAuditHandler struct {
	service LogAuditService
}

func NewLogAuditHandler(service LogAuditService) *LogAuditHandler {
	return &LogAuditHandler{
		service: service,
	}
}

func (h *LogAuditHandler) GetLogEntries(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit cannot exceed 1000"})
		return
	}

	entries, err := h.service.GetLogEntries(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve log entries"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func (h *LogAuditHandler) GetLogEntriesByService(c *gin.Context) {
	service := c.Param("service")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit cannot exceed 1000"})
		return
	}

	entries, err := h.service.GetLogEntriesByService(service, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve log entries"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func (h *LogAuditHandler) GetLogEntriesByLevel(c *gin.Context) {
	level := c.Param("level")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit cannot exceed 1000"})
		return
	}

	entries, err := h.service.GetLogEntriesByLevel(level, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve log entries"})
		return
	}

	c.JSON(http.StatusOK, entries)
}
