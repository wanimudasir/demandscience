package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"sync"
	"time"
)

func loadTest() {
	const numRequests = 10
	const serverURL = "http://localhost:8080"

	csvContent := "name,email\nJohn,john@test.com\nJane,jane@test.com"

	var wg sync.WaitGroup
	results := make(chan string, numRequests)

	start := time.Now()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Upload file
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			part, _ := writer.CreateFormFile("file", fmt.Sprintf("test%d.csv", id))
			part.Write([]byte(csvContent))
			writer.Close()

			resp, err := http.Post(serverURL+"/API/upload",
				writer.FormDataContentType(), &buf)
			if err != nil {
				results <- fmt.Sprintf("Request %d failed: %v", id, err)
				return
			}
			defer resp.Body.Close()

			var uploadResp map[string]string
			json.NewDecoder(resp.Body).Decode(&uploadResp)

			if uploadResp["error"] != "" {
				results <- fmt.Sprintf("Request %d error: %s", id, uploadResp["error"])
				return
			}

			results <- fmt.Sprintf("Request %d success: %s", id, uploadResp["id"])
		}(i)
	}

	wg.Wait()
	close(results)

	duration := time.Since(start)

	fmt.Printf("Load test completed in %v\n", duration)
	for result := range results {
		fmt.Println(result)
	}
}
