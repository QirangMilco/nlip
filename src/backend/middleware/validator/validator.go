package validator

import (
	"fmt"
	"nlip/utils/errors"
	"nlip/utils/logger"
	fileValidator "nlip/utils/validator"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

func init() {
	// 注册自定义验证器
	if err := validate.RegisterValidation("filename", validateFileName); err != nil {
		logger.Error("注册filename验证器失败: %v", err)
	}
	if err := validate.RegisterValidation("filetype", validateFileType); err != nil {
		logger.Error("注册filetype验证器失败: %v", err)
	}
	if err := validate.RegisterValidation("filemimetype", validateFileMimeType); err != nil {
		logger.Error("注册filemimetype验证器失败: %v", err)
	}
	if err := validate.RegisterValidation("fileext", validateFileExtension); err != nil {
		logger.Error("注册fileext验证器失败: %v", err)
	}

	logger.Info("验证器初始化完成")
}

// ValidateBody 验证请求体
func ValidateBody(payload interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger.Debug("Content-Type: %s", c.Get("Content-Type"))
		logger.Debug("Raw Body: %s", string(c.Body()))
		
		// 创建一个新的payload实例
		p := payload

		// 解析请求体
		if err := c.BodyParser(p); err != nil {
			logger.Warning("请求体解析失败: %+v, 请求体内容: %s", err, string(c.Body()))
			return errors.NewNlipError(errors.ErrBadRequest, "无效的请求数据")
		}

		// 验证数据
		if err := validate.Struct(p); err != nil {
			// 处理验证错误
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				errorMessages := make([]string, 0)
				for _, e := range validationErrors {
					errorMessages = append(errorMessages, formatValidationError(e))
				}
				logger.Warning("数据验证失败: %s", strings.Join(errorMessages, "; "))
				return errors.NewNlipError(errors.ErrBadRequest, strings.Join(errorMessages, "; "))
			}
			logger.Warning("数据验证失败: %v", err)
			return errors.NewNlipError(errors.ErrBadRequest, "数据验证失败")
		}

		logger.Debug("请求数据验证通过")
		// 将验证后的数据存储在上下文中
		c.Locals("validatedBody", p)
		return c.Next()
	}
}

// validateFileName 验证文件名
func validateFileName(fl validator.FieldLevel) bool {
	filename := fl.Field().String()
	return fileValidator.ValidateFileName(filename)
}

// validateFileType 验证文件类型（包括扩展名和MIME类型）
func validateFileType(fl validator.FieldLevel) bool {
	filename := fl.Field().String()
	contentType := fl.Param()
	return fileValidator.ValidateFileType(filename, contentType)
}

// validateFileMimeType 验证MIME类型
func validateFileMimeType(fl validator.FieldLevel) bool {
	mimeType := fl.Field().String()
	return fileValidator.IsAllowedMimeType(mimeType)
}

// validateFileExtension 验证文件扩展名
func validateFileExtension(fl validator.FieldLevel) bool {
	filename := fl.Field().String()
	return fileValidator.IsAllowedExtension(filename)
}

// formatValidationError 格式化验证错误信息
func formatValidationError(e validator.FieldError) string {
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
	case "filemimetype":
		return fmt.Sprintf("%s的MIME类型不支持", field)
	case "fileext":
		return fmt.Sprintf("%s的扩展名不支持", field)
	default:
		return fmt.Sprintf("%s验证失败: %s", field, tag)
	}
}
