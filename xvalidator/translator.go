// @Author beyondyyh@gmail.com
// @Date 2021/08/09 15:49:20
// @Package 参数通用校验器，并把错误信息翻译成中文

package xvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var trans ut.Translator

var (
	ErrValidatorUnsupported  = errors.New("Binding.Validator.Engine() not supported")
	ErrUtilsValidatorNotPtr  = errors.New("Utils validator must be type of pointer")
	ErrTranslatorUnsupported = errors.New("Locale not supported yet, please choose one of")
)

func InitTrans(locale string, obj ...interface{}) (ut.Translator, error) {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return nil, ErrValidatorUnsupported
	}

	var utilsV interface{}
	if len(obj) == 1 && obj[0] != nil {
		utilsV = obj[0]
		typ := reflect.TypeOf(utilsV).Kind()
		if typ != reflect.Ptr && typ != reflect.Struct {
			return nil, ErrUtilsValidatorNotPtr
		}
	}

	// 注册自定义Tag名称：用于替换 field.Name 优先级从高到低依次为：label, form, json, header
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		tags := []string{"label", "form", "json", "header"}
		for _, tag := range tags {
			if label := strings.SplitN(field.Tag.Get(tag), ",", 2)[0]; label != "" {
				return label
			}
		}
		return field.Name
	})

	// 按语言环境注册默认翻译器，英文兜底
	zhT, enT := zh.New(), en.New()
	uni := ut.New(enT, zhT, enT)
	trans, ok = uni.GetTranslator(locale)
	if !ok {
		return nil, fmt.Errorf("%s %s", ErrTranslatorUnsupported, "zh, en")
	}
	switch locale {
	case "en":
		en_translations.RegisterDefaultTranslations(v, trans)
	case "zh":
		zh_translations.RegisterDefaultTranslations(v, trans)
		registerMyTranslations(v, "required_without", "{0}和{1}不能同时为空")
	default:
		en_translations.RegisterDefaultTranslations(v, trans)
	}

	// 注册自定义验证器和翻译器，如果存在同名方法，前者会被覆盖
	registerMyValidations(v, &myValidator{})
	if utilsV != nil {
		registerMyValidations(v, utilsV)
	}

	return trans, nil
}

// registerMyTranslations 注册自定义翻译器，在默认翻译器的基础上进行扩展
// zh只支持了一些比较简单的，高级一点都不支持，如required_without就需要自己实现
// zh支持了哪些tag？see: https://github.com/go-playground/validator/blob/v10.6.1/translations/zh/zh.go
func registerMyTranslations(v *validator.Validate, tag string, txt ...string) {
	text := "{0}不符合规则"
	if len(txt) == 1 && len(txt[0]) > 0 {
		text = txt[0]
	}
	v.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
		return ut.Add(tag, text, false)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		msg, _ := ut.T(fe.Tag(), fe.Field(), fe.Param())
		return msg
	})
}

// registerMyValidations 注册自定义验证器，以`IsValid`开头，遵循下规范：
// 1、业务validator，必须放在`<project>/library/utils`路径下
// 2、公共validator，libs/xvalidator/validator.go文件
// 这要求：1和2命名的方法不能重复，否则2会被1覆盖
// 查找顺序：1 高于 2
func registerMyValidations(v *validator.Validate, myV interface{}) {
	typ := reflect.TypeOf(myV)
	val := reflect.ValueOf(myV)
	for i := 0; i < typ.NumMethod(); i++ {
		tag := typ.Method(i).Name
		if strings.HasPrefix(tag, "IsValid") && val.MethodByName(tag).IsValid() {
			vFunc := val.MethodByName(tag).Interface()
			if fn, ok := vFunc.(func() validator.Func); ok {
				v.RegisterValidation(tag, fn(), true)
			}
			registerMyTranslations(v, tag)
		}
	}
}
