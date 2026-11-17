package handler

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/middleware"
	services "agrawalManas/Valet/internal/service"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateBookingHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var currentUser = middleware.MustCurrentUser(ctx)
		var req dto.CreateBookingRequest
		var err error

		err = ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		if req.BookingMode == constants.BOOKING_MODE_B2B && currentUser.Role == constants.ROLE_CUSTOMER {
			ctx.JSON(http.StatusForbidden, CustomUtils.HandleErrorResponse("Cannot create a B2B booking with "+constants.ROLE_CUSTOMER+" role"))
			return
		}

		if req.BookingMode == constants.BOOKING_MODE_B2C && currentUser.Role == constants.ROLE_SUPERVISOR {
			ctx.JSON(http.StatusForbidden, CustomUtils.HandleErrorResponse("Cannot create a B2C booking with "+constants.ROLE_SUPERVISOR+" role"))
			return
		}

		err = req.ValidateBookingRequest()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		booking, err := services.CreateBooking(ctx, pool, currentUser.ID, req)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}
		ctx.JSON(http.StatusCreated, CustomUtils.HandleSuccessResponse(booking))
	}
}
