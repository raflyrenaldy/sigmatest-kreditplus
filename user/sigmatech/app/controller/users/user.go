package users

import (
	"errors"
	"github.com/google/uuid"
	"user/sigmatech/app/api/middleware/jwt"
	"user/sigmatech/app/constants"
	"user/sigmatech/app/controller"

	"time"
	users_DBModels "user/sigmatech/app/db/dto/users"
	userDB "user/sigmatech/app/db/repository/user"

	"encoding/json"
	"fmt"

	"net/http"
	"user/sigmatech/app/service/correlation"
	"user/sigmatech/app/service/dto/request"
	reqUser "user/sigmatech/app/service/dto/request/user"
	"user/sigmatech/app/service/logger"
	"user/sigmatech/app/service/util"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// IUserController is an interface that defines the methods for a user controller.
type IUserController interface {
	SignUp(c *gin.Context)
	SignIn(c *gin.Context)
	RefreshToken(c *gin.Context)

	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	UpdateProfilePassword(c *gin.Context)

	CreateUser(c *gin.Context)
	GetUsers(c *gin.Context)
	GetUser(c *gin.Context)
	UpdateUser(c *gin.Context)
	UpdateUserPassword(c *gin.Context)
	DeleteUser(c *gin.Context)
	DeleteUsers(c *gin.Context)
}

// UserController is a struct that implements the IUserController interface.
type UserController struct {
	UserDBClient userDB.IUserRepository // userDB represents the database client for crm-user-related operations.

	JWT jwt.IJwtService
}

// NewUserController is a constructor function that creates a new UserController.
func NewUserController(
	UserDBClient userDB.IUserRepository,
	jwt jwt.IJwtService,
) IUserController {
	return &UserController{
		UserDBClient: UserDBClient,
		JWT:          jwt,
	}
}

func (u UserController) CreateUser(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	context, exist := c.Get(constants.CTK_CLAIM_KEY.String()) // Retrieve the user context from the request
	if !exist {
		errorMsg := fmt.Sprintf("%s: %s", constants.UNAUTHORIZED_ACCESS, constants.INVALID_TOKEN)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusUnauthorized, errorMsg, nil)
		return
	}
	usr := context.(*users_DBModels.User) // Type assertion to retrieve the user information

	dataFromBody := reqUser.CreateUserReq{}
	err := json.NewDecoder(c.Request.Body).Decode(&dataFromBody)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}
	data := users_DBModels.User{
		Uuid:      uuid.New(),
		Name:      dataFromBody.Name,
		Email:     dataFromBody.Email,
		Password:  dataFromBody.Password,
		CreatedAt: time.Now(),
		CreatedBy: &usr.Uuid,
		UpdatedAt: time.Now(),
		UpdatedBy: nil,
	}

	if err := data.ValidateUser(); err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	fCheck := fmt.Sprintf("%s='%s'", users_DBModels.COLUMN_EMAIL, data.Email)
	check, err := u.UserDBClient.GetUser(ctx, fCheck)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}
	if check.Uuid != uuid.Nil {
		fmt.Println("test")
		controller.RespondWithError(c, http.StatusBadRequest, "Email already exists", err)
		return
	}

	if err = u.UserDBClient.CreateUser(ctx, &data); err != nil {
		if constraintName := util.ExtractConstraintName(err.Error()); constraintName != "" {
			controller.RespondWithError(c, http.StatusConflict, fmt.Sprintf("%s already exists", constraintName), err)
			return
		}

		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	controller.RespondWithSuccess(c, http.StatusOK, constants.CREATED_SUCCESSFULLY, data)
}

func (u UserController) GetUsers(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	var pagination request.Pagination

	if err := c.ShouldBindQuery(&pagination); err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	pagination.Validate()

	f := request.ExtractFilteredQueryParams(c, users_DBModels.User{})

	users, paginationResponse, err := u.UserDBClient.GetUsers(ctx, pagination, f)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	for _, v := range users {
		v.Password = ""
	}

	controller.RespondWithSuccessAndPagination(c, http.StatusOK, constants.GET_SUCCESSFULLY, users, paginationResponse)
}

func (u UserController) GetUser(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	id := c.Param("id")
	filter := fmt.Sprintf("%s='%s'",
		users_DBModels.COLUM_UUID, id,
	)

	r, err := u.UserDBClient.GetUser(ctx, filter)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if r.Uuid == uuid.Nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, "User not found", err)
		return
	}

	r.Password = ""

	controller.RespondWithSuccess(c, http.StatusOK, constants.GET_SUCCESSFULLY, r)
}

