package models

type HttpError struct {
	Err  string
	Code int
}

func HttpError_From(err string, code int) *HttpError {
	return &HttpError{
		Err:  err,
		Code: code,
	}
}
