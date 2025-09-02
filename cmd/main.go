package main

import (
	"demandscience/internal/handlers"
	"demandscience/internal/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	csvService := services.DSCsvProcessingService()
	csvHandler := handlers.DSCsvProcessorHandler(csvService)

	router := gin.Default()
	router.MaxMultipartMemory = 32 << 20

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

	log.Println("Starting Go backend server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}

}
