package CustomUtils

import "agrawalManas/Valet/internal/constants"

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
