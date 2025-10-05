package validate

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type InvalidParamErr struct {
	ParamName string `json:"param_name"`
	Message   string `json:"message"`
}

func (ipe InvalidParamErr) Error() string {
	return ipe.Message
}

type InvalidParamsErr []InvalidParamErr

func (ipse InvalidParamsErr) Error() string {
	return "invalid params"
}

func message(e validator.FieldError) string {
	name := e.Field()
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s param is required", name)
	case "uuid":
		return fmt.Sprintf("%s param must be a valid UUID", name)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal %s", name, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal %s", name, e.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", name, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", name, e.Param())
	}

	return e.Error()
}
