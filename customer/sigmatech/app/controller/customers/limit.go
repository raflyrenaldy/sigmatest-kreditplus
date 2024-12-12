package customers

import (
	"customer/sigmatech/app/constants"
	"customer/sigmatech/app/controller"
	customerLimits_DBModels "customer/sigmatech/app/db/dto/customer_limits"
	customers_DBModels "customer/sigmatech/app/db/dto/customers"
	"fmt"
	"net/http"

	"customer/sigmatech/app/service/correlation"
	"customer/sigmatech/app/service/dto/request"
	"customer/sigmatech/app/service/logger"
	"github.com/gin-gonic/gin"
)

func (u CustomerController) GetLimits(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger

	context, exist := c.Get(constants.CTK_CLAIM_KEY.String()) // Retrieve the customer context from the request
	if !exist {
		errorMsg := fmt.Sprintf("%s: %s", constants.UNAUTHORIZED_ACCESS, constants.INVALID_TOKEN)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusUnauthorized, errorMsg, nil)
		return
	}
	usr := context.(*customers_DBModels.Customer) // Type assertion to retrieve the customer information

	p := request.Pagination{
		GetAllData: true,
	}
	p.Validate()

	f := request.ExtractFilteredQueryParams(c, customerLimits_DBModels.CustomerLimit{})
	f[customerLimits_DBModels.COLUMN_CUSTOMER_UUID] = usr.Uuid.String()

	customerLimits, _, err := u.CustomerLimitDBClient.GetCustomerLimits(ctx, p, f)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	// Respond with success message and the customer profile (with password field cleared)
	controller.RespondWithSuccess(c, http.StatusOK, constants.GET_SUCCESSFULLY, customerLimits)
}