func (u UserController) UpdateUser(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	id := c.Param("id")
	filter := fmt.Sprintf("%s='%s'",
		users_DBModels.COLUM_UUID, id,
	)

	r, err := u.UserDBClient.GetUser(ctx, filter)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if r.Uuid == uuid.Nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, "User not found", err)
		return
	}

	dataFromBody := reqUser.CreateUserReq{}
	err = json.NewDecoder(c.Request.Body).Decode(&dataFromBody)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	var patcher = make(map[string]interface{})

	if dataFromBody.Name != "" {
		patcher[users_DBModels.COLUMN_NAME] = dataFromBody.Name
	}
	if dataFromBody.Email != "" {
		patcher[users_DBModels.COLUMN_EMAIL] = dataFromBody.Email
	}

	if dataFromBody.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dataFromBody.Password), bcrypt.DefaultCost)
		if err != nil {
			errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
			log.Error(errorMsg)
			controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
			return
		}
		patcher[users_DBModels.COLUMN_PASSWORD] = string(hashedPassword)
	}

	patcher[users_DBModels.COLUMN_UPDATED_AT] = time.Now()

	filter = fmt.Sprintf("%s='%s'",
		users_DBModels.COLUM_UUID, c.Param("id"),
	)

	if err := u.UserDBClient.UpdateUser(ctx, filter, patcher); err != nil {
		if constraintName := util.ExtractConstraintName(err.Error()); constraintName != "" {
			controller.RespondWithError(c, http.StatusConflict, fmt.Sprintf("%s already exists", constraintName), err)
			return
		}

		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	r, _ = u.UserDBClient.GetUser(ctx, fmt.Sprintf("%s='%s'",
		users_DBModels.COLUM_UUID, c.Param("id"),
	))

	r.Password = ""

	controller.RespondWithSuccess(c, http.StatusAccepted, constants.UPDATED_SUCCESSFULLY, r)
}

func (u UserController) UpdateUserPassword(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	context, exist := c.Get(constants.CTK_CLAIM_KEY.String()) // Retrieve the user context from the request
	if !exist {
		errorMsg := fmt.Sprintf("%s: %s", constants.UNAUTHORIZED_ACCESS, constants.INVALID_TOKEN)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusUnauthorized, errorMsg, nil)
		return
	}
	usr := context.(*users_DBModels.User) // Type assertion to retrieve the user information

	dataFromBody := users_DBModels.User{}
	err := json.NewDecoder(c.Request.Body).Decode(&dataFromBody)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	id := c.Param("id")
	filter := fmt.Sprintf("%s='%s'",
		users_DBModels.COLUM_UUID, id,
	)

	r, err := u.UserDBClient.GetUser(ctx, filter)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if r.Uuid == uuid.Nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, "User not found", err)
		return
	}

	var patcher = make(map[string]interface{})

	if dataFromBody.Password == "" {
		errorMsg := fmt.Sprintf("%s: %s", constants.BAD_REQUEST, "Password is required")
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	hashedPassword, err := util.GenerateHash(dataFromBody.Password) // Generate a hashed password
	if err != nil {
		log.Errorf(constants.INTERNAL_SERVER_ERROR, err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	patcher[users_DBModels.COLUMN_PASSWORD] = hashedPassword
	patcher[users_DBModels.COLUMN_UPDATED_AT] = time.Now()
	patcher[users_DBModels.COLUMN_UPDATED_BY] = usr.Uuid

	filter = fmt.Sprintf("%s='%s'",
		users_DBModels.COLUM_UUID, c.Param("id"),
	)

	if err := u.UserDBClient.UpdateUser(ctx, filter, patcher); err != nil {
		if constraintName := util.ExtractConstraintName(err.Error()); constraintName != "" {
			controller.RespondWithError(c, http.StatusConflict, fmt.Sprintf("%s already exists", constraintName), err)
			return
		}

		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	r, _ = u.UserDBClient.GetUser(ctx, fmt.Sprintf("%s='%s'",
		users_DBModels.COLUM_UUID, c.Param("id"),
	))

	r.Password = ""

	controller.RespondWithSuccess(c, http.StatusAccepted, constants.UPDATED_SUCCESSFULLY, r)
}

func (u UserController) DeleteUser(c *gin.Context) {
	ctx := correlation.WithReqContext(c)

	id := c.Param("id")

	filter := fmt.Sprintf("%s='%s'",
		users_DBModels.COLUM_UUID, id,
	)

	r, err := u.UserDBClient.GetUser(ctx, filter)
	if err != nil {
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if r.Uuid == uuid.Nil {
		controller.RespondWithError(c, http.StatusInternalServerError, "User not found", err)
		return
	}

	if err := u.UserDBClient.DeleteUser(ctx, filter); err != nil {
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	controller.RespondWithSuccess(c, http.StatusOK, constants.DELETED_SUCCESSFULLY, nil)
}

func (u UserController) DeleteUsers(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	// Parse the list of IDs from the request body or query parameters.
	var IDs []string
	if err := c.BindJSON(&IDs); err != nil {
		controller.RespondWithError(c, http.StatusBadRequest, constants.BAD_REQUEST, err)
		return
	}

	if len(IDs) == 0 {
		controller.RespondWithError(c, http.StatusBadRequest, constants.BAD_REQUEST, errors.New("empty ID list"))
		return
	}

	for _, id := range IDs {
		filter := fmt.Sprintf("%s='%s'", users_DBModels.COLUM_UUID, id)

		if err := u.UserDBClient.DeleteUser(ctx, filter); err != nil {
			log.Errorf("Error deleting user with ID %d: %s", id, err.Error())
			controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
			return
		}
	}

	controller.RespondWithSuccess(c, http.StatusOK, constants.DELETED_SUCCESSFULLY, nil)
}
