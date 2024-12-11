package request

import (
	"encoding/json"
	"fmt"
	"user/sigmatech/app/service/util"

	"github.com/gin-gonic/gin"
)

type ResetPassword struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PaginationV2 struct {
	Query  string `form:"query" json:"query"`
	Sort   string `form:"sort" json:"sort"`
	Order  string `form:"order" json:"order"`
	Limit  int    `form:"limit,default=10" json:"limit"`
	Offset int    `form:"offset,default=0" json:"offset"`
}

type Pagination struct {
	Limit      *int   `json:"limit,omitempty" form:"limit"`
	Page       *int   `json:"page,omitempty" form:"page"`
	Offset     int    `json:"offset,omitempty" form:"offset"`
	Sort       string `json:"sort,omitempty" form:"sort"`
	Order      string `json:"order,omitempty" form:"order"`
	Query      string `json:"query,omitempty" form:"query"`
	GetAllData bool   `json:"get_all_data,omitempty" form:"get_all_data"`
	Total      int    `json:"total" form:"total"`
	TotalPage  int    `json:"total_page" form:"total_page"`
}

func (r *Pagination) Validate() error {
	if !r.GetAllData {
		if r.Limit == nil || r.Page == nil {
			r.Limit = util.Int(10)
			r.Page = util.Int(1)
		}
		if r.Limit != nil {
			if *r.Limit == 0 {
				r.Limit = util.Int(10)
			}
		}
		if r.Page != nil {
			if *r.Page == 0 {
				r.Page = util.Int(1)
			}
		}
		r.Offset = *r.Limit * (*r.Page - 1)
	} else {
		r.Limit = util.Int(2147483647)
		r.Page = util.Int(1)
	}

	if r.Sort == "" {
		r.Sort = "DESC"
	}
	if r.Order == "" {
		r.Order = "created_at"
	}
	return nil
}

func (r ResetPassword) ValidateRequest() error {
	if r.Email == "" || r.Password == "" {
		return fmt.Errorf("invalid request body, require email/otp/password")
	}
	return nil
}

// ExtractFilteredQueryParams maps query parameters from the URL to a struct or map, excluding pagination fields.
func ExtractFilteredQueryParams(c *gin.Context, filter interface{}) map[string]interface{} {
	// Define the pagination query parameters to exclude
	paginationFields := []string{"limit", "page", "offset", "sort", "order", "query", "get_all_data"}

	// Remove pagination query parameters from the URL
	queryParams := c.Request.URL.Query()
	for _, field := range paginationFields {
		queryParams.Del(field)
	}

	// Create a map to store the extracted query parameters from the filter
	filterMap := make(map[string]interface{})
	filterJSON, _ := json.Marshal(filter)
	json.Unmarshal(filterJSON, &filterMap)

	// Create a result map to store the filtered query parameters
	result := make(map[string]interface{})

	// Iterate over the remaining query parameters
	for key, value := range queryParams {
		// Check if the value is not empty
		if len(value) > 0 && value[0] != "" {
			// Check if the key exists in the filter map
			if _, ok := filterMap[key]; ok {
				// Map the query parameter value to the corresponding key in the result map
				result[key] = value[0]
			}
		}
	}

	return result
}
