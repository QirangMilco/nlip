package validator

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"nlip/utils/logger"
	"reflect"
	"strings"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// 注册自定义验证器
	validate.RegisterValidation("filename", validateFileName)
	validate.RegisterValidation("filetype", validateFileType)
}

// Validate 验证结构体
func Validate(data interface{}) error {
	if err := validate.Struct(data); err != nil {
		// 添加详细的验证错误信息日志
		if _, ok := err.(*validator.InvalidValidationError); ok {
			logger.Error("验证器错误: %v", err)
			return err
		}

		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			logger.Warning("字段验证失败: Field=%s, Tag=%s, Value=%v", 
				err.Field(), err.Tag(), err.Value())
			errors = append(errors, fmt.Sprintf(
				"字段 %s 验证失败: %s", err.Field(), err.Tag()))
		}
		
		return fmt.Errorf("验证失败: %v", strings.Join(errors, "; "))
	}
	return nil
}

// formatError 格式化验证错误信息
func formatError(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	// 获取字段的json标签
	if t := reflect.TypeOf(e.Value()); t != nil {
		if f, ok := t.FieldByName(field); ok {
			if tag := f.Tag.Get("json"); tag != "" {
				field = strings.Split(tag, ",")[0]
			}
		}
	}

	switch tag {
	case "required":
		return fmt.Sprintf("%s是必需的", field)
	case "min":
		return fmt.Sprintf("%s不能小于%s", field, param)
	case "max":
		return fmt.Sprintf("%s不能大于%s", field, param)
	case "oneof":
		return fmt.Sprintf("%s必须是以下值之一: %s", field, param)
	case "filename":
		return fmt.Sprintf("%s包含非法字符", field)
	case "filetype":
		return fmt.Sprintf("%s类型不支持", field)
	default:
		return fmt.Sprintf("%s验证失败: %s", field, tag)
	}
}

// validateFileName 验证文件名是否合法
func validateFileName(fl validator.FieldLevel) bool {
	filename := fl.Field().String()
	// 禁止包含特殊字符和路径分隔符
	invalidChars := `<>:"/\|?*`
	return !strings.ContainsAny(filename, invalidChars)
}

// validateFileType 验证文件类型是否允许
func validateFileType(fl validator.FieldLevel) bool {
	fileType := fl.Field().String()
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif", "text/plain", "application/pdf"}
	for _, allowed := range allowedTypes {
		if fileType == allowed {
			return true
		}
	}
	return false
} 