package handler

import (
	"agrawalManas/Valet/internal/middleware"
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

		if bookingId == "" {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse("Please provide a booking id"))
			return
		}
		if vrn == "" {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse("Please provide the vehicle number"))
			return
		}

		err = CustomUtils.ValidateUUIDs(bookingId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse("Please provide a valid booking id"))
			return
		}

		ride, _, err := service.CreateRideService(ctx, pool, bookingId, vrn, currentUser)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, CustomUtils.HandleSuccessResponse(ride))
	}
}
