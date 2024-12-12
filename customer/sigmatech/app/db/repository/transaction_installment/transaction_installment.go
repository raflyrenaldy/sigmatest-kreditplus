package transaction_installment_installment

import (
	"context"
	"customer/sigmatech/app/constants"
	db "customer/sigmatech/app/db"
	transaction_installments_DBModels "customer/sigmatech/app/db/dto/transaction_installments"
	"customer/sigmatech/app/service/dto/request"
	"customer/sigmatech/app/service/dto/response"
	"customer/sigmatech/app/service/util"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type ITransactionInstallmentRepository interface {
	CreateTransactionInstallment(ctx context.Context, customer *transaction_installments_DBModels.TransactionInstallment) error
	GetTransactionInstallment(ctx context.Context, whr string) (transaction_installments_DBModels.TransactionInstallment, error)
	GetTransactionInstallments(ctx context.Context, pagination request.Pagination, filter map[string]interface{}) ([]*transaction_installments_DBModels.TransactionInstallment, response.Pagination, error)
	UpdateTransactionInstallment(ctx context.Context, whr string, patch map[string]interface{}) error
	DeleteTransactionInstallment(ctx context.Context, filter string) error
}

type TransactionInstallmentRepository struct {
	DBService *db.DBService
}

func NewTransactionInstallmentRepository(dbService *db.DBService) ITransactionInstallmentRepository {
	return &TransactionInstallmentRepository{
		DBService: dbService,
	}
}

var tableName = transaction_installments_DBModels.TABLE_NAME

func (u *TransactionInstallmentRepository) CreateTransactionInstallment(ctx context.Context, customer *transaction_installments_DBModels.TransactionInstallment) error {
	tx := u.DBService.GetDB().Begin()                       // Start a database transaction_installment
	defer tx.Rollback()                                     // Rollback the transaction_installment if not committed
	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode

	if err := tx.Table(transaction_installments_DBModels.TABLE_NAME).Create(&customer).Error; err != nil {
		return err // Return the error if customer creation fails
	}
	tx.Commit() // Commit the transaction_installment

	return nil // Return the created customer and no error
}

func (u *TransactionInstallmentRepository) GetTransactionInstallment(ctx context.Context, whr string) (transaction_installments_DBModels.TransactionInstallment, error) {
	tx := u.DBService.GetDB().Table(transaction_installments_DBModels.TABLE_NAME) // Get the database instance and set table name
	var customer transaction_installments_DBModels.TransactionInstallment         // Variable to store the retrieved customer

	if err := tx.Where(whr).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return transaction_installments_DBModels.TransactionInstallment{}, nil // Return an empty customer if the record is not found
		}

		return customer, err // Return the retrieved customer and error, if any
	}

	return customer, nil // Return the retrieved customer and no error
}

func (u *TransactionInstallmentRepository) GetTransactionInstallments(ctx context.Context, paginationRequest request.Pagination, filter map[string]interface{}) (record []*transaction_installments_DBModels.TransactionInstallment, paginationResponse response.Pagination, err error) {
	tx := u.DBService.GetDB().Table(transaction_installments_DBModels.TABLE_NAME)

	var columnsToSearch = []string{}

	var whr string
	if paginationRequest.Query != "" {
		var orConditions []string
		for _, column := range columnsToSearch {
			orConditions = append(orConditions, fmt.Sprintf("coalesce(%s)", column))
		}
		whr = fmt.Sprintf("(%s) ILIKE '%%%s%%'", strings.Join(orConditions, " || "), paginationRequest.Query)
	}

	query := tx.Where(whr)

	// Iterate through the filter map and add the TABLE_NAME prefix to filter parameters that don't have a "." prefix
	for key, value := range filter {
		if !strings.Contains(key, ".") {
			filter[tableName+"."+key] = value
			delete(filter, key)
		}
	}

	query, err = util.ApplyFilterCondition(query, filter)
	if err != nil {
		return nil, response.Pagination{}, err
	}

	var totalCount int
	if err := query.Count(&totalCount).Error; err == sql.ErrNoRows {
		return nil, response.Pagination{}, nil
	}

	query = query.Limit(*paginationRequest.Limit).Offset((*paginationRequest.Page - 1) * *paginationRequest.Limit)

	if totalCount == 0 || *paginationRequest.Page > ((totalCount+*paginationRequest.Limit-1) / *paginationRequest.Limit) {
		return nil, response.Pagination{}, nil
	}
	paginationResponse.TotalCount = totalCount
	paginationResponse.TotalPages = (totalCount + *paginationRequest.Limit - 1) / *paginationRequest.Limit
	paginationResponse.Page = *paginationRequest.Page
	paginationResponse.PerPage = *paginationRequest.Limit

	query = query.Order(fmt.Sprintf("%s %s", paginationRequest.Order, paginationRequest.Sort))

	if err := query.Find(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.Pagination{}, nil
		}
		return record, paginationResponse, err
	}

	return record, paginationResponse, nil
}

func (u *TransactionInstallmentRepository) UpdateTransactionInstallment(ctx context.Context, whr string, patch map[string]interface{}) error {
	tx := u.DBService.GetDB().Table(transaction_installments_DBModels.TABLE_NAME).Begin() // Start a database transaction_installment
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Rollback the transaction_installment if a panic occurs
		}
	}()

	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode
	if err := tx.Where(whr).Updates(patch).Error; err != nil {
		tx.Rollback() // Rollback the transaction_installment if customer update fails
		return err
	}

	return tx.Commit().Error // Commit the transaction_installment and return any error
}

func (u *TransactionInstallmentRepository) DeleteTransactionInstallment(ctx context.Context, filter string) error {
	tx := u.DBService.GetDB().Table(transaction_installments_DBModels.TABLE_NAME).Begin() // Start a database transaction_installment
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Rollback the transaction_installment if a panic occurs
		}
	}()

	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode
	if err := tx.Where(filter).Delete(&transaction_installments_DBModels.TransactionInstallment{}).Error; err != nil {
		tx.Rollback() // Rollback the transaction_installment if the deletion fails
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23503" {
			// Foreign key constraint violation detected, rollback the transaction_installment
			tx.Rollback()

			// Start a new transaction_installment for the schema query
			schemaTx := u.DBService.GetDB().Begin()
			defer func() {
				if r := recover(); r != nil {
					schemaTx.Rollback() // Rollback the schema transaction_installment if a panic occurs
				}
			}()

			// Query the database schema to get tables with foreign key constraints referencing the specified table
			var foreignKeys []struct {
				ConstraintName   string
				ReferencingTable string
			}
			err := schemaTx.Raw(`
			SELECT con.conname AS constraint_name, con.conrelid::regclass AS referencing_table
			FROM pg_constraint con
			WHERE con.contype = 'f'
			AND con.confrelid = ?::regclass`, tableName).Scan(&foreignKeys).Error

			if err != nil {
				return err
			}

			var referencingTables []string
			// Print or process the foreign key information
			for _, fk := range foreignKeys {
				referencingTables = append(referencingTables, fk.ReferencingTable)
			}

			// Rollback the schema transaction_installment
			schemaTx.Rollback()

			// Return the foreign key constraint violation error
			return fmt.Errorf("cannot delete from table %s because of foreign key constraints on tables %s", tableName, strings.Join(referencingTables, ", "))
		}
		return err
	}

	return tx.Commit().Error // Commit the transaction_installment and return any error
}
