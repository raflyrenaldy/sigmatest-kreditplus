package transaction

import (
	"customer/sigmatech/app/constants"
	"customer/sigmatech/app/controller"
	customerLimits_DBModels "customer/sigmatech/app/db/dto/customer_limits"
	customers_DBModels "customer/sigmatech/app/db/dto/customers"
	transaction_installments_DBModels "customer/sigmatech/app/db/dto/transaction_installments"
	transactions_DBModels "customer/sigmatech/app/db/dto/transactions"
	variableGlobals_DBModels "customer/sigmatech/app/db/dto/variable_globals"
	customerDB "customer/sigmatech/app/db/repository/customer"
	customerLimitDB "customer/sigmatech/app/db/repository/customer_limit"
	transactionDB "customer/sigmatech/app/db/repository/transaction"
	transactionInstallmentDB "customer/sigmatech/app/db/repository/transaction_installment"
	variableGlobalDB "customer/sigmatech/app/db/repository/variable_global"
	"customer/sigmatech/app/service/util"
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"time"

	"encoding/json"

	"customer/sigmatech/app/service/correlation"
	"customer/sigmatech/app/service/dto/request"
	"customer/sigmatech/app/service/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ITransactionController is an interface that defines the methods for a user controller.
type ITransactionController interface {
	GetTransactions(c *gin.Context)
	GetTransaction(c *gin.Context)
	CreateTransaction(c *gin.Context)
}

// TransactionController is a struct that implements the ITransactionController interface.
type TransactionController struct {
	CustomerDBClient               customerDB.ICustomerRepository // customerDB represents the database client for customer-related operations.
	CustomerLimitDBClient          customerLimitDB.ICustomerLimitRepository
	TransactionDBClient            transactionDB.ITransactionRepository
	transactionInstallmentDBClient transactionInstallmentDB.ITransactionInstallmentRepository
	variableGlobalDBClient         variableGlobalDB.IVariableGlobalRepository
}

// NewTransactionController is a constructor function that creates a new TransactionController.
func NewTransactionController(
	CustomerDBClient customerDB.ICustomerRepository,
	CustomerLimitDBClient customerLimitDB.ICustomerLimitRepository,
	TransactionDBClient transactionDB.ITransactionRepository,
	transactionInstallmentDBClient transactionInstallmentDB.ITransactionInstallmentRepository,
	variableGlobalDBClient variableGlobalDB.IVariableGlobalRepository,
) ITransactionController {
	return &TransactionController{
		CustomerDBClient:               CustomerDBClient,
		CustomerLimitDBClient:          CustomerLimitDBClient,
		TransactionDBClient:            TransactionDBClient,
		transactionInstallmentDBClient: transactionInstallmentDBClient,
		variableGlobalDBClient:         variableGlobalDBClient,
	}
}

func (u TransactionController) CreateTransaction(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	context, exist := c.Get(constants.CTK_CLAIM_KEY.String()) // Retrieve the user context from the request
	if !exist {
		errorMsg := fmt.Sprintf("%s: %s", constants.UNAUTHORIZED_ACCESS, constants.INVALID_TOKEN)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusUnauthorized, errorMsg, nil)
		return
	}
	usr := context.(*customers_DBModels.Customer) // Type assertion to retrieve the user information

	dataFromBody := transactions_DBModels.Transaction{}
	err := json.NewDecoder(c.Request.Body).Decode(&dataFromBody)
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

	fCustLimit := fmt.Sprintf("%s='%s' AND %s='%s'",
		customerLimits_DBModels.COLUM_UUID, dataFromBody.CustomerLimitUuid, customerLimits_DBModels.COLUMN_CUSTOMER_UUID, usr.Uuid,
	)

	customerLimit, err := u.CustomerLimitDBClient.GetCustomerLimit(ctx, fCustLimit)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if customerLimit.Uuid == uuid.Nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	fFeeAdmin := fmt.Sprintf("%s='%s'",
		variableGlobals_DBModels.COLUMN_CODE, constants.VARIABLE_ADMIN_FEE,
	)

	variableGlobalFeeAdmin, err := u.variableGlobalDBClient.GetVariableGlobal(ctx, fFeeAdmin)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if variableGlobalFeeAdmin.Uuid == uuid.Nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	fFeeInterest := fmt.Sprintf("%s='%s'",
		variableGlobals_DBModels.COLUMN_CODE, constants.VARIABLE_INTEREST_FEE,
	)

	variableGlobalFeeInterest, err := u.variableGlobalDBClient.GetVariableGlobal(ctx, fFeeInterest)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if variableGlobalFeeInterest.Uuid == uuid.Nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	admin, _ := strconv.ParseFloat(variableGlobalFeeAdmin.Value, 64)
	interest, _ := strconv.ParseFloat(variableGlobalFeeInterest.Value, 64)

	// Calculate total interest
	totalInterest := dataFromBody.Otr * interest / 100

	// Calculate total repayment (Loan amount + Interest + Admin fee)
	totalRepayment := dataFromBody.Otr + totalInterest + admin

	if customerLimit.RemainingLimit < totalRepayment {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, "limit tidak mencukupi", err)
		return
	}

	// Calculate monthly installment
	monthlyInstallment := totalRepayment / float64(customerLimit.Term)

	contractNumber, err := u.TransactionDBClient.GenerateContractNumber(ctx)
	if err != nil {
		contractNumber = fmt.Sprintf("TX_%06d_%v", 1, time.Now().Unix())
	}

	data := transactions_DBModels.Transaction{
		Uuid:              uuid.New(),
		CustomerUuid:      usr.Uuid,
		CustomerLimitUuid: customerLimit.Uuid,
		AssetName:         dataFromBody.AssetName,
		ContractNumber:    contractNumber,
		IsDone:            util.Boolean(false),
		Otr:               dataFromBody.Otr,
		AdminFee:          admin,
		Total:             totalRepayment,
		InstallmentAmount: monthlyInstallment,
		InstallmentCount:  customerLimit.Term,
		TotalInterest:     totalInterest,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err = u.TransactionDBClient.CreateTransaction(ctx, &data); err != nil {
		if constraintName := util.ExtractConstraintName(err.Error()); constraintName != "" {
			controller.RespondWithError(c, http.StatusConflict, fmt.Sprintf("%s already exists", constraintName), err)
			return
		}

		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}
	// Set the initial date
	currentDate := time.Now()

	for i := 1; i <= customerLimit.Term; i++ {
		nextMonth := currentDate.AddDate(0, i, 0)
		dataInstallment := transaction_installments_DBModels.TransactionInstallment{
			Uuid:            uuid.New(),
			TransactionUuid: data.Uuid,
			MethodPayment:   nil,
			Term:            i,
			DueDate:         &nextMonth,
			PaymentAt:       nil,
			Amount:          monthlyInstallment,
			AmountPaid:      0,
			CreatedAt:       currentDate,
			UpdatedAt:       currentDate,
		}

		if err = u.transactionInstallmentDBClient.CreateTransactionInstallment(ctx, &dataInstallment); err != nil {
			if constraintName := util.ExtractConstraintName(err.Error()); constraintName != "" {
				controller.RespondWithError(c, http.StatusConflict, fmt.Sprintf("%s already exists", constraintName), err)
				return
			}

			errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
			log.Error(errorMsg)
			controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
			return
		}

	}

	p := request.Pagination{
		GetAllData: true,
		Order:      customerLimits_DBModels.COLUMN_TERM,
		Sort:       "ASC",
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
	remainingLimit := 0.0

	for _, v := range customerLimits {
		scalingFactor := v.RemainingLimit / customerLimit.RemainingLimit

		remainingLimit = v.RemainingLimit - (totalRepayment * scalingFactor)

		if remainingLimit < 0 {
			remainingLimit = 0
		}

		var patcher = make(map[string]interface{}) // Create a patcher map to hold the fields to be updated

		patcher[customerLimits_DBModels.COLUMN_REMAINING_LIMIT] = remainingLimit

		fUpdLimit := fmt.Sprintf(`%s='%s'`, customerLimits_DBModels.COLUM_UUID, v.Uuid) // Create a filter string to match the customer ID

		if err := u.CustomerLimitDBClient.UpdateCustomerLimit(ctx, fUpdLimit, patcher); err != nil {
			if constraintName := util.ExtractConstraintName(err.Error()); constraintName != "" {
				controller.RespondWithError(c, http.StatusConflict, fmt.Sprintf("%s already exists", constraintName), err)
				return
			}

			errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
			log.Error(errorMsg)
			controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
			return
		}
	}
	controller.RespondWithSuccess(c, http.StatusOK, constants.CREATED_SUCCESSFULLY, data)
}

func (u TransactionController) GetTransactions(c *gin.Context) {
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

	var pagination request.Pagination

	if err := c.ShouldBindQuery(&pagination); err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	pagination.Validate()

	f := request.ExtractFilteredQueryParams(c, transactions_DBModels.Transaction{})
	f[transactions_DBModels.COLUMN_CUSTOMER_UUID] = usr.Uuid.String()

	transactions, paginationResponse, err := u.TransactionDBClient.GetTransactions(ctx, pagination, f)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	controller.RespondWithSuccessAndPagination(c, http.StatusOK, constants.GET_SUCCESSFULLY, transactions, paginationResponse)
}

func (u TransactionController) GetTransaction(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	context, exist := c.Get(constants.CTK_CLAIM_KEY.String()) // Retrieve the customer context from the request
	if !exist {
		errorMsg := fmt.Sprintf("%s: %s", constants.UNAUTHORIZED_ACCESS, constants.INVALID_TOKEN)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusUnauthorized, errorMsg, nil)
		return
	}
	usr := context.(*customers_DBModels.Customer) // Type assertion to retrieve the customer information

	id := c.Param("id")
	filter := fmt.Sprintf("%s='%s' AND %s='%s'",
		transactions_DBModels.COLUM_UUID, id, transactions_DBModels.COLUMN_CUSTOMER_UUID, usr.Uuid.String(),
	)

	r, err := u.TransactionDBClient.GetTransaction(ctx, filter)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if r.Uuid == uuid.Nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, "Transaction not found", err)
		return
	}

	var transactionData struct {
		Transaction             transactions_DBModels.Transaction                           `json:"transaction"`
		TransactionInstallments []*transaction_installments_DBModels.TransactionInstallment `json:"transaction_installments"`
	}

	p := request.Pagination{
		GetAllData: true,
		Order:      customerLimits_DBModels.COLUMN_TERM,
		Sort:       "ASC",
	}
	p.Validate()

	f := request.ExtractFilteredQueryParams(c, transaction_installments_DBModels.TransactionInstallment{})
	f[transaction_installments_DBModels.COLUMN_TRANSACTION_UUID] = r.Uuid.String()

	transactionInstallments, _, err := u.transactionInstallmentDBClient.GetTransactionInstallments(ctx, p, f)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	transactionData.Transaction = r
	transactionData.TransactionInstallments = transactionInstallments

	controller.RespondWithSuccess(c, http.StatusOK, constants.GET_SUCCESSFULLY, transactionData)
}
