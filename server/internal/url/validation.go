package url

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type CreateUrlSchema struct {
	Url string `json:"url" validate:"required,url"`
}

type GetUrlParams struct {
	ShortURL string `validate:"required,alphanum,min=8,max=8"`
}

func BindAndValidate[T any](ctx fiber.Ctx) (T, error) {
	var input T

	// Check if the request body is missing entirely
	if ctx.Request().Body() == nil || ctx.Request().Header.ContentLength() == 0 {
        return input, errors.New("request body is empty")
    }

	// Bind the request body
	if err := ctx.Bind().Body(&input); err != nil {
		return input, err
	}

	if err := validate.Struct(input); err != nil {
        return input, err
    }

    return input, nil
}

var validate = validator.New()

func Validate[T any](input T) (T, error) {
	if err := validate.Struct(input); err != nil {
		return input, err
	}
	return input, nil
}