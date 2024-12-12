package customer

import (
	"errors"
	"github.com/google/uuid"
	"time"
	"user/sigmatech/app/constants"
	"user/sigmatech/app/controller"
	cif_DBModels "user/sigmatech/app/db/dto/customer_information_files"
	customers_DBModels "user/sigmatech/app/db/dto/customers"
	customerDB "user/sigmatech/app/db/repository/customer"
	cifDB "user/sigmatech/app/db/repository/customer_information_file"
	customerLimitDB "user/sigmatech/app/db/repository/customer_limit"

	"encoding/json"
	"fmt"

	"net/http"
	"user/sigmatech/app/service/correlation"
	"user/sigmatech/app/service/dto/request"
	"user/sigmatech/app/service/logger"
	"user/sigmatech/app/service/util"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ICustomerController is an interface that defines the methods for a user controller.
type ICustomerController interface {
	GetCustomers(c *gin.Context)
	GetCustomersDetail(c *gin.Context)
	GetCustomer(c *gin.Context)
	UpdateCustomer(c *gin.Context)
	UpdateCustomerPassword(c *gin.Context)
	DeleteCustomer(c *gin.Context)
	DeleteCustomers(c *gin.Context)
}

// CustomerController is a struct that implements the ICustomerController interface.
type CustomerController struct {
	CustomerDBClient      customerDB.ICustomerRepository // customerDB represents the database client for crm-user-related operations.
	CustomerLimitDBClient customerLimitDB.ICustomerLimitRepository
	CifDBClient           cifDB.ICustomerInformationFileRepository
}

// NewCustomerController is a constructor function that creates a new CustomerController.
func NewCustomerController(
	CustomerDBClient customerDB.ICustomerRepository,
	CustomerLimitDBClient customerLimitDB.ICustomerLimitRepository,
	CifDBClient cifDB.ICustomerInformationFileRepository,
) ICustomerController {
	return &CustomerController{
		CustomerDBClient:      CustomerDBClient,
		CustomerLimitDBClient: CustomerLimitDBClient,
		CifDBClient:           CifDBClient,
	}
}

func (u CustomerController) GetCustomers(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	var pagination request.Pagination

	if err := c.ShouldBindQuery(&pagination); err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	pagination.Validate()

	f := request.ExtractFilteredQueryParams(c, customers_DBModels.Customer{})

	customers, paginationResponse, err := u.CustomerDBClient.GetCustomers(ctx, pagination, f)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	for _, v := range customers {
		v.Password = ""
	}

	controller.RespondWithSuccessAndPagination(c, http.StatusOK, constants.GET_SUCCESSFULLY, customers, paginationResponse)
}

func (u CustomerController) GetCustomer(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	id := c.Param("id")
	filter := fmt.Sprintf("%s='%s'",
		customers_DBModels.COLUM_UUID, id,
	)

	r, err := u.CustomerDBClient.GetCustomer(ctx, filter)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if r.Uuid == uuid.Nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, "Customer not found", err)
		return
	}

	r.Password = ""

	fCIF := fmt.Sprintf("%s='%s'", cif_DBModels.COLUMN_CUSTOMER_UUID, r.Uuid)

	cif, err := u.CifDBClient.GetCustomerInformationFile(ctx, fCIF)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	customerData := struct {
		Customer customers_DBModels.Customer          `json:"customer"`
		CIF      cif_DBModels.CustomerInformationFile `json:"cif"`
	}{
		Customer: r,
		CIF:      cif,
	}

	controller.RespondWithSuccess(c, http.StatusOK, constants.GET_SUCCESSFULLY, customerData)
}

func (u CustomerController) UpdateCustomer(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	id := c.Param("id")
	filter := fmt.Sprintf("%s='%s'",
		customers_DBModels.COLUM_UUID, id,
	)

	r, err := u.CustomerDBClient.GetCustomer(ctx, filter)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if r.Uuid == uuid.Nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, "Customer not found", err)
		return
	}

	dataFromBody := customers_DBModels.Customer{}
	err = json.NewDecoder(c.Request.Body).Decode(&dataFromBody)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	var patcher = make(map[string]interface{})

	if dataFromBody.Name != "" {
		patcher[customers_DBModels.COLUMN_NAME] = dataFromBody.Name
	}

	if dataFromBody.Email != "" {
		patcher[customers_DBModels.COLUMN_EMAIL] = dataFromBody.Email
	}

	if dataFromBody.IsActive != nil {
		patcher[customers_DBModels.COLUMN_IS_ACTIVE] = dataFromBody.IsActive
	}

	if dataFromBody.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dataFromBody.Password), bcrypt.DefaultCost)
		if err != nil {
			errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
			log.Error(errorMsg)
			controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
			return
		}
		patcher[customers_DBModels.COLUMN_PASSWORD] = string(hashedPassword)
	}

	patcher[customers_DBModels.COLUMN_UPDATED_AT] = time.Now()

	filter = fmt.Sprintf("%s='%s'",
		customers_DBModels.COLUM_UUID, c.Param("id"),
	)

	if err := u.CustomerDBClient.UpdateCustomer(ctx, filter, patcher); err != nil {
		if constraintName := util.ExtractConstraintName(err.Error()); constraintName != "" {
			controller.RespondWithError(c, http.StatusConflict, fmt.Sprintf("%s already exists", constraintName), err)
			return
		}

		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	r, _ = u.CustomerDBClient.GetCustomer(ctx, fmt.Sprintf("%s='%s'",
		customers_DBModels.COLUM_UUID, c.Param("id"),
	))

	r.Password = ""

	controller.RespondWithSuccess(c, http.StatusAccepted, constants.UPDATED_SUCCESSFULLY, r)
}

