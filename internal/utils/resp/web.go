package resp

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func OK(c *gin.Context, data interface{}) {
	Raw(c, 0, nil, data, nil)
}

func Error(c *gin.Context, err error) {
	Raw(c, 1, err, nil, nil)
}

func Raw(c *gin.Context, code int, err error, data interface{}, extra map[string]interface{}) {
	result := make(map[string]interface{})
	result["code"] = code
	result["error"] = ""
	if err != nil {
		result["error"] = err.Error()
	}
	if extra != nil {
		for k, v := range extra {
			result[k] = v
		}
	}
	result["data"] = data
	c.JSON(http.StatusOK, result)
}
