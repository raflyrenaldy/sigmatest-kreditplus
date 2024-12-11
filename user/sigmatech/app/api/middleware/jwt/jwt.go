package jwt

import (
	"errors"
	"user/sigmatech/app/constants"
	users_DBModels "user/sigmatech/app/db/dto/users"
	userDB "user/sigmatech/app/db/repository/user"

	"context"
	"encoding/json"
	"fmt"
	"time"
	"user/sigmatech/app/service/logger"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type IJwtService interface {
	GenerateUserTokens(ctx context.Context, crmUser users_DBModels.User) (*TokenDetails, error)
	VerifyUserToken(ctx context.Context, tokenString string) (*users_DBModels.User, bool)
	RefreshUserToken(ctx context.Context, tokenString string) (*TokenDetails, error)
	VerifyToken(ctx context.Context, tokenString string) (*users_DBModels.User, bool)
}

type JwtService struct {
	UserDBClient userDB.IUserRepository
}

func NewJwtService(UserDBClient userDB.IUserRepository) *JwtService {
	return &JwtService{
		UserDBClient: UserDBClient,
	}
}

type TokenDetails struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	AccessUuid   string `json:"access_uuid"`
	RefreshUuid  string `json:"refresh_uuid"`
	AtExpires    int64  `json:"at_expires"`
	RtExpires    int64  `json:"rt_expires"`
}

func (j *JwtService) GenerateUserTokens(ctx context.Context, crmUser users_DBModels.User) (*TokenDetails, error) {
	log := logger.Logger(ctx)
	log.Infof("Creating token for ", crmUser)

	crmUser.Password = ""
	var err error

	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * time.Duration(constants.Config.JwtConfig.JWT_ACCESS_EXP)).Unix()
	td.AccessUuid = uuid.NewString()

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user"] = crmUser
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)

	td.AccessToken, err = at.SignedString([]byte(constants.Config.JwtConfig.JWT_ACCESS_SECRET))
	if err != nil {
		return nil, err
	}

	//Creating Refresh Token
	td.RefreshUuid = td.AccessUuid + "++" + crmUser.Email
	td.RtExpires = time.Now().Add(time.Minute * time.Duration(constants.Config.JwtConfig.JWT_REFRESH_EXP)).Unix()

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user"] = crmUser
	rtClaims["exp"] = td.RtExpires

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(constants.Config.JwtConfig.JWT_REFRESH_SECRET))
	if err != nil {
		log.Errorf("error while generating access id ", err)
		return nil, err
	}
	return td, nil
}

func (j *JwtService) VerifyUserToken(ctx context.Context, tokenString string) (*users_DBModels.User, bool) {
	log := logger.Logger(ctx)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(constants.Config.JwtConfig.JWT_ACCESS_SECRET), nil
	})
	if err != nil {
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		user := users_DBModels.User{}
		jsonString, err := json.Marshal(claims["user"])
		if err != nil {
			log.Errorf("unable to marshal user claims", err.Error())
			return nil, false
		}
		// convert json to struct
		err = json.Unmarshal(jsonString, &user)
		if err != nil {
			return nil, false
		}

		u, err := j.UserDBClient.GetUser(ctx, fmt.Sprintf("%s='%s'",
			users_DBModels.COLUM_UUID, user.Uuid))
		if err != nil {
			return nil, false
		}

		return &u, true
	}
	return nil, false
}

func (j *JwtService) RefreshUserToken(ctx context.Context, tokenString string) (*TokenDetails, error) {
	log := logger.Logger(ctx)
	log.Infof("Refreshing token for: %s", tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(constants.Config.JwtConfig.JWT_REFRESH_SECRET), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, err
	}

	_, ok = claims["refresh_uuid"].(string)
	if !ok {
		return nil, err
	}

	userClaim, ok := claims["user"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid user claim")
	}

	user := users_DBModels.User{}
	userJSON, err := json.Marshal(userClaim)
	if err != nil {
		log.Errorf("unable to marshal user claim: %v", err)
		return nil, err
	}

	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		log.Errorf("unable to unmarshal user claim: %v", err)
		return nil, err
	}

	// Check if the user exists in the database
	u, err := j.UserDBClient.GetUser(ctx, fmt.Sprintf("%s='%s'", users_DBModels.COLUMN_EMAIL, user.Email))
	if err != nil {
		return nil, err
	}

	// Create new pairs of refresh and access tokens
	td, err := j.GenerateUserTokens(ctx, u)
	if err != nil {
		return nil, err
	}
	return td, nil
}

func (j *JwtService) VerifyToken(ctx context.Context, tokenString string) (*users_DBModels.User, bool) {
	log := logger.Logger(ctx)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(constants.Config.JwtConfig.JWT_ACCESS_SECRET), nil
	})
	if err != nil {
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, false
	}

	var user users_DBModels.User

	if userClaims, ok := claims["user"].(map[string]interface{}); ok {
		jsonString, err := json.Marshal(userClaims)
		if err != nil {
			log.Errorf("unable to marshal user claims: %v", err)
			return nil, false
		}

		err = json.Unmarshal(jsonString, &user)
		if err != nil {
			log.Errorf("unable to unmarshal user claims: %v", err)
			return nil, false
		}
	}

	if user.Uuid != uuid.Nil {
		u, err := j.UserDBClient.GetUser(ctx, fmt.Sprintf("%s='%s'",
			users_DBModels.COLUM_UUID, user.Uuid))
		if err != nil {
			return nil, false
		}

		return &u, true
	}

	return nil, false
}
