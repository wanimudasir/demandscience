package services

import (
	"demandscience/internal/models"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var MaxFileSize int

type CsvProcessingService struct {
	jobs       map[string]*models.ProcessingJob
	jobsMutex  sync.RWMutex
	storageDir string
}

func init() {
	// Load .env if available
	_ = godotenv.Load()

	maxFileSizeStr := os.Getenv("MAX_FILE_SIZE")
	if maxFileSizeStr == "" {
		MaxFileSize = 10 * 1024 * 1024 // default 10MB
	} else {
		val, err := strconv.Atoi(maxFileSizeStr)
		if err != nil {
			log.Fatalf("Invalid MAX_FILE_SIZE in .env: %v", err)
		}
		MaxFileSize = val
	}
}

func DSCsvProcessingService() *CsvProcessingService {
	// storageDir := "processed_files"
	workingDir, _ := os.Getwd()
	storageDir := filepath.Join(workingDir, "processed_files")

	log.Printf("[SERVICE] [INIT] Initializing CSV Processing Service")
	log.Printf("[SERVICE] [INIT] Storage directory: %s", storageDir)

	// Ensure storage directory exists
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Fatalf("[SERVICE] [INIT] [FATAL] Failed to create storage directory: %v", err)
	}

	log.Printf("[SERVICE] [INIT] CSV Processing Service initialized successfully")
	return &CsvProcessingService{
		jobs:       make(map[string]*models.ProcessingJob),
		storageDir: storageDir,
	}
}

func (csvService *CsvProcessingService) ProcessFile(fileHeader *multipart.FileHeader) (string, error) {
	log.Printf("[SERVICE] [PROCESS] Starting file processing - File: %s, Size: %d bytes",
		fileHeader.Filename, fileHeader.Size)

	// Validate file
	if err := csvService.validateFile(fileHeader); err != nil {
		log.Printf("[SERVICE] [PROCESS] [ERROR] File validation failed - File: %s, Error: %v",
			fileHeader.Filename, err)
		return "", err
	}

	log.Printf("[SERVICE] [PROCESS] File validation passed - File: %s", fileHeader.Filename)

	// Generate job ID
	jobID := uuid.New().String()
	log.Printf("[SERVICE] [PROCESS] Generated job ID: %s for file: %s", jobID, fileHeader.Filename)

	// Create job
	job := models.DSProcessingJob(jobID, fileHeader.Filename)

	csvService.jobsMutex.Lock()
	csvService.jobs[jobID] = job
	activeJobs := len(csvService.jobs)
	csvService.jobsMutex.Unlock()

	log.Printf("[SERVICE] [PROCESS] Job created and stored - JobID: %s, ActiveJobs: %d", jobID, activeJobs)

	// Process file asynchronously
	go csvService.processFileAsync(fileHeader, job)

	log.Printf("[SERVICE] [PROCESS] [SUCCESS] File processing initiated - JobID: %s, File: %s",
		jobID, fileHeader.Filename)
	return jobID, nil
}

func (csvService *CsvProcessingService) GetJob(jobID string) *models.ProcessingJob {
	log.Printf("[SERVICE] [GET_JOB] Retrieving job - JobID: %s", jobID)

	csvService.jobsMutex.RLock()
	job := csvService.jobs[jobID]
	totalJobs := len(csvService.jobs)
	csvService.jobsMutex.RUnlock()

	if job == nil {
		log.Printf("[SERVICE] [GET_JOB] [ERROR] Job not found - JobID: %s, TotalJobs: %d", jobID, totalJobs)
	} else {
		log.Printf("[SERVICE] [GET_JOB] [SUCCESS] Job retrieved - JobID: %s, Status: %s, File: %s",
			jobID, job.Status, job.OriginalFileName)
	}

	return job
}

