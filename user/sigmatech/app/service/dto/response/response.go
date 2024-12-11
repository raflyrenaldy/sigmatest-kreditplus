package response

import "time"

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

// ErrorResponseData -
type ErrorResponseData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse -
type ErrorResponse struct {
	Success bool              `json:"success" default:"false"`
	Error   ErrorResponseData `json:"data"`
}

type Response struct {
	Success bool `json:"success"`
	Data    Data `json:"data"`
}

type ResponseV2 struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Request interface{} `json:"request,omitempty"`
	// List    interface{} `json:"list,omitempty"`
}

type ResponseV3 struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Request interface{} `json:"request,omitempty"`
}

type Pagination struct {
	Page       int `json:"page" default:"1"`
	PerPage    int `json:"per_page" default:"10"`
	TotalPages int `json:"total_pages"`
	TotalCount int `json:"total_count"`
}
type Data struct {
	Message string `json:"message"`
}

type ProductCategory struct {
	Id                int                 `json:"id"`
	Name              string              `json:"name"`
	Url               string              `json:"url"`
	CreatedAt         time.Time           `json:"created_at"`
	CreatedBy         int                 `json:"created_by"`
	UpdatedAt         time.Time           `json:"updated_at"`
	UpdatedBy         *int                `json:"updated_by"`
	SubCategoryLevel1 []SubCategoryLevel1 `json:"sub_category_level_1,omitempty"`
}
type SubCategoryLevel1 struct {
	Id                int        `json:"id"`
	Name              string     `json:"name"`
	Url               string     `json:"url"`
	CreatedAt         time.Time  `json:"created_at"`
	CreatedBy         int        `json:"created_by"`
	UpdatedAt         time.Time  `json:"updated_at"`
	UpdatedBy         *int       `json:"updated_by"`
	SubCategoryLevel2 []Category `json:"sub_category_level_2,omitempty"`
}

type Category struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Url       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy int       `json:"created_by"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy *int      `json:"updated_by"`
}

type UpdateCategory struct {
	MainCategoryId    *int     `json:"main_category_id"`
	SubCategoryLevel1 *int     `json:"sub_category_level1_id"`
	SubCategoryLevel2 *int     `json:"sub_category_level2_id"`
	Category          Category `json:"category"`
}

type ProductMembershipPrice struct {
	Id        int       `json:"id"`
	VendorId  int       `json:"vendor_id"`
	Name      string    `json:"name"`
	Points    int       `json:"points"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy *int      `json:"created_by"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy *int      `json:"updated_by"`
}
