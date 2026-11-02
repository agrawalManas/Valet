package handler

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/dto"
	"agrawalManas/Valet/internal/middleware"
	"agrawalManas/Valet/internal/service"
	"agrawalManas/Valet/internal/storage"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxRequestSize = 50 << 20 // 50 MB
	maxMemory      = 8 << 20  // 8 MB
	maxFiles       = 10
)

func FileUploadHandler(pool *pgxpool.Pool, r2Client *storage.R2Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dto.FileUploadDTO

		user := middleware.MustCurrentUser(ctx)

		// Cap the body size first, before anything reads the body.
		// In our case the whole request should be less than 50MB.
		ctx.Request.Body = http.MaxBytesReader(
			ctx.Writer,
			ctx.Request.Body,
			maxRequestSize,
		)

		err := ctx.Request.ParseMultipartForm(maxMemory)
		if err != nil {
			var maxBytesError *http.MaxBytesError

			if errors.As(err, &maxBytesError) {
				ctx.JSON(http.StatusRequestEntityTooLarge,
					CustomUtils.HandleErrorResponse("Maximum request size is 50 MB"))
				return
			}
			ctx.JSON(http.StatusBadRequest,
				CustomUtils.HandleErrorResponse(constants.INVALID_FORM_DATA+": "+err.Error()))
			return
		}

		if ctx.Request.MultipartForm != nil {
			defer ctx.Request.MultipartForm.RemoveAll()
		}

		// The form is already parsed above, so this only maps the text fields
		// (visibility) onto req and runs the validator.
		err = ctx.ShouldBindWith(&req, binding.FormMultipart)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			return
		}

		// Get files parsed as a list. A single upload is a list of one.
		files := ctx.Request.MultipartForm.File["files"]
		if len(files) == 0 {
			ctx.JSON(http.StatusBadRequest,
				CustomUtils.HandleErrorResponse(`no files were sent: use the form field "files"`))
			return
		}

		// Pass the request context so the upload stops if the client disconnects.
		uploadedFiles, statusCode, err := service.UploadFiles(
			ctx.Request.Context(), files, r2Client, user.ID, CustomUtils.Deref(req.Visibility), pool,
		)
		if err != nil {
			// Bad file type, size or count should be a 400, not a generic 500.
			var verr *service.ValidationError
			if errors.As(err, &verr) {
				ctx.JSON(statusCode, CustomUtils.HandleErrorResponse(verr.Msg))
				return
			}
			CustomUtils.RespondRepoError(ctx, err)
			return
		}

		urls := make([]string, 0, len(uploadedFiles))
		for _, file := range uploadedFiles {
			urls = append(urls, file.Url)
		}

		ctx.JSON(http.StatusCreated, CustomUtils.HandleSuccessResponse(urls))
	}
}
