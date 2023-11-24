package presenters

import "encoding/xml"

type Error struct {
	XMLName xml.Name `json:"-" xml:"error"`
	Error   []string `json:"error" xml:"error"`
}

func TransformErrorToApiError(err ...error) *Error {
	var errs []string
	for _, e := range err {
		errs = append(errs, e.Error())
	}

	return &Error{
		Error: errs,
	}
}
