package transaction

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"sync"
	"user/sigmatech/app/constants"
	"user/sigmatech/app/controller"
	customerLimits_DBModels "user/sigmatech/app/db/dto/customer_limits"
	customers_DBModels "user/sigmatech/app/db/dto/customers"
	transaction_installments_DBModels "user/sigmatech/app/db/dto/transaction_installments"
	transactions_DBModels "user/sigmatech/app/db/dto/transactions"
	customerDB "user/sigmatech/app/db/repository/customer"
	customerLimitDB "user/sigmatech/app/db/repository/customer_limit"
	transactionDB "user/sigmatech/app/db/repository/transaction"
	transactionInstallmentDB "user/sigmatech/app/db/repository/transaction_installment"
	"user/sigmatech/app/service/correlation"
	"user/sigmatech/app/service/dto/request"
	"user/sigmatech/app/service/logger"

	"github.com/gin-gonic/gin"
)

// ITransactionController is an interface that defines the methods for a user controller.
type ITransactionController interface {
	GetTransactions(c *gin.Context)
	GetTransaction(c *gin.Context)
	GetTransactionDetails(c *gin.Context)
}

// TransactionController is a struct that implements the ITransactionController interface.
type TransactionController struct {
	CustomerDBClient               customerDB.ICustomerRepository // customerDB represents the database client for customer-related operations.
	CustomerLimitDBClient          customerLimitDB.ICustomerLimitRepository
	TransactionDBClient            transactionDB.ITransactionRepository
	transactionInstallmentDBClient transactionInstallmentDB.ITransactionInstallmentRepository
}

// NewTransactionController is a constructor function that creates a new TransactionController.
func NewTransactionController(
	CustomerDBClient customerDB.ICustomerRepository,
	CustomerLimitDBClient customerLimitDB.ICustomerLimitRepository,
	TransactionDBClient transactionDB.ITransactionRepository,
	transactionInstallmentDBClient transactionInstallmentDB.ITransactionInstallmentRepository,
) ITransactionController {
	return &TransactionController{
		CustomerDBClient:               CustomerDBClient,
		CustomerLimitDBClient:          CustomerLimitDBClient,
		TransactionDBClient:            TransactionDBClient,
		transactionInstallmentDBClient: transactionInstallmentDBClient,
	}
}

func (u TransactionController) GetTransactions(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger

	var pagination request.Pagination

	if err := c.ShouldBindQuery(&pagination); err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	pagination.Validate()

	f := request.ExtractFilteredQueryParams(c, transactions_DBModels.Transaction{})

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

	id := c.Param("id")
	filter := fmt.Sprintf("%s='%s'",
		transactions_DBModels.COLUM_UUID, id,
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

func (u TransactionController) GetTransactionDetails(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger

	var pagination request.Pagination

	if err := c.ShouldBindQuery(&pagination); err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	pagination.Validate()

	f := request.ExtractFilteredQueryParams(c, transactions_DBModels.Transaction{})

	transactions, paginationResponse, err := u.TransactionDBClient.GetTransactions(ctx, pagination, f)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	type TransactionData struct {
		Customer                customers_DBModels.Customer                                 `json:"customer"`
		Transaction             transactions_DBModels.Transaction                           `json:"transaction"`
		TransactionInstallments []*transaction_installments_DBModels.TransactionInstallment `json:"transaction_installments"`
	}

	response := make([]TransactionData, len(transactions))
	errCh := make(chan error, len(transactions))

	var wg sync.WaitGroup
	wg.Add(len(transactions))

	for i, transaction := range transactions {
		go func(index int, transaction transactions_DBModels.Transaction) {
			defer wg.Done()

			filter := fmt.Sprintf("%s='%s'",
				customers_DBModels.COLUM_UUID, transaction.CustomerUuid,
			)

			cust, err := u.CustomerDBClient.GetCustomer(ctx, filter)
			if err != nil {
				errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
				log.Error(errorMsg)
				controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
				return
			}

			p := request.Pagination{
				GetAllData: true,
				Order:      customerLimits_DBModels.COLUMN_TERM,
				Sort:       "ASC",
			}
			p.Validate()

			f := request.ExtractFilteredQueryParams(c, transaction_installments_DBModels.TransactionInstallment{})
			f[transaction_installments_DBModels.COLUMN_TRANSACTION_UUID] = transaction.Uuid.String()

			transactionInstallments, _, err := u.transactionInstallmentDBClient.GetTransactionInstallments(ctx, p, f)
			if err != nil {
				errCh <- fmt.Errorf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
				return
			}

			response[index] = TransactionData{
				Customer:                cust,
				Transaction:             transaction,
				TransactionInstallments: transactionInstallments,
			}

		}(i, *transaction)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			log.Error(err)
			controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
			return
		}
	}

	controller.RespondWithSuccessAndPagination(c, http.StatusOK, constants.GET_SUCCESSFULLY, response, paginationResponse)
}
