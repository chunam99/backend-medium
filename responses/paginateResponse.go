package responses

import (
	"math"

	"github.com/gin-gonic/gin"
)

func PaginateResponse(c *gin.Context, data interface{}, count int64, page int, perPage int) {
	// Tính số trang
	totalPages := int(math.Ceil(float64(count) / float64(perPage)))

	// Trả về JSON với phân trang
	c.JSON(200, gin.H{
		"data": data,
		"pagination": gin.H{
			"count":    count,
			"page":     page,
			"pages":    totalPages,
			"per_page": perPage,
		},
	})
}
