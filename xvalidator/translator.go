// @Author beyondyyh@gmail.com
// @Date 2021/08/09 15:49:20
// @Package 参数通用校验器，并把错误信息翻译成中文

package xvalidator

import (
	"context"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"

	"github.com/beyondyyh/libs/set"
)

type MyTranslator struct {
	zhT     ut.Translator // 中文翻译器
	zhTOnce sync.Once     // 单例
}

// lazyInit, init once
func (t *MyTranslator) lazyInit() {
	t.zhTOnce.Do(func() {
		t.zhT, _ = ut.New(zh.New()).GetTranslator("zh")
	})
}

// Get 获取注册的翻译器
func (t *MyTranslator) Get() ut.Translator {
	t.lazyInit()

	return t.zhT
}

func (t *MyTranslator) RegisterWithCtx(ctx context.Context, paramsPtr interface{}, busCommonPtr ...interface{}) {
	var commonPtr interface{} = reflect.Ptr
	if len(busCommonPtr) > 0 && busCommonPtr[0] != nil {
		commonPtr = busCommonPtr[0]
	}
	t.register(ctx, paramsPtr, commonPtr)
}

func (t *MyTranslator) Register(paramsPtr interface{}, busCommonPtr ...interface{}) {
	t.RegisterWithCtx(nil, paramsPtr, busCommonPtr...)
}

// register 各种注册
// - paramsPtr 	业务参数结构体，注意是指针
// - commonPtr	业务通用验证器，注意是指针
func (t *MyTranslator) register(ctx context.Context, paramsPtr interface{}, commonPtr interface{}) {
	t.lazyInit()

	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	// 注册默认翻译器，zh只支持了一些比较简单的，高级一点都不支持如required_without，需要自己实现
	// zh支持了哪些tag？see: https://github.com/go-playground/validator/blob/v10.6.1/translations/zh/zh.go
	zh_translations.RegisterDefaultTranslations(v, t.zhT)

	// 注册自定义翻译器
	registerCustomTranslations := func(tag string, txt ...string) {
		text := "{0}不合法"
		if len(txt) == 1 && len(txt[0]) > 0 {
			text = txt[0]
		}
		v.RegisterTranslation(tag, t.zhT, func(ut ut.Translator) error {
			return ut.Add(tag, text, false)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			msg, _ := ut.T(fe.Tag(), fe.Field(), fe.Param())
			return msg
		})
	}

	// 注册自定义翻译示例，优雅一点的写法是单独抽象出来，针对每种case单独提示错误信息
	registerCustomTranslations("required_without", "{0}和{1}不能同时为空")

	// 注册自定义Tag名称，用于替换 field.Name
	// 优先级从高到低：label, form, json, header
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		tags := []string{"label", "form", "json", "header"}
		for _, tag := range tags {
			if label := strings.SplitN(field.Tag.Get(tag), ",", 2)[0]; label != "" {
				return label
			}
		}
		return field.Name
	})

	// 注册自定义校验函数 和 翻译
	cs := set.NewSet()
	rt := reflect.TypeOf(paramsPtr).Elem() // ptr
	for i := 0; i < rt.NumField(); i++ {
		if tags, ok := rt.Field(i).Tag.Lookup("binding"); ok {
			for _, tag := range strings.Split(tags, ",") {
				if strings.HasPrefix(tag, "IsValid") {
					cs.Add(tag)
				}
			}
		}
	}
	if cs.Len() == 0 {
		return
	}

	// 优先级从高到低：paramsV, 业务通用commonV, libV
	// Switch 会保证从上到下、从左到右的执行顺序
	paramsV := reflect.ValueOf(paramsPtr)    // 绑定在params上的validator
	commonV := reflect.ValueOf(commonPtr)    // 业务通用validator，一般是在library/utils pkg下
	libsV := reflect.ValueOf(&myValidator{}) // 公共libs的validator
	for _, ele := range cs.Elements() {
		tag := ele.(string)
		var vFunc interface{}
		switch {
		case paramsV.MethodByName(tag).IsValid():
			vFunc = paramsV.MethodByName(tag).Interface()
		case commonV.MethodByName(tag).IsValid():
			vFunc = commonV.MethodByName(tag).Interface()
		case libsV.MethodByName(tag).IsValid():
			vFunc = libsV.MethodByName(tag).Interface()
		}

		if fn, ok := vFunc.(func() validator.Func); ok {
			v.RegisterValidation(tag, fn())
		}
		if fn, ok := vFunc.(func(context.Context) validator.Func); ok && ctx != nil {
			v.RegisterValidation(tag, fn(ctx))
		}

		registerCustomTranslations(tag)
	}
}
