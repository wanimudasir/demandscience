
### Test Data Files


name,email,phone,company
John Doe,john.doe@example.com,123-456-7890,TechCorp
Jane Smith,invalid-email,098-765-4321,DesignCo
Bob Wilson,bob@company.org,555-123-4567,StartupXYZ
Mary Johnson,,999-888-7777,BigCorp
Alice Brown,alice.brown@email.net,111-222-3333,
Tom Davis,not-an-email-address,444-555-6666,RetailChain
```




## Test Cases 

### Test Case 1: Valid File Upload
**Objective**: Test successful file upload and processing

**Steps:**
1. **Upload Request**
   - Method: `POST`
   - URL: `http://localhost:8080/API/upload`
   - Body: form-data, key="file", value=filename.csv

**Expected Result:**
```json
{
    "id": "uuid-string"
}
```

**Validation:**
- Status Code: 200 OK
- Response contains valid UUID
- File saved in processed_files directory

---




### Test Case 2: Job Status Progression
**Objective**: Test job status from in-progress to completed
### For testing purpose you can uncomment this line 
time.Sleep(15 * time.Second) 
in file csv_service.go in processFileAsync method

**Steps:**
1. Upload file (from Test Case 1)
2. **Immediate Status Check**
   - Method: `GET`
   - URL: `http://localhost:8080/API/download/{job-id}`

**Expected Result (In Progress):**
```json
{
    "error": "Job is still in progress",
}
```
- Status Code: 423 Locked

3. **Wait 15+ seconds and check again**

**Expected Result (Completed):**
```json
{
    "id": "uuid-string",
    "status": "COMPLETED", 
    "message": "File processed successfully",
    "file_data": blob data,
    "content_type":   "text/csv"
}
```
- Status Code: 200 OK

---

### Test Case 3: File Download
**Objective**: Test processed file download

**Steps:**
1. Use completed job ID from Test Case 2
2. **Download Request**
   - Method: `GET`
   - URL: `http://localhost:8080/API/download/{job-id}`

**Expected Result:**
- Status Code: 200 OK
- Body: CSV data with has_email column




**Validation:**
```csv
name,email,phone,company,has_email
John Doe,john.doe@example.com,123-456-7890,TechCorp,true
Jane Smith,invalid-email,098-765-4321,DesignCo,false
Bob Wilson,bob@company.org,555-123-4567,StartupXYZ,true
Mary Johnson,,999-888-7777,BigCorp,false
Alice Brown,alice.brown@email.net,111-222-3333,,true
Tom Davis,not-an-email-address,444-555-6666,RetailChain,false
```

---

### Test Case 4: Invalid File Type
**Objective**: Test file type validation

**Steps:**
1. **Upload Invalid File**
   - Method: `POST`
   - URL: `http://localhost:8080/API/upload`
   - Body: form-data, key="file", value=test.pdf

**Expected Result:**
```json
{
    "error": "Invalid file type. Only CSV files are allowed"
}
```
- Status Code: 400 Bad Request

---



### Test Case 5: No File Provided
**Objective**: Test missing file parameter

**Steps:**
1. **Upload Without File**
   - Method: `POST`
   - URL: `http://localhost:8080/API/upload`
   - Body: form-data (no file parameter)

**Expected Result:**
```json
{
    "error": "No file provided"
}
```
- Status Code: 400 Bad Request

---




### Test Case 6: Invalid Job ID
**Objective**: Test invalid job ID handling

**Steps:**
1. **Check Status with Invalid ID**
   - Method: `GET`
   - URL: `http://localhost:8080/API/download/invalid-job-id`

**Expected Result:**
```json
{
    "error": "Invalid job ID"
}
```
- Status Code: 400 Bad Request

---




### Test Case 7: Large File Size
**Objective**: Test with large files Size

**Steps:**
1. **Upload Large File**
   - Use test_large.csv more than 10MB
   - Method: `POST`
   - URL: `http://localhost:8080/API/upload`

**Expected Result:**
```json
{
    "error": "file size exceeds 10MB limit"
}
```
- Status Code: 400 Bad Request

---
