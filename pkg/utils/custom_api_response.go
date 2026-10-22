package CustomUtils

import "agrawalManas/Valet/internal/constants"

type SuccessResponse struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func HandleErrorResponse(message string) ErrorResponse {
	return ErrorResponse{
		Status:  constants.RESPONSE_STATUS_FAILED,
		Message: message,
	}
}

func HandleSuccessResponse(data any) SuccessResponse {
	return SuccessResponse{
		Status: constants.RESPONSE_STATUS_SUCCESS,
		Data:   data,
	}
}
