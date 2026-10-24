package CustomUtils

import (
	"agrawalManas/Valet/internal/constants"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type successResponse struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

type errorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func HandleErrorResponse(message string) errorResponse {
	return errorResponse{
		Status:  constants.RESPONSE_STATUS_FAILED,
		Message: message,
	}
}

func HandleSuccessResponse(data any) successResponse {
	return successResponse{
		Status: constants.RESPONSE_STATUS_SUCCESS,
		Data:   data,
	}
}

func RespondRepoError(ctx *gin.Context, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		ctx.JSON(http.StatusNotFound, HandleErrorResponse(constants.USER_NOT_FOUND))
		return
	}
	log.Printf("%s %s: %v", ctx.Request.Method, ctx.FullPath(), err)
	ctx.JSON(http.StatusInternalServerError, HandleErrorResponse(constants.SOMETHING_WENT_WRONG))
}
