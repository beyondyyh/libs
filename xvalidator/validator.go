// @Author beyondyyh@gmail.com
// @Date 2021/08/09 15:49:20
// @Package 参数通用校验器，公共validator

package xvalidator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

type myValidator struct{}

var (
	phoneTpl string         = `^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\d{8}$`
	phoneReg *regexp.Regexp = regexp.MustCompile(phoneTpl)
)

func (v *myValidator) IsValidPhone() validator.Func {
	return func(fl validator.FieldLevel) bool {
		return phoneReg.MatchString(fl.Field().String())
	}
}
