package validate

import (
	"github.com/go-playground/validator/v10"
)

var I *validator.Validate

func init() {
	I = validator.New()
}
