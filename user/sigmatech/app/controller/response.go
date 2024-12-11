package controller

import (
	"user/sigmatech/app/constants"
	"user/sigmatech/app/service/dto/response"
	"user/sigmatech/pkg/encrypt"

	"github.com/gin-gonic/gin"
)

func RespondWithError(c *gin.Context, code int, message string, err error) {
	c.Set(constants.STATUS_CODE, code)
	responseData := response.ResponseV2{
		Success: false,
		Message: message,
		Data:    nil,
	}

	if err != nil {
		errorMsgEncrypt, encryptErr := encrypt.EncryptWithNaCl([]byte(err.Error()))
		if encryptErr != nil {
			panic(encryptErr)
		}

		if constants.Config.Environment != string(constants.Production) {
			responseData.Data = err.Error()
		} else {
			responseData.Data = errorMsgEncrypt
		}
	}

	c.AbortWithStatusJSON(code, responseData)
}

func RespondWithSuccess(c *gin.Context, code int, message string, data interface{}) {
	c.Set(constants.STATUS_CODE, code)
	c.JSON(code, response.ResponseV2{Success: true, Message: message, Data: data})
}

func RespondWithSuccessAndPagination(c *gin.Context, code int, message string, data interface{}, pagination response.Pagination) {
	c.Set(constants.STATUS_CODE, code)
	c.JSON(code, response.ResponseV3{Success: true, Message: message, Data: data, Meta: pagination})
}
