package timeout

import (
	"customer/sigmatech/app/constants"
	"customer/sigmatech/app/controller"
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(15*time.Second),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
		timeout.WithResponse(timeoutHandler),
	)
}

func timeoutHandler(c *gin.Context) {
	controller.RespondWithError(c, http.StatusRequestTimeout, constants.TIMEOUT_ERROR, nil)
	return
}
