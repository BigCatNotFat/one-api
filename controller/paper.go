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
	processingQueue = make(chan struct{}, 30)
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
	// Check queue
	select {
	case processingQueue <- struct{}{}:
		defer func() { <-processingQueue }()
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":  "当前使用量过大，请稍后重试",
			"total":  0,
			"cursor": 0,
			"data":   []interface{}{},
		})
		return
	}

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
	// Check queue
	select {
	case processingQueue <- struct{}{}:
		defer func() { <-processingQueue }()
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":  "当前使用量过大，请稍后重试",
			"total":  0,
			"cursor": 0,
			"sort":   "paperId:asc",
			"year":   nil,
			"data":   []interface{}{},
		})
		return
	}

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
		req.Sort = "paperId:asc"
	}

	// Safety check for limit
	if req.Limit > 100 {
		req.Limit = 100
	} else if req.Limit < 1 {
		req.Limit = 10
	}

	// We need to fetch items to satisfy [req.Cursor, req.Cursor+req.Limit]
	// Since we use Bulk API, we must iterate from the beginning.

	var accumulatedPapers []model.Paper
	var total int
	var nextToken string

	neededCount := req.Cursor + req.Limit

	// Loop to fetch pages until we have enough data
	for {
		// Stop if we have enough papers
		if len(accumulatedPapers) >= neededCount {
			break
		}

		result, err := rateLimitCall(func() (interface{}, error) {
			params := url.Values{}
			params.Add("query", req.Query)
			params.Add("fields", "paperId,title,abstract,authors,publicationDate,citationCount,year,url")
			params.Add("sort", req.Sort)

			if req.Year != "" {
				params.Add("year", req.Year)
			}

			if nextToken != "" {
				params.Add("token", nextToken)
			}

			// Use Bulk Search Endpoint
			// Note: Bulk search does not support 'limit' or 'offset' parameters directly like standard search
			finalURL := fmt.Sprintf("%s/paper/search/bulk?%s", BaseURL, params.Encode())
			return doPaperRequest(finalURL)
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  err.Error(),
				"total":  total,
				"cursor": req.Cursor,
				"sort":   req.Sort,
				"year":   req.Year,
				"data":   []interface{}{},
			})
			return
		}

		resp := result.(*model.SemanticScholarResponse)

		// Update total (Bulk API returns total matches)
		if resp.Total > 0 {
			total = resp.Total
		}

		// Append papers
		if len(resp.Data) > 0 {
			accumulatedPapers = append(accumulatedPapers, resp.Data...)
		}

		// Check for next token
		if resp.Token != nil && *resp.Token != "" {
			nextToken = *resp.Token
		} else if resp.NextToken != nil && *resp.NextToken != "" {
			nextToken = *resp.NextToken
		} else {
			// No more pages
			break
		}
	}

	// Slice the results to match requested cursor and limit
	var resultData []model.Paper
	var nextCursor int

	if req.Cursor < len(accumulatedPapers) {
		end := req.Cursor + req.Limit
		if end > len(accumulatedPapers) {
			end = len(accumulatedPapers)
		}
		resultData = accumulatedPapers[req.Cursor:end]
	} else {
		resultData = []model.Paper{}
	}

	nextCursor = req.Cursor + len(resultData)

	c.JSON(http.StatusOK, gin.H{
		"error":  nil,
		"total":  total,
		"cursor": nextCursor,
		"sort":   req.Sort,
		"year":   req.Year,
		"data":   resultData,
	})
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
