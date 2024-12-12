package server

import (
	"context"
	"customer/sigmatech/app/api/middleware/auth"
	"customer/sigmatech/app/api/middleware/jwt"
	timeoutMiddleware "customer/sigmatech/app/api/middleware/timeout"
	"customer/sigmatech/app/constants"
	"customer/sigmatech/app/controller/healthcheck"
	"customer/sigmatech/app/db"

	awsS3 "customer/sigmatech/app/service/aws/s3"

	customerController "customer/sigmatech/app/controller/customers"
	customerDBClient "customer/sigmatech/app/db/repository/customer"

	cifDBClient "customer/sigmatech/app/db/repository/customer_information_file"

	customerLimitDBClient "customer/sigmatech/app/db/repository/customer_limit"

	"customer/sigmatech/app/service/logger"
	"strings"

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
		customerDBClient      = customerDBClient.NewCustomerRepository(dbConnection)
		cifDBClient           = cifDBClient.NewCustomerInformationFileRepository(dbConnection)
		customerLimitDBClient = customerLimitDBClient.NewCustomerLimitRepository(dbConnection)
	)

	// SERVICES
	var (
		jwt = jwt.NewJwtService(customerDBClient)
		s3  = awsS3.NewS3Service()
	)

	// Controller
	var (
		healthCheckController = healthcheck.NewHealthCheckController()
		customerController    = customerController.NewCustomerController(customerDBClient, cifDBClient, customerLimitDBClient, jwt, s3)
	)

	// API version v1
	v1 := router.Group("/v1")
	{
		// Health Check
		v1.GET(HEALTH_CHECK, healthCheckController.HealthCheck)

		// Customer routes
		customer := v1.Group(CUSTOMER)
		{
			// Public customer sign-up and sign-in routes
			v1.POST(CUSTOMER+SIGN_UP+"/", customerController.SignUp)
			v1.POST(CUSTOMER+SIGN_IN+"/", customerController.SignIn)
			v1.POST(CUSTOMER+REFRESH_TOKEN+"/", customerController.RefreshToken)

			// User profile routes
			customer.Use(auth.Authentication(jwt)) // pass allowed roles for the APIs
			customer.GET(PROFILE+"/", customerController.GetProfile)
			customer.PATCH(PROFILE+"/", customerController.UpdateProfile)
			customer.PATCH(PROFILE_PASSWORD+"/", customerController.UpdateProfilePassword)

			// Limit routes
			limit := customer.Group(LIMIT)
			{
				limit.GET("/", customerController.GetLimits)
			}

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
