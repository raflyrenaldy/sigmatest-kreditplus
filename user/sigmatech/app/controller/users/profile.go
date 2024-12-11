package users

import (
	"encoding/json"
	"fmt"
	"net/http"
	"user/sigmatech/app/constants"
	"user/sigmatech/app/controller"
	users_DBModels "user/sigmatech/app/db/dto/users"

	"user/sigmatech/app/service/correlation"
	"user/sigmatech/app/service/dto/request"
	"user/sigmatech/app/service/logger"
	"user/sigmatech/app/service/util"

	reqUser "user/sigmatech/app/service/dto/request/user"

	"github.com/gin-gonic/gin"
)

func (u UserController) GetProfile(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger

	context, exist := c.Get(constants.CTK_CLAIM_KEY.String()) // Retrieve the user context from the request
	if !exist {
		errorMsg := fmt.Sprintf("%s: %s", constants.UNAUTHORIZED_ACCESS, constants.INVALID_TOKEN)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusUnauthorized, errorMsg, nil)
		return
	}
	usr := context.(*users_DBModels.User) // Type assertion to retrieve the user information

	user, err := u.UserDBClient.GetUser(ctx, fmt.Sprintf("%s='%s'",
		users_DBModels.COLUMN_EMAIL, usr.Email,
	))
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}
	user.Password = "" // Clear the password field for security reasons

	p := request.Pagination{
		GetAllData: true,
	}
	p.Validate()

	// Create the vendor profile struct
	userProfile := struct {
		User users_DBModels.User `json:"user"`
	}{
		User: user,
	}

	// Respond with success message and the user profile (with password field cleared)
	controller.RespondWithSuccess(c, http.StatusOK, constants.GET_SUCCESSFULLY, userProfile)
}

func (u UserController) UpdateProfile(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger

	context, exist := c.Get(constants.CTK_CLAIM_KEY.String()) // Retrieve the user context from the request
	if !exist {
		errorMsg := fmt.Sprintf("%s: %s", constants.UNAUTHORIZED_ACCESS, constants.INVALID_TOKEN)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusUnauthorized, errorMsg, nil)
		return
	}
	usr := context.(*users_DBModels.User) // Type assertion to retrieve the user information

	dataFromBody := users_DBModels.User{}                        // Create an empty Users struct
	err := json.NewDecoder(c.Request.Body).Decode(&dataFromBody) // Decode the request body into the Users struct
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	var patcher = make(map[string]interface{}) // Create a patcher map to hold the fields to be updated

	// Check if each field is present in the request body and add it to the patcher map if not empty
	if dataFromBody.Name != "" {
		patcher[users_DBModels.COLUMN_NAME] = dataFromBody.Name
	}

	if dataFromBody.Email != "" {
		patcher[users_DBModels.COLUMN_EMAIL] = dataFromBody.Email
	}

	filter := fmt.Sprintf(`%s='%s'`, users_DBModels.COLUMN_EMAIL, usr.Email) // Create a filter string to match the user ID

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

	r, _ := u.UserDBClient.GetUser(ctx, fmt.Sprintf("%s='%s'",
		users_DBModels.COLUMN_EMAIL, usr.Email,
	))
	r.Password = ""

	// Respond with success message and the updated user profile (with password field cleared)
	controller.RespondWithSuccess(c, http.StatusAccepted, constants.UPDATED_SUCCESSFULLY, r)
}

func (u UserController) UpdateProfilePassword(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger

	context, exist := c.Get(constants.CTK_CLAIM_KEY.String()) // Retrieve the user context from the request
	if !exist {
		errorMsg := fmt.Sprintf("%s: %s", constants.UNAUTHORIZED_ACCESS, constants.INVALID_TOKEN)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusUnauthorized, errorMsg, nil)
		return
	}
	usr := context.(*users_DBModels.User) // Type assertion to retrieve the user information

	dataFromBody := reqUser.UpdatePassword{}                     // Create an empty Users struct
	err := json.NewDecoder(c.Request.Body).Decode(&dataFromBody) // Decode the request body into the Users struct
	if err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	if err := dataFromBody.Validate(); err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.BAD_REQUEST, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusBadRequest, errorMsg, err)
		return
	}

	if !util.ValidatePassword(dataFromBody.OldPassword, usr.Password) {
		log.Errorf("Wrong credentials")
		controller.RespondWithError(c, http.StatusUnauthorized, "Password lama salah.", nil)
		return
	}

	hashedPassword, err := util.GenerateHash(dataFromBody.Password) // Generate a hashed password
	if err != nil {
		log.Errorf(constants.INTERNAL_SERVER_ERROR, err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	filter := fmt.Sprintf("%s='%s'", users_DBModels.COLUMN_EMAIL, usr.Email) // Create a filter string to match the user ID

	if err := u.UserDBClient.UpdateUser(ctx, filter, map[string]interface{}{
		users_DBModels.COLUMN_PASSWORD: hashedPassword,
	}); err != nil {
		errorMsg := fmt.Sprintf("%s: %v", constants.INTERNAL_SERVER_ERROR, err)
		log.Error(errorMsg)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.INTERNAL_SERVER_ERROR, err)
		return
	}

	r, _ := u.UserDBClient.GetUser(ctx, fmt.Sprintf("%s='%s'",
		users_DBModels.COLUMN_EMAIL, usr.Email,
	))
	r.Password = ""

	// Respond with success message and the updated user profile (with password field cleared)
	controller.RespondWithSuccess(c, http.StatusAccepted, constants.UPDATED_SUCCESSFULLY, r)
}
