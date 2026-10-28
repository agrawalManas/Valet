package handler

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/middleware"
	"agrawalManas/Valet/internal/repository"
	services "agrawalManas/Valet/internal/service"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AddValetSiteHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		admin := middleware.MustCurrentUser(ctx)

		var req dto.AddSiteRequest
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		site, err := services.CreateSite(ctx, pool, admin.ID, req)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}
		ctx.JSON(http.StatusCreated, CustomUtils.HandleSuccessResponse(site))
	}
}

func GetValetSitesHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var queryParams dto.GetSiteRequest
		var err error

		err = ctx.ShouldBindQuery(&queryParams)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		err = queryParams.Validate()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		sites, err := repository.GetSites(pool, queryParams.CreatedBy, queryParams.SupervisorId)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, CustomUtils.HandleSuccessResponse(sites))
	}
}

func UpdateSiteHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dto.UpdateSiteRequest
		var err error
		err = ctx.ShouldBindJSON(&req)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.INVALID_REQUEST_BODY+": "+err.Error()))
			return
		}

		err = req.Validate()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.INVALID_REQUEST_BODY+": "+err.Error()))
			return
		}

		id := ctx.Param("id")
		if id == "" {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.INVALID_USER_ID))
			return
		}

		user, err := repository.GetUserByMobile(pool, req.SupervisorMobileNumber)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				ctx.JSON(http.StatusForbidden, CustomUtils.HandleErrorResponse("Supervisor with this mobile number not found"))
				return
			}
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		if CustomUtils.Deref(user.Role) != constants.ROLE_SUPERVISOR {
			ctx.JSON(http.StatusForbidden, CustomUtils.HandleErrorResponse("Cannot add a site with this supervisor id - role mismatch"))
			return
		}

		if !user.IsActive {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse("Supervisor selected is not active"))
			return
		}

		site, err := repository.UpdateSite(pool, id, req, user.Id)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		site.SupervisorDetails = *user

		ctx.JSON(http.StatusOK, CustomUtils.HandleSuccessResponse(site))
	}
}
