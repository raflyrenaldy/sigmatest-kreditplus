package users

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
	"user/sigmatech/app/service/dto/request/user"

	"user/sigmatech/app/constants"
	"user/sigmatech/app/controller"
	users_DBModels "user/sigmatech/app/db/dto/users"

	"user/sigmatech/app/service/correlation"
	"user/sigmatech/app/service/logger"
	"user/sigmatech/app/service/util"

	"github.com/gin-gonic/gin"
)

// SignUp handles the sign-up functionality
func (u UserController) SignUp(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger

	var dataFromBody users_DBModels.User

	fCheck := fmt.Sprintf("%s='%s'", users_DBModels.COLUMN_NAME, "super-sigmatech")
	check, err := u.UserDBClient.GetUser(ctx, fCheck)
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if check.Uuid == uuid.Nil {
		controller.RespondWithError(c, http.StatusBadRequest, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	now := time.Now()
	dataFromBody.Uuid = uuid.New()
	dataFromBody.Name = "super-sigmatech"
	dataFromBody.Email = "super@sigmatech.id"
	dataFromBody.Password = "##password##sigmatech##"
	dataFromBody.CreatedAt = now
	dataFromBody.UpdatedAt = now

	if err := dataFromBody.ValidateUser(); err != nil {
		controller.RespondWithError(c, http.StatusBadRequest, constants.BAD_REQUEST, err)
		return
	}

	if err := u.UserDBClient.CreateUser(ctx, &dataFromBody); err != nil {
		if constraintName := util.ExtractConstraintName(err.Error()); constraintName != "" {
			controller.RespondWithError(c, http.StatusConflict, fmt.Sprintf("%s already exists", constraintName), err)
			return
		}

		log.Errorf(constants.INTERNAL_SERVER_ERROR, err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	controller.RespondWithSuccess(c, http.StatusAccepted, "User Created Successfully", dataFromBody)
}

// SignIn handles the login functionality
func (u UserController) SignIn(c *gin.Context) {
	ctx := correlation.WithReqContext(c)
	log := logger.Logger(ctx)

	var dataFromBody user.SignInRequest
	if err := c.ShouldBindJSON(&dataFromBody); err != nil {
		log.Errorf(constants.BAD_REQUEST, err)
		controller.RespondWithError(c, http.StatusBadRequest, constants.BAD_REQUEST, err)
		return
	}

	if err := dataFromBody.Validate(); err != nil {
		controller.RespondWithError(c, http.StatusBadRequest, constants.BAD_REQUEST, err)
		return
	}

	filter := fmt.Sprintf("%s='%s'", users_DBModels.COLUMN_EMAIL, dataFromBody.Email)

	user, err := u.UserDBClient.GetUser(ctx, filter)
	if err != nil {
		log.Errorf("Error getting user: %v", err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	if !util.ValidatePassword(dataFromBody.Password, user.Password) {
		log.Errorf("Wrong credentials")
		controller.RespondWithError(c, http.StatusUnauthorized, "Wrong credentials", errors.New(constants.UNAUTHORIZED_ACCESS))
		return
	}

	token, err := u.JWT.GenerateUserTokens(ctx, user)
	if err != nil {
		log.Errorf("Error while creating access token: %v", err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	controller.RespondWithSuccess(c, http.StatusAccepted, "Login Successfully", token)
}

// RefreshToken handles the refresh token functionality
func (u UserController) RefreshToken(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger

	tokenString := strings.TrimPrefix(c.GetHeader(constants.AUTHORIZATION), constants.BEARER)

	claims, err := u.JWT.RefreshUserToken(ctx, tokenString)
	if err != nil {
		log.Errorf("Error while verifying refresh token: %v", err)
		controller.RespondWithError(c, http.StatusBadRequest, constants.INVALID_TOKEN, err)
		return
	}

	controller.RespondWithSuccess(c, http.StatusAccepted, "Refresh Token Successfully", claims)
}
