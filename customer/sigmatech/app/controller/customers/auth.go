package customers

import (
	cif_DBModels "customer/sigmatech/app/db/dto/customer_information_files"
	customerLimits_DBModels "customer/sigmatech/app/db/dto/customer_limits"
	reqCustomer "customer/sigmatech/app/service/dto/request/customer"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"strings"
	"time"

	"customer/sigmatech/app/constants"
	"customer/sigmatech/app/controller"
	customers_DBModels "customer/sigmatech/app/db/dto/customers"

	"customer/sigmatech/app/service/correlation"
	"customer/sigmatech/app/service/logger"
	"customer/sigmatech/app/service/util"

	"github.com/gin-gonic/gin"
)

// SignUp handles the sign-up functionality
func (u CustomerController) SignUp(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger
	var dataFromBody reqCustomer.SignUpReq

	email := c.PostForm("email")
	fullName := c.PostForm("full_name")
	legalName := c.PostForm("legal_name")
	placeOfBirth := c.PostForm("place_of_birth")
	dateOfBirth := c.PostForm("date_of_birth")
	gender := c.PostForm("gender")
	password := c.PostForm("password")
	nik := c.PostForm("nik")
	salary, _ := strconv.ParseFloat(c.PostForm("salary"), 64)

	dataFromBody.Email = email
	dataFromBody.LegalName = legalName
	dataFromBody.FullName = fullName
	dataFromBody.PlaceOfBirth = placeOfBirth
	dataFromBody.Gender = gender
	dataFromBody.Nik = nik
	dataFromBody.Name = fullName
	dataFromBody.Password = password
	dataFromBody.Salary = salary

	if dateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", dateOfBirth)
		if err != nil {
			errorMsg := fmt.Sprintf("Error parsing date: %v", err)
			log.Error(errorMsg)
			controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
			return
		}

		// Assign the parsed date to the pointer
		dataFromBody.DateOfBirth = &dob
	}

	// Get the multipart form
	form, err := c.MultipartForm()
	if err != nil {
		controller.RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err), err)
		return
	}

	// Get the files from the form
	files := form.File["card_photo"]
	if len(files) != 0 {
		// Get the first file
		file := files[0]

		// Check if the file is an image
		if !util.IsImage(file) {
			controller.RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("%s: %s", constants.BAD_REQUEST, "File is not an image"), nil)
			return
		}

		// Check if the file size is not more than 2MB
		if file.Size > 2<<20 {
			controller.RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("%s: %s", constants.BAD_REQUEST, "File size cannot be more than 2MB"), nil)
			return
		}

		// Open the file for reading
		fileReader, err := file.Open()
		if err != nil {
			controller.RespondWithError(c, http.StatusInternalServerError, fmt.Sprintf("%s: %s", constants.INTERNAL_SERVER_ERROR, err), err)
			return
		}
		defer fileReader.Close()

		fileBytes, err := util.CompressImage(fileReader, file.Size, 70)

		if err != nil {
			controller.RespondWithError(c, http.StatusInternalServerError, fmt.Sprintf("%s: %s", constants.INTERNAL_SERVER_ERROR, err), err)
			return
		}
		// Generate a unique file name
		// This is to ensure that the file name is unique
		// The file name is also used as the key in the S3 bucket
		fileName := fmt.Sprintf("%s%s.jpg", constants.CUSTOMER_CARD_PHOTO, uuid.New().String())

		// Save the file to the S3 bucket
		if _, err = u.S3Client.PutObject(fileName, fileBytes); err != nil {
			controller.RespondWithError(c, http.StatusInternalServerError, fmt.Sprintf("%s: %s", constants.INTERNAL_SERVER_ERROR, err), err)
			return
		}
		dataFromBody.CardPhoto = fileName
	}

	// Get the files from the form
	selfieFiles := form.File["selfie_photo"]
	if len(selfieFiles) != 0 {
		// Get the first file
		file := selfieFiles[0]

		// Check if the file is an image
		if !util.IsImage(file) {
			controller.RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("%s: %s", constants.BAD_REQUEST, "File is not an image"), nil)
			return
		}

		// Check if the file size is not more than 2MB
		if file.Size > 2<<20 {
			controller.RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("%s: %s", constants.BAD_REQUEST, "File size cannot be more than 2MB"), nil)
			return
		}

		// Open the file for reading
		fileReader, err := file.Open()
		if err != nil {
			controller.RespondWithError(c, http.StatusInternalServerError, fmt.Sprintf("%s: %s", constants.INTERNAL_SERVER_ERROR, err), err)
			return
		}
		defer fileReader.Close()

		fileBytes, err := util.CompressImage(fileReader, file.Size, 70)

		if err != nil {
			controller.RespondWithError(c, http.StatusInternalServerError, fmt.Sprintf("%s: %s", constants.INTERNAL_SERVER_ERROR, err), err)
			return
		}
		// Generate a unique file name
		// This is to ensure that the file name is unique
		// The file name is also used as the key in the S3 bucket
		fileName := fmt.Sprintf("%s%s.jpg", constants.CUSTOMER_SELFIE_PHOTO, uuid.New().String())

		// Save the file to the S3 bucket
		if _, err = u.S3Client.PutObject(fileName, fileBytes); err != nil {
			controller.RespondWithError(c, http.StatusInternalServerError, fmt.Sprintf("%s: %s", constants.INTERNAL_SERVER_ERROR, err), err)
			return
		}
		dataFromBody.SelfiePhoto = fileName
	}

	if err := dataFromBody.Validate(); err != nil {
		controller.RespondWithError(c, http.StatusBadRequest, constants.BAD_REQUEST, err)
		return
	}

	fCheck := fmt.Sprintf("%s='%s'", cif_DBModels.COLUMN_NIK, dataFromBody.Nik)
	check, err := u.CIFDBClient.GetCustomerInformationFile(ctx, fCheck)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if check.Uuid != uuid.Nil {
		if _, err := u.S3Client.DeleteObject(dataFromBody.CardPhoto); err != nil {
			log.Errorf("Error deleting card photo: %s", err.Error())
		}

		if _, err := u.S3Client.DeleteObject(dataFromBody.SelfiePhoto); err != nil {
			log.Errorf("Error deleting selfie photo: %s", err.Error())
		}

		controller.RespondWithError(c, http.StatusBadRequest, "NIK already registered", err)
		return
	}

	now := time.Now()

	customerData := customers_DBModels.Customer{
		Uuid:      uuid.New(),
		Name:      dataFromBody.Name,
		Email:     dataFromBody.Email,
		Password:  dataFromBody.Password,
		IsActive:  false, // Default user registered is false.
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.CustomerDBClient.CreateCustomer(ctx, &customerData); err != nil {
		if constraintName := util.ExtractConstraintName(err.Error()); constraintName != "" {
			controller.RespondWithError(c, http.StatusConflict, fmt.Sprintf("%s already exists", constraintName), err)
			return
		}

		if _, err := u.S3Client.DeleteObject(dataFromBody.CardPhoto); err != nil {
			log.Errorf("Error deleting card photo: %s", err.Error())
		}

		if _, err := u.S3Client.DeleteObject(dataFromBody.SelfiePhoto); err != nil {
			log.Errorf("Error deleting selfie photo: %s", err.Error())
		}

		log.Errorf(constants.INTERNAL_SERVER_ERROR, err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	cifNumber, err := u.CIFDBClient.GenerateCIFNumber(ctx)
	if err != nil {
		cifNumber = fmt.Sprintf("CF_%06d_%v", 1, time.Now().Unix())
	}

	cifData := cif_DBModels.CustomerInformationFile{
		Uuid:         uuid.New(),
		CustomerUuid: customerData.Uuid,
		CifNumber:    cifNumber,
		Nik:          dataFromBody.Nik,
		FullName:     dataFromBody.FullName,
		LegalName:    dataFromBody.LegalName,
		PlaceOfBirth: dataFromBody.PlaceOfBirth,
		DateOfBirth:  dataFromBody.DateOfBirth,
		Gender:       &dataFromBody.Gender,
		Salary:       dataFromBody.Salary,
		CardPhoto:    dataFromBody.CardPhoto,
		SelfiePhoto:  dataFromBody.SelfiePhoto,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := u.CIFDBClient.CreateCustomerInformationFile(ctx, &cifData); err != nil {
		if constraintName := util.ExtractConstraintName(err.Error()); constraintName != "" {
			controller.RespondWithError(c, http.StatusConflict, fmt.Sprintf("%s already exists", constraintName), err)
			return
		}

		if _, err := u.S3Client.DeleteObject(dataFromBody.CardPhoto); err != nil {
			log.Errorf("Error deleting card photo: %s", err.Error())
		}

		if _, err := u.S3Client.DeleteObject(dataFromBody.SelfiePhoto); err != nil {
			log.Errorf("Error deleting selfie photo: %s", err.Error())
		}

		log.Errorf(constants.INTERNAL_SERVER_ERROR, err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	var term = []int{1, 2, 3, 6}

	for _, v := range term {
		customerLimitData := customerLimits_DBModels.CustomerLimit{
			Uuid:           uuid.New(),
			CustomerUuid:   customerData.Uuid,
			Term:           v,
			Status:         util.Boolean(false),
			AmountLimit:    0,
			RemainingLimit: 0,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if err := u.CustomerLimitDBClient.CreateCustomerLimit(ctx, &customerLimitData); err != nil {
			if constraintName := util.ExtractConstraintName(err.Error()); constraintName != "" {
				controller.RespondWithError(c, http.StatusConflict, fmt.Sprintf("%s already exists", constraintName), err)
				return
			}

			if _, err := u.S3Client.DeleteObject(dataFromBody.CardPhoto); err != nil {
				log.Errorf("Error deleting card photo: %s", err.Error())
			}

			if _, err := u.S3Client.DeleteObject(dataFromBody.SelfiePhoto); err != nil {
				log.Errorf("Error deleting selfie photo: %s", err.Error())
			}

			log.Errorf(constants.INTERNAL_SERVER_ERROR, err)
			controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
			return
		}
	}

	// Create the vendor profile struct
	customerProfile := struct {
		Customer customers_DBModels.Customer          `json:"customer"`
		CIF      cif_DBModels.CustomerInformationFile `json:"cif"`
	}{
		Customer: customerData,
		CIF:      cifData,
	}

	controller.RespondWithSuccess(c, http.StatusAccepted, "Customer Created Successfully", customerProfile)
}

// SignIn handles the login functionality
func (u CustomerController) SignIn(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	var dataFromBody reqCustomer.SignInRequest
	if err := c.ShouldBindJSON(&dataFromBody); err != nil {
		log.Errorf(constants.BAD_REQUEST, err)
		controller.RespondWithError(c, http.StatusBadRequest, constants.BAD_REQUEST, err)
		return
	}

	if err := dataFromBody.Validate(); err != nil {
		controller.RespondWithError(c, http.StatusBadRequest, constants.BAD_REQUEST, err)
		return
	}

	filter := fmt.Sprintf("%s='%s'", customers_DBModels.COLUMN_EMAIL, dataFromBody.Email)

	customer, err := u.CustomerDBClient.GetCustomer(ctx, filter)
	if err != nil {
		log.Errorf("Error getting customer: %v", err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if !util.ValidatePassword(dataFromBody.Password, customer.Password) {
		log.Errorf("Wrong credentials")
		controller.RespondWithError(c, http.StatusUnauthorized, "Wrong credentials", errors.New(constants.UNAUTHORIZED_ACCESS))
		return
	}

	if !customer.IsActive {
		controller.RespondWithError(c, http.StatusUnauthorized, "Harap tunggu untuk konfirmasi Admin", errors.New(constants.UNAUTHORIZED_ACCESS))
		return
	}

	token, err := u.JWT.GenerateCustomerTokens(ctx, customer)
	if err != nil {
		log.Errorf("Error while creating access token: %v", err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	controller.RespondWithSuccess(c, http.StatusAccepted, "Login Successfully", token)
}

// RefreshToken handles the refresh token functionality
func (u CustomerController) RefreshToken(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger

	tokenString := strings.TrimPrefix(c.GetHeader(constants.AUTHORIZATION), constants.BEARER)

	claims, err := u.JWT.RefreshCustomerToken(ctx, tokenString)
	if err != nil {
		log.Errorf("Error while verifying refresh token: %v", err)
		controller.RespondWithError(c, http.StatusBadRequest, constants.INVALID_TOKEN, err)
		return
	}

	controller.RespondWithSuccess(c, http.StatusAccepted, "Refresh Token Successfully", claims)
}
