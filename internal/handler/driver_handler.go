package handler

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/middleware"
	"agrawalManas/Valet/internal/models"
	"agrawalManas/Valet/internal/repository"
	services "agrawalManas/Valet/internal/service"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AddDriverToSiteHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		supervisor := middleware.MustCurrentUser(ctx)

		var req dto.CreateDriverRequest
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		driver, err := services.AddDriverToSite(ctx, pool, supervisor.ID, req)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}
		ctx.JSON(http.StatusCreated, CustomUtils.HandleSuccessResponse(driver))
	}
}

func GetAllDriversForSiteHandler(pool *pgxpool.Pool, bySiteId bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUser := middleware.MustCurrentUser(ctx)
		var id string

		if !bySiteId && currentUser.Role == constants.ROLE_SUPERVISOR {
			id = currentUser.ID
		} else if !bySiteId && currentUser.Role == constants.ROLE_ADMIN {
			id = ctx.Query("supervisor_id")

			if id == "" {
				ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.INVALID+" "+constants.QUERY_PARAMS+": "+constants.SUPERVISOR_ID+" is missing"))
				return
			}
		} else {
			id = ctx.Param("id")
			if id == "" {
				ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.SITE_ID+" is missing"))
				return
			}
		}

		drivers, err := repository.GetAllDriversForSite(pool, id, bySiteId)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, CustomUtils.HandleSuccessResponse(drivers))
	}
}

func UpdateSiteDriverHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dto.UpdateDriverRequest
		supervisor := middleware.MustCurrentUser(ctx)

		var err error
		var driver *models.Driver

		err = ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		err = req.ValidateUpdateDriverRequest()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}
		driver, err = repository.UpdateDriver(pool, req.Id, req.Status, req.ShiftStartTime, req.ShiftStartTime, req.TotalExperience, supervisor.ID)

		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, CustomUtils.HandleSuccessResponse(driver))
	}
}

func ChangeSupervisorForDriverHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dto.ChangeSupervisorForDriverRequest
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		err = validateIDs(req.DriverId, req.NewSuperVisorId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		driver, err := services.ChangeSupervisorForDriver(ctx, pool, req.DriverId, req.NewSuperVisorId)
		if err != nil {
			CustomUtils.RespondRepoError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, CustomUtils.HandleSuccessResponse(driver))
	}
}

func validateIDs(driverId, supervisorId string) error {
	var err error

	if driverId == supervisorId {
		err = CustomUtils.Invalid("Given driver_id and new_supervisor_id is same")
	}
	if err != nil {
		return err
	}

	err = CustomUtils.ValidateUUIDS(driverId, supervisorId)
	if err != nil {
		return err
	}

	return nil
}
