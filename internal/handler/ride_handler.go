package handler

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/middleware"
	"agrawalManas/Valet/internal/repository"
	"agrawalManas/Valet/internal/service"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateRideHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err error
		var currentUser = middleware.MustCurrentUser(ctx)

		bookingId := ctx.Query("booking_id")
		vrn := ctx.Query("vrn")
		siteId := ctx.Query("site_id")

		if vrn == "" {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse("Please provide the vehicle number"))
			return
		}

		err = CustomUtils.ValidateUUIDs(bookingId, siteId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse("Please provide a valid booking id and site id"))
			return
		}

		ride, err := service.CreateRideService(ctx, pool, bookingId, vrn, currentUser)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, CustomUtils.HandleSuccessResponse(ride))
	}
}

func UpdateRideHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dto.UpdateRideRequest
		currentUser := middleware.MustCurrentUser(ctx)

		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		err = req.Validate()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		ride, err := service.UpdateRideService(ctx, pool, req, currentUser)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		ctx.JSON(http.StatusAccepted, CustomUtils.HandleSuccessResponse(ride))
	}
}

func RequestGMCHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err error
		var req dto.UpdateRideRequest
		currentUser := middleware.MustCurrentUser(ctx)

		err = ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		err = req.Validate()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		ride, err := service.RequestGMCService(ctx, pool, req, currentUser)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, CustomUtils.HandleSuccessResponse(ride))
	}
}

func RideTimelineHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUser := middleware.MustCurrentUser(ctx)

		rideId := ctx.Param("id")
		if rideId == "" {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse("Please provide a ride id"))
			return
		}

		err := CustomUtils.ValidateUUIDs(rideId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse("Please provide a valid ride id"))
			return
		}

		if currentUser.Role == constants.ROLE_SUPERVISOR {
			siteId, err := repository.GetSiteIDAgainstRideId(pool, rideId)
			if err != nil {
				CustomUtils.RespondRepoError(ctx, err)
				return
			}

			site, err := repository.GetSiteBySiteId(pool, CustomUtils.Deref(siteId))
			if err != nil {
				CustomUtils.RespondRepoError(ctx, err)
				return
			}

			if !(site.SupervisorId == currentUser.ID) {
				ctx.JSON(http.StatusForbidden, CustomUtils.HandleErrorResponse("Supervisor is not associated with this site"))
				return
			}
		}

		timeline, err := repository.GetRideTimeline(pool, rideId)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		ctx.JSON(http.StatusAccepted, CustomUtils.HandleSuccessResponse(timeline))
	}
}
