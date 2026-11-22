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
		rideId := ctx.Param("ride_id")

		err := CustomUtils.ValidateUUIDs(rideId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
		}

		err = service.UpdatePaymentStatusService(ctx, pool, rideId)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}
	}
}