func (u CustomerController) UpdateCustomerPassword(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	dataFromBody := customers_DBModels.Customer{}
	err := json.NewDecoder(c.Request.Body).Decode(&dataFromBody)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	id := c.Param("id")
	filter := fmt.Sprintf("%s='%s'",
		customers_DBModels.COLUM_UUID, id,
	)

	r, err := u.CustomerDBClient.GetCustomer(ctx, filter)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if r.Uuid == uuid.Nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, "Customer not found", err)
		return
	}

	var patcher = make(map[string]interface{})

	if dataFromBody.Password == "" {
		errorMsg := fmt.Sprintf("%s: %s", constants.BAD_REQUEST, "Password is required")
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	hashedPassword, err := util.GenerateHash(dataFromBody.Password) // Generate a hashed password
	if err != nil {
		log.Errorf(constants.INTERNAL_SERVER_ERROR, err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	patcher[customers_DBModels.COLUMN_PASSWORD] = hashedPassword
	patcher[customers_DBModels.COLUMN_UPDATED_AT] = time.Now()

	filter = fmt.Sprintf("%s='%s'",
		customers_DBModels.COLUM_UUID, c.Param("id"),
	)

	if err := u.CustomerDBClient.UpdateCustomer(ctx, filter, patcher); err != nil {
		if constraintName := util.ExtractConstraintName(err.Error()); constraintName != "" {
			controller.RespondWithError(c, http.StatusConflict, fmt.Sprintf("%s already exists", constraintName), err)
			return
		}

		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	r, _ = u.CustomerDBClient.GetCustomer(ctx, fmt.Sprintf("%s='%s'",
		customers_DBModels.COLUM_UUID, c.Param("id"),
	))

	r.Password = ""

	controller.RespondWithSuccess(c, http.StatusAccepted, constants.UPDATED_SUCCESSFULLY, r)
}

func (u CustomerController) DeleteCustomer(c *gin.Context) {
	ctx := correlation.WithReqContext(c)

	id := c.Param("id")

	filter := fmt.Sprintf("%s='%s'",
		customers_DBModels.COLUM_UUID, id,
	)

	r, err := u.CustomerDBClient.GetCustomer(ctx, filter)
	if err != nil {
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if r.Uuid == uuid.Nil {
		controller.RespondWithError(c, http.StatusInternalServerError, "Customer not found", err)
		return
	}

	if err := u.CustomerDBClient.DeleteCustomer(ctx, filter); err != nil {
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	controller.RespondWithSuccess(c, http.StatusOK, constants.DELETED_SUCCESSFULLY, nil)
}

func (u CustomerController) DeleteCustomers(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	// Parse the list of IDs from the request body or query parameters.
	var IDs []string
	if err := c.BindJSON(&IDs); err != nil {
		controller.RespondWithError(c, http.StatusBadRequest, constants.BAD_REQUEST, err)
		return
	}

	if len(IDs) == 0 {
		controller.RespondWithError(c, http.StatusBadRequest, constants.BAD_REQUEST, errors.New("empty ID list"))
		return
	}

	for _, id := range IDs {
		filter := fmt.Sprintf("%s='%s'", customers_DBModels.COLUM_UUID, id)

		if err := u.CustomerDBClient.DeleteCustomer(ctx, filter); err != nil {
			log.Errorf("Error deleting user with ID %d: %s", id, err.Error())
			controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
			return
		}
	}

	controller.RespondWithSuccess(c, http.StatusOK, constants.DELETED_SUCCESSFULLY, nil)
}

func (u CustomerController) GetCustomersDetail(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	var pagination request.Pagination

	if err := c.ShouldBindQuery(&pagination); err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	pagination.Validate()

	var customerDatas []struct {
		Customer customers_DBModels.Customer          `json:"customer"`
		CIF      cif_DBModels.CustomerInformationFile `json:"cif"`
	}

	f := request.ExtractFilteredQueryParams(c, customers_DBModels.Customer{})

	customers, paginationResponse, err := u.CustomerDBClient.GetCustomers(ctx, pagination, f)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	for _, v := range customers {
		v.Password = ""

		var customerData struct {
			Customer customers_DBModels.Customer          `json:"customer"`
			CIF      cif_DBModels.CustomerInformationFile `json:"cif"`
		}

		customerData.Customer = *v

		fCIF := fmt.Sprintf("%s='%s'", cif_DBModels.COLUMN_CUSTOMER_UUID, v.Uuid)

		cif, err := u.CifDBClient.GetCustomerInformationFile(ctx, fCIF)
		if err != nil {
			errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
			log.Error(errorMsg)
			controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
			return
		}

		customerData.CIF = cif

		customerDatas = append(customerDatas, customerData)
	}

	controller.RespondWithSuccessAndPagination(c, http.StatusOK, constants.GET_SUCCESSFULLY, customerDatas, paginationResponse)
}
