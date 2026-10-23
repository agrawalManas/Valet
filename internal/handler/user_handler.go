package handler

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/models"
	"agrawalManas/Valet/internal/repository"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UpdateUserHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		value, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusBadRequest, constants.NO_USER_WITH_TOKEN)
			return
		}
		userId := value.(string)

		var request dto.UpdateUserRequest
		err := ctx.ShouldBindJSON(&request)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		if request.IsEmpty() {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.REQUEST_BODY_MISSING))
			return
		}

		var user *models.User
		user, err = repository.UpdateUser(pool, request.FullName, request.Email, request.Role, userId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				ctx.JSON(http.StatusNotFound, CustomUtils.HandleErrorResponse(constants.NO_USER_WITH_TOKEN))
				return
			}
			ctx.JSON(http.StatusInternalServerError, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		ctx.JSON(http.StatusOK, CustomUtils.HandleSuccessResponse(user))
	}
}
