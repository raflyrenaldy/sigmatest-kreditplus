package timeout

import (
	"net/http"
	"time"
	"user/sigmatech/app/constants"
	"user/sigmatech/app/controller"

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
