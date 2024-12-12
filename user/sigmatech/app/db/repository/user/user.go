package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"user/sigmatech/app/constants"
	db "user/sigmatech/app/db"
	users_DBModels "user/sigmatech/app/db/dto/users"
	"user/sigmatech/app/service/dto/request"
	"user/sigmatech/app/service/dto/response"
	"user/sigmatech/app/service/util"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type IUserRepository interface {
	CreateUser(ctx context.Context, user *users_DBModels.User) error
	GetUser(ctx context.Context, whr string) (users_DBModels.User, error)
	GetUsers(ctx context.Context, pagination request.Pagination, filter map[string]interface{}) ([]*users_DBModels.User, response.Pagination, error)
	UpdateUser(ctx context.Context, whr string, patch map[string]interface{}) error
	DeleteUser(ctx context.Context, filter string) error
}

type UserRepository struct {
	DBService *db.DBService
}

func NewUserRepository(dbService *db.DBService) IUserRepository {
	return &UserRepository{
		DBService: dbService,
	}
}

var tableName = users_DBModels.TABLE_NAME

func (u *UserRepository) CreateUser(ctx context.Context, user *users_DBModels.User) error {
	tx := u.DBService.GetDB().Begin()                       // Start a database transaction
	defer tx.Rollback()                                     // Rollback the transaction if not committed
	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode

	if err := tx.Table(users_DBModels.TABLE_NAME).Create(&user).Error; err != nil {
		return err // Return the error if user creation fails
	}
	tx.Commit() // Commit the transaction

	return nil // Return the created user and no error
}

func (u *UserRepository) GetUser(ctx context.Context, whr string) (users_DBModels.User, error) {
	tx := u.DBService.GetDB().Table(users_DBModels.TABLE_NAME) // Get the database instance and set table name
	var user users_DBModels.User                               // Variable to store the retrieved user

	if err := tx.Where(whr).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return users_DBModels.User{}, nil // Return an empty user if the record is not found
		}

		return user, err // Return the retrieved user and error, if any
	}

	return user, nil // Return the retrieved user and no error
}

func (u *UserRepository) GetUsers(ctx context.Context, paginationRequest request.Pagination, filter map[string]interface{}) (record []*users_DBModels.User, paginationResponse response.Pagination, err error) {
	tx := u.DBService.GetDB().Table(users_DBModels.TABLE_NAME)

	var columnsToSearch = []string{
		users_DBModels.COLUMN_NAME,
		users_DBModels.COLUMN_EMAIL,
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

func (u *UserRepository) UpdateUser(ctx context.Context, whr string, patch map[string]interface{}) error {
	tx := u.DBService.GetDB().Table(users_DBModels.TABLE_NAME).Begin() // Start a database transaction
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Rollback the transaction if a panic occurs
		}
	}()

	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode
	if err := tx.Where(whr).Updates(patch).Error; err != nil {
		tx.Rollback() // Rollback the transaction if user update fails
		return err
	}

	return tx.Commit().Error // Commit the transaction and return any error
}

func (u *UserRepository) DeleteUser(ctx context.Context, filter string) error {
	tx := u.DBService.GetDB().Table(users_DBModels.TABLE_NAME).Begin() // Start a database transaction
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Rollback the transaction if a panic occurs
		}
	}()

	tx.LogMode(constants.Config.DatabaseConfig.DB_LOG_MODE) // Set the database log mode
	if err := tx.Where(filter).Delete(&users_DBModels.User{}).Error; err != nil {
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