func (csvService *CsvProcessingService) GetProcessedFile(jobID string) ([]byte, error) {
	log.Printf("[SERVICE] [GET_FILE] Retrieving processed file - JobID: %s", jobID)
	csvService.jobsMutex.RLock()
	job := csvService.jobs[jobID]
	csvService.jobsMutex.RUnlock()

	if job == nil {
		log.Printf("[SERVICE] [GET_FILE] [ERROR] Job not found - JobID: %s", jobID)
		return nil, errors.New("job not found")
	}

	if job.ProcessedFilePath == "" {
		log.Printf("[SERVICE] [GET_FILE] [ERROR] Processed file path empty - JobID: %s, Status: %s",
			jobID, job.Status)
		return nil, errors.New("processed file not found")
	}

	log.Printf("[SERVICE] [GET_FILE] Reading file from disk - JobID: %s, Path: %s",
		jobID, job.ProcessedFilePath)

	data, err := os.ReadFile(job.ProcessedFilePath)

	if err != nil {
		log.Printf("[SERVICE] [GET_FILE] [ERROR] Failed to read file - JobID: %s, Path: %s, Error: %v",
			jobID, job.ProcessedFilePath, err)
		return nil, fmt.Errorf("failed to read processed file: %v", err)
	}

	log.Printf("[SERVICE] [GET_FILE] [SUCCESS] File read successfully - JobID: %s, Size: %d bytes",
		jobID, len(data))

	return data, nil
}

func (csvService *CsvProcessingService) validateFile(fileHeader *multipart.FileHeader) error {
	log.Printf("[SERVICE] [VALIDATE] Starting file validation - File: %s", fileHeader.Filename)

	filename := strings.ToLower(fileHeader.Filename)
	if !strings.HasSuffix(filename, ".csv") {
		log.Printf("[SERVICE] [VALIDATE] [ERROR] Invalid file type - File: %s", fileHeader.Filename)
		return errors.New("invalid file type. Only CSV files are allowed")
	}

	if fileHeader.Size > int64(MaxFileSize) {
		log.Printf("[SERVICE] [VALIDATE] [ERROR] File too large - File: %s, Size: %d bytes, Limit: %d bytes",
			fileHeader.Filename, fileHeader.Size, MaxFileSize)
		return errors.New("file size exceeds 10MB limit")
	}

	log.Printf("[SERVICE] [VALIDATE] [SUCCESS] File validation passed - File: %s, Size: %d bytes",
		fileHeader.Filename, fileHeader.Size)

	return nil
}

func (csvService *CsvProcessingService) processFileAsync(fileHeader *multipart.FileHeader, job *models.ProcessingJob) {
	startTime := time.Now()
	log.Printf("[SERVICE] [ASYNC] Starting async processing - JobID: %s, File: %s",
		job.ID, job.OriginalFileName)

	defer func() {
		if r := recover(); r != nil {
			csvService.jobsMutex.Lock()
			job.Status = models.JobStatusFailed
			csvService.jobsMutex.Unlock()
			log.Printf("[SERVICE] [ASYNC] [ERROR] Job marked as failed due to panic - JobID: %s", job.ID)
		}
	}()

	// time.Sleep(15 * time.Second)

	if err := csvService.processFile(fileHeader, job); err != nil {
		duration := time.Since(startTime)
		log.Printf("[SERVICE] [ASYNC] [ERROR] Processing failed - JobID: %s, Error: %v, Duration: %v",
			job.ID, err, duration)

		csvService.jobsMutex.Lock()
		job.Status = models.JobStatusFailed
		csvService.jobsMutex.Unlock()
		return
	}

	duration := time.Since(startTime)
	csvService.jobsMutex.Lock()
	job.Status = models.JobStatusCompleted
	activeJobs := len(csvService.jobs)
	csvService.jobsMutex.Unlock()

	log.Printf("[SERVICE] [ASYNC] [SUCCESS] Processing completed - JobID: %s, File: %s, Duration: %v, ActiveJobs: %d",
		job.ID, job.OriginalFileName, duration, activeJobs)
}

