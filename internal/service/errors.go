package service

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrConflict        = errors.New("conflict")
	ErrInvalidInput    = errors.New("invalid input")
	ErrInactiveUser    = errors.New("account inactive")
	ErrTOTPRequired    = errors.New("totp required")
	ErrInvalidTOTP     = errors.New("invalid totp code")
	ErrWallpaperLimit  = errors.New("wallpaper limit reached")
	ErrUnsupportedFile = errors.New("unsupported file type")
)
