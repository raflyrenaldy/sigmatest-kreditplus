package customer_information_file

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"user/sigmatech/app/constants"
	db "user/sigmatech/app/db"
	customerInformationFiles_DBModels "user/sigmatech/app/db/dto/customer_information_files"
	"user/sigmatech/app/service/dto/request"
	"user/sigmatech/app/service/dto/response"
	"user/sigmatech/app/service/util"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type ICustomerInformationFileRepository interface {
	CreateCustomerInformationFile(ctx context.Context, customer *customerInformationFiles_DBModels.CustomerInformationFile) error
	GetCustomerInformationFile(ctx context.Context, whr string) (customerInformationFiles_DBModels.CustomerInformationFile, error)
	GetCustomerInformationFiles(ctx context.Context, pagination request.Pagination, filter map[string]interface{}) ([]*customerInformationFiles_DBModels.CustomerInformationFile, response.Pagination, error)
	UpdateCustomerInformationFile(ctx context.Context, whr string, patch map[string]interface{}) error
	DeleteCustomerInformationFile(ctx context.Context, filter string) error
	GenerateCIFNumber(ctx context.Context) (string, error)
}

type CustomerInformationFileRepository struct {
	DBService *db.DBService
}

func NewCustomerInformationFileRepository(dbService *db.DBService) ICustomerInformationFileRepository {
	return &CustomerInformationFileRepository{
		DBService: dbService,
	}
}

var tableName = customerInformationFiles_DBModels.TABLE_NAME

func (u *CustomerInformationFileRepository) CreateCustomerInformationFile(ctx context.Context, customer *customerInformationFiles_DBModels.CustomerInformationFile) error {
	tx := u.DBService.GetDB().Begin()                       // Start a database transaction
	defer tx.Rollback()                                     // Rollback the transaction if not committed
	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode

	if err := tx.Table(customerInformationFiles_DBModels.TABLE_NAME).Create(&customer).Error; err != nil {
		return err // Return the error if customer creation fails
	}
	tx.Commit() // Commit the transaction

	return nil // Return the created customer and no error
}

func (u *CustomerInformationFileRepository) GetCustomerInformationFile(ctx context.Context, whr string) (customerInformationFiles_DBModels.CustomerInformationFile, error) {
	tx := u.DBService.GetDB().Table(customerInformationFiles_DBModels.TABLE_NAME) // Get the database instance and set table name
	var customer customerInformationFiles_DBModels.CustomerInformationFile        // Variable to store the retrieved customer

	if err := tx.Where(whr).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerInformationFiles_DBModels.CustomerInformationFile{}, nil // Return an empty customer if the record is not found
		}

		return customer, err // Return the retrieved customer and error, if any
	}

	return customer, nil // Return the retrieved customer and no error
}

func (u *CustomerInformationFileRepository) GetCustomerInformationFiles(ctx context.Context, paginationRequest request.Pagination, filter map[string]interface{}) (record []*customerInformationFiles_DBModels.CustomerInformationFile, paginationResponse response.Pagination, err error) {
	tx := u.DBService.GetDB().Table(customerInformationFiles_DBModels.TABLE_NAME)

	if filter["outlet_id"] != nil {
		tx = tx.Where("EXISTS (SELECT 1 FROM customer_outlets WHERE customers.id = customer_outlets.customer_id AND customer_outlets.outlet_id = ?)", filter["outlet_id"])
		delete(filter, "outlet_id") // Remove the filter after using it
	}

	var columnsToSearch = []string{
		customerInformationFiles_DBModels.COLUMN_CIF_NUMBER,
		customerInformationFiles_DBModels.COLUMN_FULL_NAME,
		customerInformationFiles_DBModels.COLUMN_LEGAL_NAME,
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

func (u *CustomerInformationFileRepository) UpdateCustomerInformationFile(ctx context.Context, whr string, patch map[string]interface{}) error {
	tx := u.DBService.GetDB().Table(customerInformationFiles_DBModels.TABLE_NAME).Begin() // Start a database transaction
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

func (u *CustomerInformationFileRepository) DeleteCustomerInformationFile(ctx context.Context, filter string) error {
	tx := u.DBService.GetDB().Table(customerInformationFiles_DBModels.TABLE_NAME).Begin() // Start a database transaction
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Rollback the transaction if a panic occurs
		}
	}()

	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode
	if err := tx.Where(filter).Delete(&customerInformationFiles_DBModels.CustomerInformationFile{}).Error; err != nil {
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

func (u *CustomerInformationFileRepository) GenerateCIFNumber(ctx context.Context) (string, error) {
	tx := u.DBService.GetDB().Table(customerInformationFiles_DBModels.TABLE_NAME)

	latestData := fmt.Sprintf(" date(%s)='%s'", customerInformationFiles_DBModels.COLUMN_CREATED_AT, time.Now().Format("2006-01-02"))

	var record customerInformationFiles_DBModels.CustomerInformationFile
	if err := tx.Where(latestData).Order("id DESC").First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Sprintf("CIF_%06d_%v", 1, time.Now().Unix()), nil
		}
		return "", err
	}

	// Extract the numeric part from the last transaction ID (sequence number)
	var num int
	var timestamp int64
	_, err := fmt.Sscanf(record.CifNumber, "CIF_%06d_%d", &num, &timestamp) // Capture both the sequence number and timestamp
	if err != nil {
		return "", fmt.Errorf("failed to parse Cif Number: %v", err)
	}
	// Increment the transaction number
	num++
	// Generate the new transaction ID with the incremented number and the current timestamp
	newContractNumber := fmt.Sprintf("CIF_%06d_%v", num, time.Now().Unix())
	return newContractNumber, nil
}
