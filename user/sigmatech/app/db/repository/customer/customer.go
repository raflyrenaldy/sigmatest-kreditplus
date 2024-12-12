package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"user/sigmatech/app/constants"
	db "user/sigmatech/app/db"
	customers_DBModels "user/sigmatech/app/db/dto/customers"
	"user/sigmatech/app/service/dto/request"
	"user/sigmatech/app/service/dto/response"
	"user/sigmatech/app/service/util"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type ICustomerRepository interface {
	CreateCustomer(ctx context.Context, customer *customers_DBModels.Customer) error
	GetCustomer(ctx context.Context, whr string) (customers_DBModels.Customer, error)
	GetCustomers(ctx context.Context, pagination request.Pagination, filter map[string]interface{}) ([]*customers_DBModels.Customer, response.Pagination, error)
	UpdateCustomer(ctx context.Context, whr string, patch map[string]interface{}) error
	DeleteCustomer(ctx context.Context, filter string) error
}

type CustomerRepository struct {
	DBService *db.DBService
}

func NewCustomerRepository(dbService *db.DBService) ICustomerRepository {
	return &CustomerRepository{
		DBService: dbService,
	}
}

var tableName = customers_DBModels.TABLE_NAME

func (u *CustomerRepository) CreateCustomer(ctx context.Context, customer *customers_DBModels.Customer) error {
	tx := u.DBService.GetDB().Begin()                       // Start a database transaction
	defer tx.Rollback()                                     // Rollback the transaction if not committed
	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode

	if err := tx.Table(customers_DBModels.TABLE_NAME).Create(&customer).Error; err != nil {
		return err // Return the error if customer creation fails
	}
	tx.Commit() // Commit the transaction

	return nil // Return the created customer and no error
}

func (u *CustomerRepository) GetCustomer(ctx context.Context, whr string) (customers_DBModels.Customer, error) {
	tx := u.DBService.GetDB().Table(customers_DBModels.TABLE_NAME) // Get the database instance and set table name
	var customer customers_DBModels.Customer                       // Variable to store the retrieved customer

	if err := tx.Where(whr).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customers_DBModels.Customer{}, nil // Return an empty customer if the record is not found
		}

		return customer, err // Return the retrieved customer and error, if any
	}

	return customer, nil // Return the retrieved customer and no error
}

func (u *CustomerRepository) GetCustomers(ctx context.Context, paginationRequest request.Pagination, filter map[string]interface{}) (record []*customers_DBModels.Customer, paginationResponse response.Pagination, err error) {
	tx := u.DBService.GetDB().Table(customers_DBModels.TABLE_NAME)

	var columnsToSearch = []string{
		customers_DBModels.COLUMN_NAME,
		customers_DBModels.COLUMN_EMAIL,
	}

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

func (u *CustomerRepository) UpdateCustomer(ctx context.Context, whr string, patch map[string]interface{}) error {
	tx := u.DBService.GetDB().Table(customers_DBModels.TABLE_NAME).Begin() // Start a database transaction
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Rollback the transaction if a panic occurs
		}
	}()

	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode
	if err := tx.Where(whr).Updates(patch).Error; err != nil {
		tx.Rollback() // Rollback the transaction if customer update fails
		return err
	}

	return tx.Commit().Error // Commit the transaction and return any error
}

func (u *CustomerRepository) DeleteCustomer(ctx context.Context, filter string) error {
	tx := u.DBService.GetDB().Table(customers_DBModels.TABLE_NAME).Begin() // Start a database transaction
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Rollback the transaction if a panic occurs
		}
	}()

	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode
	if err := tx.Where(filter).Delete(&customers_DBModels.Customer{}).Error; err != nil {
		tx.Rollback() // Rollback the transaction if the deletion fails
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23503" {
			// Foreign key constraint violation detected, rollback the transaction
			tx.Rollback()

			// Start a new transaction for the schema query
			schemaTx := u.DBService.GetDB().Begin()
			defer func() {
				if r := recover(); r != nil {
					schemaTx.Rollback() // Rollback the schema transaction if a panic occurs
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

			// Rollback the schema transaction
			schemaTx.Rollback()

			// Return the foreign key constraint violation error
			return fmt.Errorf("cannot delete from table %s because of foreign key constraints on tables %s", tableName, strings.Join(referencingTables, ", "))
		}
		return err
	}

	return tx.Commit().Error // Commit the transaction and return any error
}
