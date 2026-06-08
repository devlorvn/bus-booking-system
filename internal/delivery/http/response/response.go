package response

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, gin.H{
		"message": message,
		"data":    data,
	})
}

func Error(c *gin.Context, status int, message string, code any) {
	c.JSON(status, gin.H{
		"error": message,
		"code":  code,
	})
}
