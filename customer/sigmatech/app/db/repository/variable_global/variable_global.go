package variable_global

import (
	"context"
	"customer/sigmatech/app/constants"
	db "customer/sigmatech/app/db"
	variableGlobals_DBModels "customer/sigmatech/app/db/dto/variable_globals"
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

type IVariableGlobalRepository interface {
	CreateVariableGlobal(ctx context.Context, customer *variableGlobals_DBModels.VariableGlobal) error
	GetVariableGlobal(ctx context.Context, whr string) (variableGlobals_DBModels.VariableGlobal, error)
	GetVariableGlobals(ctx context.Context, pagination request.Pagination, filter map[string]interface{}) ([]*variableGlobals_DBModels.VariableGlobal, response.Pagination, error)
	UpdateVariableGlobal(ctx context.Context, whr string, patch map[string]interface{}) error
	DeleteVariableGlobal(ctx context.Context, filter string) error
}

type VariableGlobalRepository struct {
	DBService *db.DBService
}

func NewVariableGlobalRepository(dbService *db.DBService) IVariableGlobalRepository {
	return &VariableGlobalRepository{
		DBService: dbService,
	}
}

var tableName = variableGlobals_DBModels.TABLE_NAME

func (u *VariableGlobalRepository) CreateVariableGlobal(ctx context.Context, customer *variableGlobals_DBModels.VariableGlobal) error {
	tx := u.DBService.GetDB().Begin()                       // Start a database variableGlobal
	defer tx.Rollback()                                     // Rollback the variableGlobal if not committed
	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode

	if err := tx.Table(variableGlobals_DBModels.TABLE_NAME).Create(&customer).Error; err != nil {
		return err // Return the error if customer creation fails
	}
	tx.Commit() // Commit the variableGlobal

	return nil // Return the created customer and no error
}

func (u *VariableGlobalRepository) GetVariableGlobal(ctx context.Context, whr string) (variableGlobals_DBModels.VariableGlobal, error) {
	tx := u.DBService.GetDB().Table(variableGlobals_DBModels.TABLE_NAME) // Get the database instance and set table name
	var customer variableGlobals_DBModels.VariableGlobal                 // Variable to store the retrieved customer

	if err := tx.Where(whr).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return variableGlobals_DBModels.VariableGlobal{}, nil // Return an empty customer if the record is not found
		}

		return customer, err // Return the retrieved customer and error, if any
	}

	return customer, nil // Return the retrieved customer and no error
}

func (u *VariableGlobalRepository) GetVariableGlobals(ctx context.Context, paginationRequest request.Pagination, filter map[string]interface{}) (record []*variableGlobals_DBModels.VariableGlobal, paginationResponse response.Pagination, err error) {
	tx := u.DBService.GetDB().Table(variableGlobals_DBModels.TABLE_NAME)

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

func (u *VariableGlobalRepository) UpdateVariableGlobal(ctx context.Context, whr string, patch map[string]interface{}) error {
	tx := u.DBService.GetDB().Table(variableGlobals_DBModels.TABLE_NAME).Begin() // Start a database variableGlobal
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Rollback the variableGlobal if a panic occurs
		}
	}()

	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode
	if err := tx.Where(whr).Updates(patch).Error; err != nil {
		tx.Rollback() // Rollback the variableGlobal if customer update fails
		return err
	}

	return tx.Commit().Error // Commit the variableGlobal and return any error
}

func (u *VariableGlobalRepository) DeleteVariableGlobal(ctx context.Context, filter string) error {
	tx := u.DBService.GetDB().Table(variableGlobals_DBModels.TABLE_NAME).Begin() // Start a database variableGlobal
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Rollback the variableGlobal if a panic occurs
		}
	}()

	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode
	if err := tx.Where(filter).Delete(&variableGlobals_DBModels.VariableGlobal{}).Error; err != nil {
		tx.Rollback() // Rollback the variableGlobal if the deletion fails
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23503" {
			// Foreign key constraint violation detected, rollback the variableGlobal
			tx.Rollback()

			// Start a new variableGlobal for the schema query
			schemaTx := u.DBService.GetDB().Begin()
			defer func() {
				if r := recover(); r != nil {
					schemaTx.Rollback() // Rollback the schema variableGlobal if a panic occurs
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

			// Rollback the schema variableGlobal
			schemaTx.Rollback()

			// Return the foreign key constraint violation error
			return fmt.Errorf("cannot delete from table %s because of foreign key constraints on tables %s", tableName, strings.Join(referencingTables, ", "))
		}
		return err
	}

	return tx.Commit().Error // Commit the variableGlobal and return any error
}
