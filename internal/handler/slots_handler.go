package handler

import (
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/repository"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateSlotHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dto.CreateSlotRequest

		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		err = dto.ValidateTime(req.StartTime, req.EndTime)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		// site, err := repository.GetSiteBySiteId(pool, req.SiteId)
		// if err != nil {
		// 	CustomUtils.RespondRepoError(ctx, err)
		// 	return
		// }

		// if !site.IsBookingAllowed {
		// 	ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse("Booking for this site is not enabled, cannot create a new slot"))
		// 	return
		// }

		slot, err := repository.CreateSlot(pool, req.StartTime, req.EndTime, req.SiteId, req.IsOpen)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, CustomUtils.HandleSuccessResponse(slot))
	}
}

func UpdateSlotHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dto.UpdateSlotRequest

		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		err = req.ValidateUpdateSlotRequest()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		slot, err := repository.UpdateSlot(pool, req.SlotId, req.StartTime, req.EndTime, req.IsOpen)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, CustomUtils.HandleSuccessResponse(slot))
	}
}
