package db

import (
	"context"
	"customer/sigmatech/app/constants"
	"customer/sigmatech/app/service/logger"
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"

	"bitbucket.org/liamstask/goose/lib/goose"
)

// Init : Initializes the database migrations
func Init(ctx context.Context) (db *gorm.DB, err error) {
	log := logger.Logger(ctx)

	// Get database configuration parameters from constants
	dbUserName := constants.Config.DatabaseConfig.DB_USER
	dbPassword := constants.Config.DatabaseConfig.DB_PASSWORD
	dbHost := constants.Config.DatabaseConfig.DB_HOST
	dbName := constants.Config.DatabaseConfig.DB_NAME
	dbPort := constants.Config.DatabaseConfig.DB_PORT
	dbSchema := constants.Config.DatabaseConfig.DB_SCHEMA

	// Get additional database connection configuration parameters from constants
	maxIdleConnections := constants.Config.DatabaseConfig.DB_MAX_IDLE_CONNECTION
	maxOpenConnections := constants.Config.DatabaseConfig.DB_MAX_OPEN_CONNECTION
	connectionMaxLifetime := constants.Config.DatabaseConfig.DB_CONNECTION_MAX_LIFETIME

	// Construct the database URI
	dbURI := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUserName, dbPassword, dbName)

	log.Info("Connecting to DB", dbURI)

	// Open a connection to the database
	db, err = gorm.Open("postgres", dbURI)
	if err != nil {
		log.Fatalf("Failed to connect to DB", dbURI, err.Error())
		os.Exit(1)
	}

	// Set the SQL dialect to Postgres
	dialect := &goose.PostgresDialect{}

	// Set the maximum number of idle connections
	db.DB().SetMaxIdleConns(maxIdleConnections)

	// Set the maximum number of open connections
	db.DB().SetMaxOpenConns(maxOpenConnections)

	// Set the maximum lifetime of a connection
	db.DB().SetConnMaxLifetime(time.Minute * time.Duration(connectionMaxLifetime))

	// Enable singular table name
	db.SingularTable(true)

	// Check if a database schema needs to be created
	if dbSchema != "" {
		sch := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", dbSchema)
		db.Exec(sch)
	}

	// Fetch the working directory for migrations
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Not able to fetch the working directory")
		os.Exit(1)
	}

	migrationsDir := workingDir + "/sigmatech/app/db/migrations"

	// Check if the migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		log.Fatalf("Migrations directory does not exist")
		os.Exit(1)
	}

	err = os.Chdir(migrationsDir)
	if err != nil {
		log.Fatalf("Not able to change the working directory")
		os.Exit(1)
	}

	migrateConf := &goose.DBConf{
		MigrationsDir: workingDir,
		Driver: goose.DBDriver{
			Name:    "postgres",
			OpenStr: dbURI,
			Import:  "github.com/lib/pq",
			Dialect: dialect,
		},
	}

	// Get the most recent database version
	log.Info("Fetching the most recent DB version")
	latest, err := goose.GetMostRecentDBVersion(migrateConf.MigrationsDir)
	if err != nil {
		log.Errorf("Unable to get recent goose db version", err)
	}

	log.Info(" Most recent DB version ", latest)

	// Run the database migrations
	log.Info("Running the migrations on db", workingDir)
	err = goose.RunMigrationsOnDb(migrateConf, migrateConf.MigrationsDir, latest, db.DB())
	if err != nil {
		log.Fatalf("Error while running migrations", err)
		os.Exit(1)
	}

	return
}

func New(dbConn *gorm.DB) *DBService {
	return &DBService{
		DB: dbConn,
	}
}

type DBService struct {
	DB *gorm.DB
}

// GetDB : Get an instance of DB to connect to the database connection pool
func (d DBService) GetDB() *gorm.DB {
	return d.DB
	// return d.DB.Debug()
}
