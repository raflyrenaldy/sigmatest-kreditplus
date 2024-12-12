package server

import (
	"context"
	"user/sigmatech/app/api/middleware/auth"
	"user/sigmatech/app/api/middleware/jwt"
	timeoutMiddleware "user/sigmatech/app/api/middleware/timeout"
	"user/sigmatech/app/constants"
	"user/sigmatech/app/controller/healthcheck"
	transactionController "user/sigmatech/app/controller/transaction"
	userController "user/sigmatech/app/controller/users"
	"user/sigmatech/app/db"
	transactionDBClient "user/sigmatech/app/db/repository/transaction"
	transactionInstallmentDBClient "user/sigmatech/app/db/repository/transaction_installment"
	userDBClient "user/sigmatech/app/db/repository/user"

	customerDBClient "user/sigmatech/app/db/repository/customer"
	cifDBClient "user/sigmatech/app/db/repository/customer_information_file"
	customerLimitDBClient "user/sigmatech/app/db/repository/customer_limit"

	customerController "user/sigmatech/app/controller/customer"

	"strings"
	"user/sigmatech/app/service/logger"

	helmet "github.com/danielkov/gin-helmet"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Init(ctx context.Context, dbConnection *db.DBService) *gin.Engine {
	if strings.EqualFold(constants.Config.Environment, string(constants.Production)) {
		gin.SetMode(gin.ReleaseMode)
	}
	return NewRouter(ctx, dbConnection)

}
func addCSPHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

func addReferrerPolicyHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

func addPermissionsPolicyHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Permissions-Policy", "default-src 'none'")
		c.Next()
	}
}

func addFeaturePolicyHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Feature-Policy", "none")
		c.Next()
	}
}

func NewRouter(ctx context.Context, dbConnection *db.DBService) *gin.Engine {
	log := logger.Logger(ctx)

	log.Info("setting up service and controllers")

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(helmet.Default())
	//Content-Security-Policy
	router.Use(addCSPHeader())
	//Referrer-Policy
	router.Use(addReferrerPolicyHeader())
	//Permissions-Policy
	router.Use(addPermissionsPolicyHeader())
	//Feature-Policy
	router.Use(addFeaturePolicyHeader())

	// router.Use(corsMiddleware() )

	// Enable CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PATCH", "DELETE", "PUT", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Accept", "Content-Type", constants.AUTHORIZATION, constants.CORRELATION_KEY_ID.String()}
	router.Use(cors.New(config))

	router.Use(uuidInjectionMiddleware())
	router.Use(timeoutMiddleware.TimeoutMiddleware())

	// DB Clients
	var (
		userDBClient          = userDBClient.NewUserRepository(dbConnection)
		customerDBClient      = customerDBClient.NewCustomerRepository(dbConnection)
		customerLimitDBClient = customerLimitDBClient.NewCustomerLimitRepository(dbConnection)
		cifDBClient           = cifDBClient.NewCustomerInformationFileRepository(dbConnection)

		transactionDBClient            = transactionDBClient.NewTransactionRepository(dbConnection)
		transactionInstallmentDBClient = transactionInstallmentDBClient.NewTransactionInstallmentRepository(dbConnection)
	)

	// SERVICES
	var (
		jwt = jwt.NewJwtService(userDBClient)
	)

	// Controller
	var (
		healthCheckController = healthcheck.NewHealthCheckController()
		userController        = userController.NewUserController(userDBClient, jwt)
		customerController    = customerController.NewCustomerController(customerDBClient, customerLimitDBClient, cifDBClient)

		transactionController = transactionController.NewTransactionController(customerDBClient, customerLimitDBClient, transactionDBClient, transactionInstallmentDBClient)
	)

	// API version v1
	v1 := router.Group("/v1")
	{
		// Health Check
		v1.GET(HEALTH_CHECK, healthCheckController.HealthCheck)

		// User routes
		user := v1.Group(USER)
		{
			// Public user sign-up and sign-in routes
			v1.POST(USER+SIGN_UP+"/", userController.SignUp)
			v1.POST(USER+SIGN_IN+"/", userController.SignIn)
			v1.POST(USER+REFRESH_TOKEN+"/", userController.RefreshToken)

			// User profile routes
			user.Use(auth.Authentication(jwt)) // pass allowed roles for the APIs
			user.GET(PROFILE+"/", userController.GetProfile)
			user.PATCH(PROFILE+"/", userController.UpdateProfile)
			user.PATCH(PROFILE_PASSWORD+"/", userController.UpdateProfilePassword)

			// User CRUD routes
			user.POST("/", userController.CreateUser)
			user.GET("/", userController.GetUsers)
			user.GET("/:id/", userController.GetUser)
			user.PATCH("/:id/", userController.UpdateUser)
			user.PATCH("/:id/"+PASSWORD+"/", userController.UpdateUserPassword)
			user.DELETE("/:id/", userController.DeleteUser)
			user.DELETE("/", userController.DeleteUsers)
		}

		// Customer routes
		customer := v1.Group(CUSTOMER)
		{
			// Customer route
			customer.Use(auth.Authentication(jwt)) // pass allowed roles for the APIs
			customer.GET("/", customerController.GetCustomers)
			customer.GET("/"+DETAIL, customerController.GetCustomersDetail)
			customer.GET("/:id/", customerController.GetCustomer)
			customer.PATCH("/:id/", customerController.UpdateCustomer)
			customer.PATCH("/:id/"+PASSWORD+"/", customerController.UpdateCustomerPassword)
			customer.DELETE("/:id/", customerController.DeleteCustomer)
			customer.DELETE("/", customerController.DeleteCustomers)

			// Customer routes
			customerLimit := customer.Group(LIMIT)
			{
				// Customer Limit route
				customerLimit.GET("/:id/", customerController.GetCustomerLimits)
				customerLimit.PATCH(APPROVE+"/", customerController.ApproveCustomer)
			}
		}

		// Transaction routes
		transaction := v1.Group(TRANSACTION)
		{
			transaction.Use(auth.Authentication(jwt)) // pass allowed roles for the APIs
			transaction.GET("/", transactionController.GetTransactions)
			transaction.GET(DETAIL+"/", transactionController.GetTransactionDetails)
			transaction.GET("/:id/", transactionController.GetTransaction)
		}

	}

	return router
}

// uuidInjectionMiddleware injects the request context with a correlation id of type uuid
func uuidInjectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationId := c.GetHeader(string(constants.CORRELATION_KEY_ID))
		if len(correlationId) == 0 {
			correlationID, _ := uuid.NewUUID()
			correlationId = correlationID.String()
			c.Request.Header.Set(constants.CORRELATION_KEY_ID.String(), correlationId)
		}
		c.Writer.Header().Set(constants.CORRELATION_KEY_ID.String(), correlationId)

		c.Next()
	}
}
