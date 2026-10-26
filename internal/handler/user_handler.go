package handler

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/middleware"
	"agrawalManas/Valet/internal/repository"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UpdateProfileHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := middleware.MustCurrentUser(ctx)

		var req dto.UpdateProfileRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}
		if req.IsEmpty() {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.REQUEST_BODY_MISSING))
			return
		}

		updated, err := repository.UpdateUser(pool, user.ID, req.FullName, req.Email, req.Dob, req.Gender)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, CustomUtils.HandleSuccessResponse(updated))
	}
}

func AdminUpdateUserHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		admin := middleware.MustCurrentUser(ctx)

		var req dto.AdminUpdateUserRequest
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}
		if req.IsValid() {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.REQUEST_BODY_MISSING))
			return
		}

		// an admin shouldn't lock themselves out
		if req.TargetId == admin.ID &&
			((req.IsActive != nil && !*req.IsActive) || (req.Role != nil && *req.Role != constants.ROLE_ADMIN)) {
			ctx.JSON(http.StatusConflict, CustomUtils.HandleErrorResponse(constants.CANNOT_MODIFY_OWN_ACCESS))
			return
		}

		updated, err := repository.UpdateUserAccess(pool, req.TargetId, req.Role, req.IsActive)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		// _, err = repository.AddToUserAccessAudit(pool, req.TargetId, admin.ID, req.Role, updated.Role, req.IsActive, updated.IsActive)
		// if err != nil {
		// 	CustomUtils.RespondRepoError(ctx, err)
		// 	return
		// }

		ctx.JSON(http.StatusOK, CustomUtils.HandleSuccessResponse(updated))
	}
}
