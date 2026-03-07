package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaginationRequest struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type PaginationMeta struct {
	TotalData   int64 `json:"totalData"`
	TotalPage   int   `json:"totalPage"`
	CurrentPage int   `json:"currentPage"`
	Limit       int   `json:"limit"`
}

type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

func GetPaginationRequest(c *gin.Context) PaginationRequest {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	return PaginationRequest{
		Page:  page,
		Limit: limit,
	}
}

func (p PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

func CreatePaginationMeta(totalData int64, page, limit int) PaginationMeta {
	totalPage := int(math.Ceil(float64(totalData) / float64(limit)))
	if totalPage == 0 {
		totalPage = 1
	}

	return PaginationMeta{
		TotalData:   totalData,
		TotalPage:   totalPage,
		CurrentPage: page,
		Limit:       limit,
	}
}
