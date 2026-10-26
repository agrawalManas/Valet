package handler

import (
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/middleware"
	services "agrawalManas/Valet/internal/service"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
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
