package core

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type PaginatedData struct {
	Items      interface{} `json:"items"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
}

func RespondJSON(c *gin.Context, statusCode int, success bool, message string, data interface{}, errors interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: success,
		Message: message,
		Data:    data,
		Errors:  errors,
	})
}

func RespondSuccess(c *gin.Context, message string, data interface{}) {
	RespondJSON(c, http.StatusOK, true, message, data, nil)
}

func RespondCreated(c *gin.Context, message string, data interface{}) {
	RespondJSON(c, http.StatusCreated, true, message, data, nil)
}

func RespondError(c *gin.Context, statusCode int, message string, err interface{}) {
	RespondJSON(c, statusCode, false, message, nil, err)
}

func RespondValidationError(c *gin.Context, errors interface{}) {
	RespondJSON(c, http.StatusUnprocessableEntity, false, "Validation failed", nil, errors)
}

func RespondPaginated(c *gin.Context, items interface{}, page int, perPage int, total int64) {
	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	data := PaginatedData{
		Items:      items,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
	RespondSuccess(c, "Data retrieved successfully", data)
}
