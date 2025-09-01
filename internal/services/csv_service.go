package services

import (
	"demandscience/internal/models"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type CsvProcessingService struct {
	jobs       map[string]*models.ProcessingJob
	jobsMutex  sync.RWMutex
	storageDir string
}

func DSCsvProcessingService() *CsvProcessingService {
	storageDir := "processed_files"

	// Ensure storage directory exists
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create storage directory: %v", err))
	}

	return &CsvProcessingService{
		jobs:       make(map[string]*models.ProcessingJob),
		storageDir: storageDir,
	}
}

func (csvService *CsvProcessingService) ProcessFile(fileHeader *multipart.FileHeader) (string, error) {
	// Validate file
	if err := csvService.validateFile(fileHeader); err != nil {
		return "", err
	}

	jobID := uuid.New().String()
	job := models.DSProcessingJob(jobID, fileHeader.Filename)

	csvService.jobsMutex.Lock()
	csvService.jobs[jobID] = job
	csvService.jobsMutex.Unlock()

	// Process file asynchronously
	go csvService.processFileAsync(fileHeader, job)

	return jobID, nil
}

func (csvService *CsvProcessingService) GetJob(jobID string) *models.ProcessingJob {
	csvService.jobsMutex.RLock()
	defer csvService.jobsMutex.RUnlock()
	return csvService.jobs[jobID]
}

func (csvService *CsvProcessingService) GetProcessedFile(jobID string) ([]byte, error) {
	csvService.jobsMutex.RLock()
	job := csvService.jobs[jobID]
	csvService.jobsMutex.RUnlock()

	if job == nil || job.ProcessedFilePath == "" {
		return nil, errors.New("processed file not found")
	}

	return os.ReadFile(job.ProcessedFilePath)
}

func (csvService *CsvProcessingService) validateFile(fileHeader *multipart.FileHeader) error {

	filename := strings.ToLower(fileHeader.Filename)
	if !strings.HasSuffix(filename, ".csv") {
		return errors.New("invalid file type. Only CSV files are allowed")
	}

	// Check file size (10MB limit)
	const maxFileSize = 10 * 1024 * 1024
	if fileHeader.Size > maxFileSize {
		return errors.New("file size exceeds 10MB limit")
	}

	return nil
}

func (csvService *CsvProcessingService) processFileAsync(fileHeader *multipart.FileHeader, job *models.ProcessingJob) {
	defer func() {
		if r := recover(); r != nil {
			csvService.jobsMutex.Lock()
			job.Status = models.JobStatusFailed
			csvService.jobsMutex.Unlock()
		}
	}()

	// time.Sleep(15 * time.Second)

	if err := csvService.processFile(fileHeader, job); err != nil {
		csvService.jobsMutex.Lock()
		job.Status = models.JobStatusFailed
		csvService.jobsMutex.Unlock()
		return
	}

	csvService.jobsMutex.Lock()
	job.Status = models.JobStatusCompleted
	csvService.jobsMutex.Unlock()
}

func (csvService *CsvProcessingService) processFile(fileHeader *multipart.FileHeader, job *models.ProcessingJob) error {
	// Open uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Create output file
	outputPath := filepath.Join(csvService.storageDir, job.ID+"_processed.csv")
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Parse CSV
	reader := csv.NewReader(file)
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Read and process header
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read headers: %w", err)
	}

	// Add email flag column to headers
	newHeaders := append(headers, "has_email")
	if err := writer.Write(newHeaders); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	// Process each record
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read record: %w", err)
		}

		// Skip empty records
		if csvService.isEmptyRecord(record) {
			continue
		}

		// Check if any field contains a valid email
		hasEmail := false
		for _, field := range record {
			if emailRegex.MatchString(strings.TrimSpace(field)) {
				hasEmail = true
				break
			}
		}

		// Add email flag to record
		newRecord := append(record, fmt.Sprintf("%t", hasEmail))
		if err := writer.Write(newRecord); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	job.ProcessedFilePath = outputPath
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

// Cleanup old jobs (optional - run periodically)
func (s *CsvProcessingService) CleanupOldJobs() {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	for jobID, job := range s.jobs {
		if job.CreatedAt.Before(cutoff) {
			if job.ProcessedFilePath != "" {
				os.Remove(job.ProcessedFilePath)
			}
			delete(s.jobs, jobID)
		}
	}
}
