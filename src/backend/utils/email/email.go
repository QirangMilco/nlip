package email

import (
	"fmt"
	"net/smtp"
	"nlip/config"
	"nlip/utils/logger"
)

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

var emailConfig EmailConfig

// InitEmailConfig 初始化邮件配置
func InitEmailConfig() {
	emailConfig = EmailConfig{
		Host:     config.AppConfig.Email.Host,
		Port:     config.AppConfig.Email.Port,
		Username: config.AppConfig.Email.Username,
		Password: config.AppConfig.Email.Password,
		From:     config.AppConfig.Email.From,
	}
}

// SendInviteEmail 发送邀请邮件
func SendInviteEmail(toEmail, spaceName, inviteLink string) error {
	subject := fmt.Sprintf("邀请您加入空间：%s", spaceName)
	body := fmt.Sprintf(`
        <h2>空间协作邀请</h2>
        <p>您好，</p>
        <p>您被邀请加入空间：%s</p>
        <p>请点击以下链接接受邀请：</p>
        <p><a href="%s">%s</a></p>
        <p>此链接24小时内有效。</p>
        <p>如果您没有要求此邀请，请忽略此邮件。</p>
    `, spaceName, inviteLink, inviteLink)

	return sendEmail(toEmail, subject, body)
}

// sendEmail 发送邮件的通用方法
func sendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, emailConfig.Host)

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n%s\r\n%s",
		to, emailConfig.From, subject, mime, body)

	addr := fmt.Sprintf("%s:%d", emailConfig.Host, emailConfig.Port)
	err := smtp.SendMail(addr, auth, emailConfig.From, []string{to}, []byte(msg))
	if err != nil {
		logger.Error("发送邮件失败: %v", err)
		return err
	}

	logger.Info("成功发送邮件至: %s", to)
	return nil
}
