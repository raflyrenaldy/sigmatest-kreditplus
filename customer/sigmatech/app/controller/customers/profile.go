package customers

import (
	"customer/sigmatech/app/constants"
	"customer/sigmatech/app/controller"
	cif_DBModels "customer/sigmatech/app/db/dto/customer_information_files"
	customers_DBModels "customer/sigmatech/app/db/dto/customers"
	"encoding/json"
	"fmt"
	"net/http"

	"customer/sigmatech/app/service/correlation"
	"customer/sigmatech/app/service/dto/request"
	"customer/sigmatech/app/service/logger"
	"customer/sigmatech/app/service/util"

	reqCustomer "customer/sigmatech/app/service/dto/request/customer"

	"github.com/gin-gonic/gin"
)

func (u CustomerController) GetProfile(c *gin.Context) {
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

	customer, err := u.CustomerDBClient.GetCustomer(ctx, fmt.Sprintf("%s='%s'",
		customers_DBModels.COLUMN_EMAIL, usr.Email,
	))
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}
	customer.Password = "" // Clear the password field for security reasons

	fCIF := fmt.Sprintf("%s='%s'", cif_DBModels.COLUMN_CUSTOMER_UUID, customer.Uuid)

	cif, err := u.CIFDBClient.GetCustomerInformationFile(ctx, fCIF)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	p := request.Pagination{
		GetAllData: true,
	}
	p.Validate()

	// Create the vendor profile struct
	customerProfile := struct {
		Customer customers_DBModels.Customer          `json:"customer"`
		CIF      cif_DBModels.CustomerInformationFile `json:"cif"`
	}{
		Customer: customer,
		CIF:      cif,
	}

	// Respond with success message and the customer profile (with password field cleared)
	controller.RespondWithSuccess(c, http.StatusOK, constants.GET_SUCCESSFULLY, customerProfile)
}

func (u CustomerController) UpdateProfile(c *gin.Context) {
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

	dataFromBody := cif_DBModels.CustomerInformationFile{}       // Create an empty Customers struct
	err := json.NewDecoder(c.Request.Body).Decode(&dataFromBody) // Decode the request body into the Customers struct
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	var patcher = make(map[string]interface{}) // Create a patcher map to hold the fields to be updated

	// Check if each field is present in the request body and add it to the patcher map if not empty
	if dataFromBody.FullName != "" {
		patcher[cif_DBModels.COLUMN_FULL_NAME] = dataFromBody.FullName
	}

	if dataFromBody.LegalName != "" {
		patcher[cif_DBModels.COLUMN_LEGAL_NAME] = dataFromBody.LegalName
	}

	if dataFromBody.Gender != nil {
		patcher[cif_DBModels.COLUMN_GENDER] = dataFromBody.Gender
	}

	if dataFromBody.PlaceOfBirth != "" {
		patcher[cif_DBModels.COLUMN_PLACE_OF_BIRTH] = dataFromBody.PlaceOfBirth
	}

	if dataFromBody.DateOfBirth != nil {
		patcher[cif_DBModels.COLUMN_DATE_OF_BIRTH] = dataFromBody.DateOfBirth
	}

	if dataFromBody.Salary >= 0 {
		patcher[cif_DBModels.COLUMN_SALARY] = dataFromBody.Salary
	}

	filter := fmt.Sprintf(`%s='%s'`, cif_DBModels.COLUMN_CUSTOMER_UUID, usr.Uuid) // Create a filter string to match the customer ID

	if err := u.CIFDBClient.UpdateCustomerInformationFile(ctx, filter, patcher); err != nil {
		if constraintName := util.ExtractConstraintName(err.Error()); constraintName != "" {
			controller.RespondWithError(c, http.StatusConflict, fmt.Sprintf("%s already exists", constraintName), err)
			return
		}

		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	r, _ := u.CustomerDBClient.GetCustomer(ctx, fmt.Sprintf("%s='%s'",
		customers_DBModels.COLUMN_EMAIL, usr.Email,
	))
	r.Password = ""

	cifData, _ := u.CIFDBClient.GetCustomerInformationFile(ctx, fmt.Sprintf("%s='%s'",
		cif_DBModels.COLUMN_CUSTOMER_UUID, usr.Uuid,
	))

	// Create the vendor profile struct
	customerProfile := struct {
		Customer customers_DBModels.Customer          `json:"customer"`
		CIF      cif_DBModels.CustomerInformationFile `json:"cif"`
	}{
		Customer: r,
		CIF:      cifData,
	}

	// Respond with success message and the updated customer profile (with password field cleared)
	controller.RespondWithSuccess(c, http.StatusAccepted, constants.UPDATED_SUCCESSFULLY, customerProfile)
}

func (u CustomerController) UpdateProfilePassword(c *gin.Context) {
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

	dataFromBody := reqCustomer.UpdatePassword{}                 // Create an empty Customers struct
	err := json.NewDecoder(c.Request.Body).Decode(&dataFromBody) // Decode the request body into the Customers struct
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	if err := dataFromBody.Validate(); err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	if !util.ValidatePassword(dataFromBody.OldPassword, usr.Password) {
		log.Errorf("Wrong credentials")
		controller.RespondWithError(c, http.StatusUnauthorized, "Password lama salah.", nil)
		return
	}

	hashedPassword, err := util.GenerateHash(dataFromBody.Password) // Generate a hashed password
	if err != nil {
		log.Errorf(constants.INTERNAL_SERVER_ERROR, err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	filter := fmt.Sprintf("%s='%s'", customers_DBModels.COLUMN_EMAIL, usr.Email) // Create a filter string to match the customer ID

	if err := u.CustomerDBClient.UpdateCustomer(ctx, filter, map[string]interface{}{
		customers_DBModels.COLUMN_PASSWORD: hashedPassword,
	}); err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	r, _ := u.CustomerDBClient.GetCustomer(ctx, fmt.Sprintf("%s='%s'",
		customers_DBModels.COLUMN_EMAIL, usr.Email,
	))
	r.Password = ""

	// Respond with success message and the updated customer profile (with password field cleared)
	controller.RespondWithSuccess(c, http.StatusAccepted, constants.UPDATED_SUCCESSFULLY, r)
}
