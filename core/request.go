package core

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func QueryString(c *gin.Context, key string, defaultValue string) string {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	return val
}

func QueryInt(c *gin.Context, key string, defaultValue int) int {
	valStr := c.Query(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

func QueryBool(c *gin.Context, key string, defaultValue bool) bool {
	valStr := c.Query(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

func ParamInt(c *gin.Context, key string, defaultValue int) int {
	valStr := c.Param(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

func GetPaginationParams(c *gin.Context) (int, int) {
	page := QueryInt(c, "page", 1)
	if page < 1 {
		page = 1
	}

	perPage := QueryInt(c, "per_page", 10)
	if perPage < 1 {
		perPage = 10
	} else if perPage > 100 {
		perPage = 100
	}

	return page, perPage
}
