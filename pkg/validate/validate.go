package validate

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = validator.New()

func init() {
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func Struct(i any) InvalidParamsErr {
	if err := v.Struct(i); err != nil {
		var params []InvalidParamErr
		for _, e := range err.(validator.ValidationErrors) {
			params = append(params, InvalidParamErr{
				ParamName: e.Field(),
				Message:   message(e),
			})
		}
		return params
	}
	return nil
}

func Var(field any, tag, fieldName string) InvalidParamsErr {
	if err := v.Var(field, tag); err != nil {
		var params []InvalidParamErr
		for _, e := range err.(validator.ValidationErrors) {
			params = append(params, InvalidParamErr{
				ParamName: fieldName,
				Message:   message(e),
			})
		}
		return params
	}
	return nil
}
