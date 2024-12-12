package customers

import (
	"customer/sigmatech/app/api/middleware/jwt"
	customerDB "customer/sigmatech/app/db/repository/customer"
	cifDB "customer/sigmatech/app/db/repository/customer_information_file"
	customerLimitDB "customer/sigmatech/app/db/repository/customer_limit"
	"customer/sigmatech/app/service/aws/s3"

	"github.com/gin-gonic/gin"
)

// ICustomerController is an interface that defines the methods for a customer controller.
type ICustomerController interface {
	SignUp(c *gin.Context)
	SignIn(c *gin.Context)
	RefreshToken(c *gin.Context)

	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	UpdateProfilePassword(c *gin.Context)

	GetLimits(c *gin.Context)
}

// CustomerController is a struct that implements the ICustomerController interface.
type CustomerController struct {
	CustomerDBClient      customerDB.ICustomerRepository // customerDB represents the database client for crm-customer-related operations.
	CIFDBClient           cifDB.ICustomerInformationFileRepository
	CustomerLimitDBClient customerLimitDB.ICustomerLimitRepository

	JWT jwt.IJwtService

	S3Client s3.IS3Client // S3Client represents the AWS S3 client for file storage.
}

// NewCustomerController is a constructor function that creates a new CustomerController.
func NewCustomerController(
	CustomerDBClient customerDB.ICustomerRepository,
	CIFDBClient cifDB.ICustomerInformationFileRepository,
	CustomerLimitDBClient customerLimitDB.ICustomerLimitRepository,
	jwt jwt.IJwtService,
	S3Client s3.IS3Client,
) ICustomerController {
	return &CustomerController{
		CustomerDBClient:      CustomerDBClient,
		CIFDBClient:           CIFDBClient,
		CustomerLimitDBClient: CustomerLimitDBClient,
		JWT:                   jwt,
		S3Client:              S3Client,
	}
}
