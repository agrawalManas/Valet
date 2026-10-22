package handler

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/models"
	"agrawalManas/Valet/internal/repository"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UpdateUserHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		value, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find the user"})
			return
		}
		userId := value.(string)

		var request dto.UpdateUserRequest
		err := ctx.BindJSON(request)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		var user *models.User
		user, err = repository.UpdateUser(pool, request.FullName, request.Email, request.Role, userId)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				ctx.JSON(http.StatusInternalServerError, CustomUtils.HandleErrorResponse(constants.USER_EXISTS))
				return
			}
			ctx.JSON(http.StatusInternalServerError, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		ctx.JSON(http.StatusAccepted, CustomUtils.HandleSuccessResponse(user))
	}
}
