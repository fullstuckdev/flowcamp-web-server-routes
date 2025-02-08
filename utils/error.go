package utils

import "errors"

var (
	ErrNotFound = errors.New("record not found")
)

type ErrorResponse struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

// func GetErrorResp(err error) ErrorResponse {
// 	switch {
// 	case errors.Is(err, ErrNotFound):
// 		return ErrorResponse{
// 			Code: 404,
// 			Message: "Data tidak ditemukan",
// 		}
// 	}
	
// }