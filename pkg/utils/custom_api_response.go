package CustomUtils

import (
	"agrawalManas/Valet/internal/constants"
	"errors"
	"log"
	"net/http"
	"runtime"

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
	var verr *ValidationError
	if errors.As(err, &verr) {
		ctx.JSON(http.StatusBadRequest, HandleErrorResponse(verr.Msg))
		return
	}
	if errors.Is(err, pgx.ErrNoRows) {
		ctx.JSON(http.StatusNotFound, HandleErrorResponse(constants.NOT_FOUND))
		return
	}
	log.Printf("%s %s %s: %v", ctx.Request.Method, ctx.FullPath(), getParentFunctionName(), err)
	ctx.JSON(http.StatusInternalServerError, HandleErrorResponse(constants.SOMETHING_WENT_WRONG))
}

func getParentFunctionName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown method"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown function"
	}
	return fn.Name()
}
