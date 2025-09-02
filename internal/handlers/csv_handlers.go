package handlers

import (
	"demandscience/internal/models"
	"demandscience/internal/services"
	"encoding/base64"
	"log"
	"net/http"
	"time"

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
	startTime := time.Now()
	clientIP := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")

	log.Printf("[UPLOAD] Starting file upload request from IP: %s, User-Agent: %s", clientIP, userAgent)

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		log.Printf("[UPLOAD] [ERROR] No file provided in request from IP: %s, Error: %v", clientIP, err)
		ctx.JSON(http.StatusBadRequest, models.UploadResponse{
			Error: "No file provided",
		})
		return
	}

	jobID, err := handler.csvService.ProcessFile(fileHeader)
	if err != nil {
		log.Printf("[UPLOAD] [ERROR] File processing initiation failed - File: %s, IP: %s, Error: %v",
			fileHeader.Filename, clientIP, err)
		ctx.JSON(http.StatusBadRequest, models.UploadResponse{
			Error: err.Error(),
		})
		return
	}

	duration := time.Since(startTime)
	log.Printf("[UPLOAD] [SUCCESS] File upload successful - JobID: %s, File: %s, Size: %d bytes, Duration: %v, IP: %s",
		jobID, fileHeader.Filename, fileHeader.Size, duration, clientIP)

	ctx.JSON(http.StatusOK, models.UploadResponse{
		ID: jobID,
	})
}

func (handler *CsvProcessorHandler) DownloadFile(ctx *gin.Context) {
	startTime := time.Now()
	clientIP := ctx.ClientIP()
	downloadParam := ctx.Query("download")
	jobID := ctx.Param("id")

	log.Printf("[DOWNLOAD] Starting download/status request - JobID: %s, IP: %s, Download: %s",
		jobID, clientIP, downloadParam)

	job := handler.csvService.GetJob(jobID)
	if job == nil {
		log.Printf("[DOWNLOAD] [ERROR] Job not found - JobID: %s, IP: %s", jobID, clientIP)
		ctx.JSON(http.StatusBadRequest, models.UploadResponse{
			Error: "Invalid job ID",
		})
		return
	}

	log.Printf("[DOWNLOAD] Job found - JobID: %s, Status: %s, OriginalFile: %s, CreatedAt: %v",
		jobID, job.Status, job.OriginalFileName, job.CreatedAt)

	switch job.Status {
	case models.JobStatusInProgress:
		duration := time.Since(startTime)
		log.Printf("[DOWNLOAD] [STATUS] Job in progress - JobID: %s, IP: %s, ProcessingTime: %v, Duration: %v",
			jobID, clientIP, time.Since(job.CreatedAt), duration)

		ctx.JSON(http.StatusLocked, models.UploadResponse{
			Error: "Job is still in progress",
		})
		return

	case models.JobStatusFailed:
		duration := time.Since(startTime)
		log.Printf("[DOWNLOAD] [ERROR] Job failed - JobID: %s, IP: %s, TotalTime: %v, Duration: %v",
			jobID, clientIP, time.Since(job.CreatedAt), duration)

		ctx.JSON(http.StatusBadRequest, models.UploadResponse{
			Error: "Job failed to process",
		})
		return

	case models.JobStatusCompleted:
		processingTime := time.Since(job.CreatedAt)

		fileContent, err := handler.csvService.GetProcessedFile(jobID)
		if err != nil {
			log.Printf("[DOWNLOAD] [ERROR] Failed to read processed file - JobID: %s, Error: %v", jobID, err)
			ctx.JSON(http.StatusBadRequest, models.UploadResponse{
				Error: "Processed file not found",
			})
			return
		}
		encoded := base64.StdEncoding.EncodeToString(fileContent)

		log.Printf("[DOWNLOAD] [SUCCESS] Job completed - JobID: %s, ProcessingTime: %v, File: %s",
			jobID, processingTime, job.OriginalFileName)
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
		//*******************************************************************************************
		// filename := job.OriginalFileName + "_processed.csv"
		// c.Header("Content-Disposition", "attachment; filename="+filename)
		// c.Header("Content-Type", "application/octet-stream")
		// c.Data(http.StatusOK, "application/octet-stream", fileContent)

	default:
		duration := time.Since(startTime)
		log.Printf("[DOWNLOAD] [ERROR] Unknown job status - JobID: %s, Status: %s, IP: %s, Duration: %v",
			jobID, job.Status, clientIP, duration)

		ctx.JSON(http.StatusInternalServerError, models.UploadResponse{
			Error: "Unknown job status",
		})
	}
}
