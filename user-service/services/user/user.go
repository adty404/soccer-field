package services

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
	"user-service/config"
	"user-service/constants"
	errConstant "user-service/constants/error"
	"user-service/domain/dto"
	"user-service/domain/models"
	"user-service/repositories"
)

type UserServices struct {
	repository repositories.IRepositoryRegistry
}

type IUserServices interface {
	Login(context.Context, *dto.LoginRequest) (*dto.LoginResponse, error)
	Register(context.Context, *dto.RegisterRequest) (*dto.RegisterResponse, error)
	Update(context.Context, *dto.UpdateRequest, string) (*dto.UserResponse, error)
	GetUserLogin(context.Context) (*dto.UserResponse, error)
	GetUserByUUID(context.Context, string) (*dto.UserResponse, error)
}

type Claims struct {
	User *dto.UserResponse
	jwt.RegisteredClaims
}

func NewUserService(repository repositories.IRepositoryRegistry) IUserServices {
	return &UserServices{repository: repository}
}

func (u *UserServices) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := u.repository.GetUser().FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, err
	}

	expirationTime := time.Now().Add(time.Duration(config.Config.JwtExpirationTime) * time.Minute).Unix()
	data := &dto.UserResponse{
		UUID:        user.UUID,
		Name:        user.Name,
		Username:    user.Username,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        strings.ToLower(user.Role.Code),
	}

	claims := &Claims{
		User: data,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(expirationTime, 0)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Config.JwtSecretKey))
	if err != nil {
		return nil, err
	}

	response := &dto.LoginResponse{
		User:  *data,
		Token: tokenString,
	}

	return response, nil
}

func (u *UserServices) isUsernameExist(ctx context.Context, username string) bool {
	user, err := u.repository.GetUser().FindByUsername(ctx, username)
	if err != nil {
		return false
	}

	if user != nil {
		return true
	}

	return false
}

func (u *UserServices) isEmailExist(ctx context.Context, email string) bool {
	user, err := u.repository.GetUser().FindByEmail(ctx, email)
	if err != nil {
		return false
	}

	if user != nil {
		return true
	}

	return false
}

func (u *UserServices) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	if u.isUsernameExist(ctx, req.Username) {
		return nil, errConstant.ErrUsernameExist
	}
	if u.isEmailExist(ctx, req.Email) {
		return nil, errConstant.ErrEmailExist
	}
	if *req.Password != req.ConfirmPassword {
		return nil, errConstant.ErrPasswordDoesNotMatch
	}

	password := string(hashedPassword)
	user, err := u.repository.GetUser().Register(
		ctx, &dto.RegisterRequest{
			Name:        req.Name,
			Username:    req.Username,
			Password:    &password,
			Email:       req.Email,
			PhoneNumber: req.PhoneNumber,
			RoleID:      req.RoleID,
		},
	)
	if err != nil {
		return nil, err
	}

	response := &dto.RegisterResponse{
		User: dto.UserResponse{
			UUID:        user.UUID,
			Name:        user.Name,
			Username:    user.Username,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Role:        strings.ToLower(user.Role.Code),
		},
	}

	return response, nil
}

func (u *UserServices) Update(ctx context.Context, request *dto.UpdateRequest, uuid string) (*dto.UserResponse, error) {
	var (
		password                  string
		checkUsername, checkEmail *models.User
		hashedPassword            []byte
		user, userResult          *models.User
		err                       error
		data                      dto.UserResponse
	)

	user, err = u.repository.GetUser().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	isUsernameExist := u.isUsernameExist(ctx, request.Username)
	if isUsernameExist && user.Username != request.Username {
		checkUsername, err = u.repository.GetUser().FindByUsername(ctx, request.Username)
		if err != nil {
			return nil, err
		}

		if checkUsername != nil {
			return nil, errConstant.ErrUsernameExist
		}
	}

	isEmailExist := u.isEmailExist(ctx, request.Email)
	if isEmailExist && user.Email != request.Email {
		checkEmail, err = u.repository.GetUser().FindByEmail(ctx, request.Email)
		if err != nil {
			return nil, err
		}

		if checkEmail != nil {
			return nil, errConstant.ErrEmailExist
		}
	}

	if request.Password != nil {
		if *request.Password != *request.ConfirmPassword {
			return nil, errConstant.ErrPasswordDoesNotMatch
		}

		hashedPassword, err = bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}

		password = string(hashedPassword)
	}

	userResult, err = u.repository.GetUser().Update(
		ctx, &dto.UpdateRequest{
			Name:        request.Name,
			Username:    request.Username,
			Password:    &password,
			Email:       request.Email,
			PhoneNumber: request.PhoneNumber,
			RoleID:      request.RoleID,
		}, uuid,
	)
	if err != nil {
		return nil, err
	}

	data = dto.UserResponse{
		UUID:        userResult.UUID,
		Name:        userResult.Name,
		Username:    userResult.Username,
		Email:       userResult.Email,
		PhoneNumber: userResult.PhoneNumber,
	}

	return &data, nil
}

func (u *UserServices) GetUserLogin(ctx context.Context) (*dto.UserResponse, error) {
	var (
		userLogin = ctx.Value(constants.UserLogin).(*dto.UserResponse)
		data      dto.UserResponse
	)

	data = dto.UserResponse{
		UUID:        userLogin.UUID,
		Name:        userLogin.Name,
		Username:    userLogin.Username,
		Email:       userLogin.Email,
		Role:        userLogin.Role,
		PhoneNumber: userLogin.PhoneNumber,
	}

	return &data, nil
}

func (u *UserServices) GetUserByUUID(ctx context.Context, uuid string) (*dto.UserResponse, error) {
	user, err := u.repository.GetUser().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	data := dto.UserResponse{
		UUID:        user.UUID,
		Name:        user.Name,
		Username:    user.Username,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
	}

	return &data, nil
}
