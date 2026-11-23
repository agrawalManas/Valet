package handler

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/middleware"
	"agrawalManas/Valet/internal/models"
	"agrawalManas/Valet/internal/repository"
	"agrawalManas/Valet/internal/service"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func StartRideHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err error
		var currentUser = middleware.MustCurrentUser(ctx)

		bookingId := ctx.Query("booking_id")
		vrn := ctx.Query("vrn")

		if vrn == "" {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse("Please provide the vehicle number"))
			return
		}

		err = CustomUtils.ValidateUUIDs(bookingId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse("Please provide a valid booking id"))
			return
		}

		ride, err := service.StartRideService(ctx, pool, bookingId, vrn, currentUser)
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

		err = req.Validate(false, false)
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

func AcceptPickDropRequestHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dto.UpdateRideRequest
		currentUser := middleware.MustCurrentUser(ctx)

		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		err = req.Validate(false, true)
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

		err = req.Validate(true, false)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		ride, err := service.RequestGMCService(pool, req, currentUser)
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

func GetDriverRidesHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var siteId = ctx.Param("id")
		var currentUser = middleware.MustCurrentUser(ctx)
		var err error

		err = CustomUtils.ValidateUUIDs(siteId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.Invalid(err.Error()))
		}

		driver, err := repository.GetDriverForSite(ctx, pool, currentUser.ID, siteId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				ctx.JSON(http.StatusNotFound, CustomUtils.HandleErrorResponse("No sites found with the given site id"))
				return
			}
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		err = validateDriver(driver)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		bookings, err := repository.GetActiveBookingsForASite(pool, driver.SiteId)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		if len(bookings) == 0 {
			ctx.JSON(http.StatusOK, CustomUtils.HandleSuccessResponse([]models.Booking{}))
			return
		}
		ctx.JSON(http.StatusOK, CustomUtils.HandleSuccessResponse(bookings))
	}
}

func validateDriver(driver *models.Driver) error {
	var err error

	err = handleDriverCurrentStatus(driver.CurrentStatus)
	if err != nil {
		return err
	}

	err = handleDriverKYCStatus(driver.KYCStatus)
	if err != nil {
		return err
	}

	return err
}

func handleDriverCurrentStatus(currentStatus string) error {
	if currentStatus == constants.DRIVER_STATUS_AVAILABLE || currentStatus == constants.DRIVER_STATUS_ON_DUTY {
		return nil
	}
	return CustomUtils.Invalid("The driver is " + currentStatus + " at the moment!")
}

func handleDriverKYCStatus(kycStatus string) error {
	if kycStatus == constants.KYC_STATUS_VERIFIED {
		return nil
	}
	return CustomUtils.Invalid("Couldn't get drive requests when the kyc status " + kycStatus + "")
}

func AddRideFeedbackHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dto.AddRideReviewDTO
		var currentUser = middleware.MustCurrentUser(ctx)
		var err error

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

		err = CustomUtils.ValidateUUIDs(req.RideId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		rideReview, err := service.AddRideReview(pool, req, currentUser)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, CustomUtils.HandleSuccessResponse(rideReview))
	}
}
