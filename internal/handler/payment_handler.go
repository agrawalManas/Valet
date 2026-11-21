package handler

import (
	"agrawalManas/Valet/internal/service"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UpdatePaymentStatusHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		bookingId := ctx.Query("booking_id")
		rideId := ctx.Query("ride_id")

		err := CustomUtils.ValidateUUIDs(bookingId, rideId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
		}

		err = service.UpdatePaymentStatusService(ctx, pool, bookingId, rideId)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}
	}
}
