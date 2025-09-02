package main

import (
	"demandscience/internal/handlers"
	"demandscience/internal/services"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var maxFileSize int64

func main() {
	err := godotenv.Load()
	port := os.Getenv("PORT")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	maxFileSizeStr := os.Getenv("MAX_FILE_SIZE_MB")
	maxFileSize, strErr := strconv.Atoi(maxFileSizeStr)
	if strErr != nil {
		log.Fatal("Error loading .env file")
	}

	csvService := services.DSCsvProcessingService()
	csvHandler := handlers.DSCsvProcessorHandler(csvService)

	router := gin.Default()
	router.MaxMultipartMemory = int64(maxFileSize) << 20

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Go backend server is running!",
			"status":  "ok",
		})
	})

	api := router.Group("/API")
	{
		api.POST("/upload", csvHandler.UploadFile)
		api.GET("/download/:id", csvHandler.DownloadFile)
	}

	log.Println("Starting Go backend server on :", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}

}
