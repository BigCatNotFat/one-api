package model

// PaperSearchRequest corresponds to the parameters used in the Flask app
type PaperSearchRequest struct {
	Query  string `json:"query" form:"query" binding:"required"`
	Limit  int    `json:"limit" form:"limit,default=10"`
	Cursor int    `json:"cursor" form:"cursor,default=0"`
	Sort   string `json:"sort" form:"sort"` // e.g., "publicationDate:desc"
	Year   string `json:"year" form:"year"` // e.g., "2023", "2019-2023", "2019-", "-2023"
}

// SemanticScholarResponse corresponds to the API response structure
type SemanticScholarResponse struct {
	Total     int     `json:"total"`
	Token     *string `json:"token,omitempty"`
	NextToken *string `json:"next_token,omitempty"`
	Data      []Paper `json:"data"`
	Error     string  `json:"error,omitempty"`
	Cursor    int     `json:"cursor,omitempty"` // Added to match Python response
	Sort      string  `json:"sort,omitempty"`   // Added to match Python response
	Year      string  `json:"year,omitempty"`   // Added to match Python response
}

type Paper struct {
	PaperId         string   `json:"paperId"`
	Title           string   `json:"title"`
	Abstract        string   `json:"abstract"`
	PublicationDate string   `json:"publicationDate"`
	CitationCount   int      `json:"citationCount"`
	Authors         []Author `json:"authors"`
	Year            int      `json:"year"`
	Url             string   `json:"url"`
}

type Author struct {
	AuthorId string `json:"authorId"`
	Name     string `json:"name"`
}

