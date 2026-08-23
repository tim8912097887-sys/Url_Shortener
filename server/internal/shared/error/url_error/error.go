package urlerror

import "errors"

var (
	ErrUrlNotFound = errors.New("url not found")
	ErrShortURLCollision = errors.New("short url collision")
)