package handlers

import (
	"demandscience/internal/models"
	"demandscience/internal/services"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CsvProcessorHandler struct {
	csvService *services.CsvProcessingService
}

func DSCsvProcessorHandler(csvService *services.CsvProcessingService) *CsvProcessorHandler {
	return &CsvProcessorHandler{
		csvService: csvService,
	}
}

func (handler *CsvProcessorHandler) UploadFile(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.UploadResponse{
			Error: "No file provided",
		})
		return
	}

	jobID, err := handler.csvService.ProcessFile(fileHeader)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.UploadResponse{
			Error: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.UploadResponse{
		ID: jobID,
	})
}

func (handler *CsvProcessorHandler) DownloadFile(ctx *gin.Context) {
	jobID := ctx.Param("id")
	job := handler.csvService.GetJob(jobID)
	if job == nil {
		ctx.JSON(http.StatusBadRequest, models.UploadResponse{
			Error: "Invalid job ID",
		})
		return
	}

	switch job.Status {
	case models.JobStatusInProgress:
		ctx.JSON(http.StatusLocked, models.UploadResponse{
			Error: "Job is still in progress",
		})
		return

	case models.JobStatusFailed:
		ctx.JSON(http.StatusBadRequest, models.UploadResponse{
			Error: "Job failed to process",
		})
		return

	case models.JobStatusCompleted:
		fileContent, err := handler.csvService.GetProcessedFile(jobID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, models.UploadResponse{
				Error: "Processed file not found",
			})
			return
		}
		encoded := base64.StdEncoding.EncodeToString(fileContent)
		ctx.JSON(http.StatusOK, gin.H{
			"id":             jobID,
			"status":         string(job.Status),
			"message":        "File processed successfully",
			"filename":       job.OriginalFileName,
			"processed_name": job.OriginalFileName + "_processed.csv",
			"file_data":      encoded,
			"content_type":   "text/csv",
			"size":           len(fileContent),
			"created_at":     job.CreatedAt,
		})
		return

		// This code below is if we want directly download the file then comment the above http.StatusOK code and
		// uncomment the below code.
		// filename := job.OriginalFileName + "_processed.csv"
		// c.Header("Content-Disposition", "attachment; filename="+filename)
		// c.Header("Content-Type", "application/octet-stream")
		// c.Data(http.StatusOK, "application/octet-stream", fileContent)

	default:
		ctx.JSON(http.StatusInternalServerError, models.UploadResponse{
			Error: "Unknown job status",
		})
	}
}
