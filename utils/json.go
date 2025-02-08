package utils

import(
	"github.com/gin-gonic/gin"
)

func Validate(c *gin.Context, data interface{}) error {
    if err := c.ShouldBindJSON(data); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
        return err
    }
    return nil
}

