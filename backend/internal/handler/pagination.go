package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultListPerPage = 20
	maxListPerPage     = 100
)

// parseListPagination reads ?page= (1-based) and ?per_page= from the query.
// Returns page, perPage, and offset for SQL OFFSET.
func parseListPagination(c *gin.Context) (page, perPage, offset int) {
	page = 1
	perPage = defaultListPerPage
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	if pp := c.Query("per_page"); pp != "" {
		if n, err := strconv.Atoi(pp); err == nil && n > 0 {
			perPage = n
		}
	}
	if perPage > maxListPerPage {
		perPage = maxListPerPage
	}
	offset = (page - 1) * perPage
	return page, perPage, offset
}
