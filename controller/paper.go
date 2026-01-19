package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/model"
)

// API Key Configuration
// TODO: Move to config or environment variable
const SemanticScholarAPIKey = "DQq61qJivBafzkhLNL5m93QDBpV4YRdH4PafBqHV"
const BaseURL = "https://api.semanticscholar.org/graph/v1"

// Global lock and timer for rate limiting (1 request per second)
var (
	requestLock     sync.Mutex
	lastRequestTime time.Time
	minInterval     = 1100 * time.Millisecond // Slightly more than 1s to be safe
)

// rateLimitCall executes the function with rate limiting
func rateLimitCall(f func() (interface{}, error)) (interface{}, error) {
	requestLock.Lock()
	defer requestLock.Unlock()

	now := time.Now()
	elapsed := now.Sub(lastRequestTime)
	if elapsed < minInterval {
		time.Sleep(minInterval - elapsed)
	}

	result, err := f()
	lastRequestTime = time.Now()
	return result, err
}

// PaperSemanticSearch handles semantic search requests
func PaperSemanticSearch(c *gin.Context) {
	var req model.PaperSearchRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Missing required parameter: query",
			"total":  0,
			"cursor": 0,
			"data":   []interface{}{},
		})
		return
	}

	// Safety check for limit
	if req.Limit > 100 {
		req.Limit = 100
	} else if req.Limit < 1 {
		req.Limit = 10
	}

	result, err := rateLimitCall(func() (interface{}, error) {
		// Use standard search endpoint which supports query, offset, limit
		// Note: Semantic Scholar /paper/search endpoint uses 'offset' instead of 'cursor' for non-bulk search
		// Construct URL using url.Values to handle encoding properly
		params := url.Values{}
		params.Add("query", req.Query)
		params.Add("offset", fmt.Sprintf("%d", req.Cursor))
		params.Add("limit", fmt.Sprintf("%d", req.Limit))
		params.Add("fields", "paperId,title,abstract,authors,publicationDate,citationCount,year,url")

		finalURL := fmt.Sprintf("%s/paper/search?%s", BaseURL, params.Encode())

		return doPaperRequest(finalURL)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"total":  0,
			"cursor": 0,
			"data":   []interface{}{},
		})
		return
	}

	resp := result.(*model.SemanticScholarResponse)
	// Fix Next Cursor logic if API returns data
	if resp.Total > 0 && len(resp.Data) > 0 {
		resp.Cursor = req.Cursor + len(resp.Data)
	} else {
		resp.Cursor = req.Cursor
	}

	c.JSON(http.StatusOK, resp)
}

// PaperBooleanSearch handles boolean search requests with year filtering
func PaperBooleanSearch(c *gin.Context) {
	var req model.PaperSearchRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Missing required parameter: query",
			"total":  0,
			"cursor": 0,
			"sort":   "paperId:asc",
			"year":   nil,
			"data":   []interface{}{},
		})
		return
	}

	// Default sort if not provided
	if req.Sort == "" {
		req.Sort = "paperId:asc" // Default from Python code
	}

	// Safety check for limit
	if req.Limit > 100 {
		req.Limit = 100
	} else if req.Limit < 1 {
		req.Limit = 10
	}

	result, err := rateLimitCall(func() (interface{}, error) {
		// Construct URL parameters using url.Values
		params := url.Values{}
		params.Add("query", req.Query)
		params.Add("offset", fmt.Sprintf("%d", req.Cursor))
		params.Add("limit", fmt.Sprintf("%d", req.Limit))
		params.Add("fields", "paperId,title,abstract,authors,publicationDate,citationCount,year,url")

		if req.Sort != "" {
			params.Add("sort", req.Sort)
		}

		if req.Year != "" {
			params.Add("year", req.Year)
		}

		finalURL := fmt.Sprintf("%s/paper/search?%s", BaseURL, params.Encode())

		return doPaperRequest(finalURL)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"total":  0,
			"cursor": 0,
			"sort":   "paperId:asc",
			"year":   nil,
			"data":   []interface{}{},
		})
		return
	}

	resp := result.(*model.SemanticScholarResponse)
	if resp.Total > 0 && len(resp.Data) > 0 {
		resp.Cursor = req.Cursor + len(resp.Data)
	} else {
		resp.Cursor = req.Cursor
	}
	resp.Sort = req.Sort
	resp.Year = req.Year

	c.JSON(http.StatusOK, resp)
}

func doPaperRequest(url string) (*model.SemanticScholarResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", SemanticScholarAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var data model.SemanticScholarResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
