package presenters

import "encoding/xml"

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type ApiDataFormat struct {
	XMLName    xml.Name    `json:"-" xml:"data"`
	Data       interface{} `json:"data" xml:"data"`
	Pagination *Pagination `json:"pagination,omitempty" xml:"pagination,omitempty"`
}

func TransformDataToApiFormat(data interface{}) *ApiDataFormat {
	return &ApiDataFormat{
		Data: data,
	}
}

func (format *ApiDataFormat) WithPagination(page, pageSize int) *ApiDataFormat {
	format.Pagination = &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
	return format
}
