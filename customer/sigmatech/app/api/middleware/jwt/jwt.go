package jwt

import (
	"customer/sigmatech/app/constants"
	customers_DBModels "customer/sigmatech/app/db/dto/customers"
	customerDB "customer/sigmatech/app/db/repository/customer"
	"errors"

	"context"
	"customer/sigmatech/app/service/logger"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type IJwtService interface {
	GenerateCustomerTokens(ctx context.Context, customer customers_DBModels.Customer) (*TokenDetails, error)
	VerifyCustomerToken(ctx context.Context, tokenString string) (*customers_DBModels.Customer, bool)
	RefreshCustomerToken(ctx context.Context, tokenString string) (*TokenDetails, error)
	VerifyToken(ctx context.Context, tokenString string) (*customers_DBModels.Customer, bool)
}

type JwtService struct {
	CustomerDBClient customerDB.ICustomerRepository
}

func NewJwtService(CustomerDBClient customerDB.ICustomerRepository) *JwtService {
	return &JwtService{
		CustomerDBClient: CustomerDBClient,
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

func (j *JwtService) GenerateCustomerTokens(ctx context.Context, customer customers_DBModels.Customer) (*TokenDetails, error) {
	log := logger.Logger(ctx)
	log.Infof("Creating token for ", customer)

	customer.Password = ""
	var err error

	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * time.Duration(constants.Config.JwtConfig.JWT_ACCESS_EXP)).Unix()
	td.AccessUuid = uuid.NewString()

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["customer"] = customer
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)

	td.AccessToken, err = at.SignedString([]byte(constants.Config.JwtConfig.JWT_ACCESS_SECRET))
	if err != nil {
		return nil, err
	}

	//Creating Refresh Token
	td.RefreshUuid = td.AccessUuid + "++" + customer.Email
	td.RtExpires = time.Now().Add(time.Minute * time.Duration(constants.Config.JwtConfig.JWT_REFRESH_EXP)).Unix()

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["customer"] = customer
	rtClaims["exp"] = td.RtExpires

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(constants.Config.JwtConfig.JWT_REFRESH_SECRET))
	if err != nil {
		log.Errorf("error while generating access id ", err)
		return nil, err
	}
	return td, nil
}

func (j *JwtService) VerifyCustomerToken(ctx context.Context, tokenString string) (*customers_DBModels.Customer, bool) {
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
		customer := customers_DBModels.Customer{}
		jsonString, err := json.Marshal(claims["customer"])
		if err != nil {
			log.Errorf("unable to marshal customer claims", err.Error())
			return nil, false
		}
		// convert json to struct
		err = json.Unmarshal(jsonString, &customer)
		if err != nil {
			return nil, false
		}

		u, err := j.CustomerDBClient.GetCustomer(ctx, fmt.Sprintf("%s='%s'",
			customers_DBModels.COLUM_UUID, customer.Uuid))
		if err != nil {
			return nil, false
		}

		return &u, true
	}
	return nil, false
}

func (j *JwtService) RefreshCustomerToken(ctx context.Context, tokenString string) (*TokenDetails, error) {
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

	customerClaim, ok := claims["customer"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid customer claim")
	}

	customer := customers_DBModels.Customer{}
	customerJSON, err := json.Marshal(customerClaim)
	if err != nil {
		log.Errorf("unable to marshal customer claim: %v", err)
		return nil, err
	}

	err = json.Unmarshal(customerJSON, &customer)
	if err != nil {
		log.Errorf("unable to unmarshal customer claim: %v", err)
		return nil, err
	}

	// Check if the customer exists in the database
	u, err := j.CustomerDBClient.GetCustomer(ctx, fmt.Sprintf("%s='%s'", customers_DBModels.COLUMN_EMAIL, customer.Email))
	if err != nil {
		return nil, err
	}

	// Create new pairs of refresh and access tokens
	td, err := j.GenerateCustomerTokens(ctx, u)
	if err != nil {
		return nil, err
	}
	return td, nil
}

func (j *JwtService) VerifyToken(ctx context.Context, tokenString string) (*customers_DBModels.Customer, bool) {
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

	var customer customers_DBModels.Customer

	if customerClaims, ok := claims["customer"].(map[string]interface{}); ok {
		jsonString, err := json.Marshal(customerClaims)
		if err != nil {
			log.Errorf("unable to marshal customer claims: %v", err)
			return nil, false
		}

		err = json.Unmarshal(jsonString, &customer)
		if err != nil {
			log.Errorf("unable to unmarshal customer claims: %v", err)
			return nil, false
		}
	}

	if customer.Uuid != uuid.Nil {
		u, err := j.CustomerDBClient.GetCustomer(ctx, fmt.Sprintf("%s='%s'",
			customers_DBModels.COLUM_UUID, customer.Uuid))
		if err != nil {
			return nil, false
		}

		if !u.IsActive {
			return nil, false
		}

		return &u, true
	}

	return nil, false
}
