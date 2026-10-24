package handler

import (
	"agrawalManas/Valet/internal/config"
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/repository"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GenerateOTPHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		mobileNumber := ctx.Query("mobile_number")

		if mobileNumber == "" {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.INVALID_MOBILE_NUMBER))
			return
		}

		var resp = map[string]string{"otp": "8888"}
		ctx.JSON(http.StatusOK, CustomUtils.HandleSuccessResponse(resp))
	}
}

func VerifyOTPHandler(pool *pgxpool.Pool, cfg *config.Config, role string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request *dto.VerifyOTPRequest

		err := ctx.ShouldBindJSON(&request)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		mobileNumber := request.MobileNumber
		if request.Otp != "8888" {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.INVALID_OTP))
			return
		}

		user, err := repository.GetUserByMobile(pool, mobileNumber)
		if err != nil {
			if !(errors.Is(err, pgx.ErrNoRows)) {
				ctx.JSON(http.StatusInternalServerError, CustomUtils.HandleErrorResponse(err.Error()))
				return
			} else {
				user, err = repository.CreateUser(pool, mobileNumber, role)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, CustomUtils.HandleErrorResponse(err.Error()))
					return
				}
			}
		}

		if *user.Role != role {
			ctx.JSON(http.StatusForbidden, CustomUtils.HandleErrorResponse("Cannot login as "+role+" role with this mobile number"))
			return
		}

		if !user.IsActive {
			ctx.JSON(http.StatusForbidden, CustomUtils.HandleErrorResponse(constants.USER_INACTIVE))
			return
		}

		token, err := CustomUtils.NewJWT(*user.Role, user.Id, user.IsActive, cfg)
		if err != nil || token == "" {
			ctx.JSON(http.StatusInternalServerError, CustomUtils.HandleErrorResponse(constants.FAILED_TO_GENERATE_TOKEN+"for user with number: "+mobileNumber))
			return
		}
		var resp = map[string]any{"user_details": &user, "token": token}

		ctx.JSON(http.StatusOK, CustomUtils.HandleSuccessResponse(resp))
	}
}