func (csvService *CsvProcessingService) processFile(fileHeader *multipart.FileHeader, job *models.ProcessingJob) error {
	log.Printf("[SERVICE] [PROCESS_FILE] Starting file processing - JobID: %s", job.ID)

	// Open uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("[SERVICE] [PROCESS_FILE] [ERROR] Failed to open uploaded file - JobID: %s, Error: %v",
			job.ID, err)
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Create output file
	outputPath := filepath.Join(csvService.storageDir, job.ID+"_processed.csv")
	log.Printf("[SERVICE] [PROCESS_FILE] Creating output file - JobID: %s, Path: %s", job.ID, outputPath)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Printf("[SERVICE] [PROCESS_FILE] [ERROR] Failed to create output file - JobID: %s, Path: %s, Error: %v",
			job.ID, outputPath, err)
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Parse CSV
	reader := csv.NewReader(file)
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	log.Printf("[SERVICE] [PROCESS_FILE] Reading CSV headers - JobID: %s", job.ID)

	// Read and process header
	headers, err := reader.Read()
	if err != nil {
		log.Printf("[SERVICE] [PROCESS_FILE] [ERROR] Failed to read headers - JobID: %s, Error: %v",
			job.ID, err)
		return fmt.Errorf("failed to read headers: %w", err)
	}

	// Add email flag column to headers
	newHeaders := append(headers, "has_email")
	if err := writer.Write(newHeaders); err != nil {
		log.Printf("[SERVICE] [PROCESS_FILE] [ERROR] Failed to write headers - JobID: %s, Error: %v",
			job.ID, err)
		return fmt.Errorf("failed to write headers: %w", err)
	}

	log.Printf("[SERVICE] [PROCESS_FILE] Headers written with has_email column - JobID: %s", job.ID)

	// Process each record
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	recordCount := 0
	emailFoundCount := 0
	emptyRecordCount := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			log.Printf("[SERVICE] [PROCESS_FILE] Reached end of file - JobID: %s", job.ID)
			break
		}
		if err != nil {
			log.Printf("[SERVICE] [PROCESS_FILE] [ERROR] Failed to read record - JobID: %s, Record: %d, Error: %v",
				job.ID, recordCount+1, err)
			return fmt.Errorf("failed to read record: %w", err)
		}
		recordCount++

		// Skip empty records
		if csvService.isEmptyRecord(record) {
			emptyRecordCount++
			log.Printf("[SERVICE] [PROCESS_FILE] Skipping empty record - JobID: %s, Record: %d",
				job.ID, recordCount)
			continue
		}

		// Check if any field contains a valid email
		hasEmail := false
		emailField := ""
		for _, field := range record {
			if emailRegex.MatchString(strings.TrimSpace(field)) {
				hasEmail = true
				emailField = field
				emailFoundCount++
				break
			}
		}

		if hasEmail {
			log.Printf("[SERVICE] [PROCESS_FILE] Valid email found - JobID: %s, Record: %d, Email: %s",
				job.ID, recordCount, emailField)
		}

		// Add email flag to record
		newRecord := append(record, fmt.Sprintf("%t", hasEmail))
		if err := writer.Write(newRecord); err != nil {
			log.Printf("[SERVICE] [PROCESS_FILE] [ERROR] Failed to write record - JobID: %s, Record: %d, Error: %v",
				job.ID, recordCount, err)
			return fmt.Errorf("failed to write record: %w", err)
		}

		// Log progress for large files
		if recordCount%100 == 0 {
			log.Printf("[SERVICE] [PROCESS_FILE] Progress update - JobID: %s, ProcessedRecords: %d, EmailsFound: %d",
				job.ID, recordCount, emailFoundCount)
		}
	}

	log.Printf("[SERVICE] [PROCESS_FILE] File processing statistics - JobID: %s, TotalRecords: %d, EmailsFound: %d, EmptyRecords: %d",
		job.ID, recordCount, emailFoundCount, emptyRecordCount)

	// Flush and close writer
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Printf("[SERVICE] [PROCESS_FILE] [ERROR] CSV writer error - JobID: %s, Error: %v", job.ID, err)
		return fmt.Errorf("CSV writer error: %w", err)
	}

	job.ProcessedFilePath = outputPath

	// Get file size for logging
	if fileInfo, err := os.Stat(outputPath); err == nil {
		log.Printf("[SERVICE] [PROCESS_FILE] [SUCCESS] File processing completed - JobID: %s, InputSize: %d bytes, OutputSize: %d bytes, OutputPath: %s",
			job.ID, fileHeader.Size, fileInfo.Size(), outputPath)
	} else {
		log.Printf("[SERVICE] [PROCESS_FILE] [SUCCESS] File processing completed - JobID: %s, OutputPath: %s",
			job.ID, outputPath)
	}
	return nil
}

func (csvService *CsvProcessingService) isEmptyRecord(record []string) bool {
	for _, field := range record {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}
