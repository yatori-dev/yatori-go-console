package utils

import (
	"crypto/tls"
	"fmt"
	"html"
	"strings"

	"gopkg.in/gomail.v2"
)

// SendMail 发送邮件
func SendMail(host string, port int, userName, password string, toMail []string, content string) {

	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(userName, "Yatori课程助手")) // 发件人

	m.SetHeader("To", toMail...) // 收件人，可以多个收件人，但必须使用相同的 SMTP 连接

	m.SetHeader("Subject", "Yatori课程助手通知") // 邮件主题

	// 可以通过 text/html 处理文本格式进行特殊处理，如换行、缩进、加粗等等
	emailHTML := buildEmailHTML("Yatori课程助手", content, false)
	m.SetBody("text/html", emailHTML)

	// text/plain的意思是将文件设置为纯文本的形式，浏览器在获取到这种文件时并不会对其进行处理
	// m.SetBody("text/plain", "纯文本")
	// m.Attach("test.sh")   // 附件文件，可以是文件，照片，视频等等
	// m.Attach("lolcatVideo.mp4") // 视频
	// m.Attach("lolcat.jpg") // 照片

	d := gomail.NewDialer(
		host,
		port,
		userName,
		password,
	)
	// 关闭SSL协议认证
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

// 生成邮件 HTML：title 是邮件标题展示，contentHTML 是正文（可以传纯文本或 HTML 片段）
func buildEmailHTML(title string, contentHTML string, asPlainText bool) string {
	logoURL := "https://avatars.githubusercontent.com/u/185567923?s=1000&v=4"

	// 如果传的是纯文本，则做一次转义并换行转 <br>
	if asPlainText {
		contentHTML = html.EscapeString(contentHTML)
		contentHTML = strings.ReplaceAll(contentHTML, "\n", "<br>")
	}

	// 使用 table + 内联样式，提升邮箱客户端兼容性
	return fmt.Sprintf(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="x-apple-disable-message-reformatting">
  <meta name="color-scheme" content="light dark">
  <meta name="supported-color-schemes" content="light dark">
  <title>%s</title>
</head>
<body style="margin:0;padding:0;background:#f5f7fb;">
  <table role="presentation" cellpadding="0" cellspacing="0" width="100%%" style="background:#f5f7fb;">
    <tr>
      <td align="center" style="padding:32px 16px;">
        <table role="presentation" cellpadding="0" cellspacing="0" width="600" style="max-width:600px;background:#ffffff;border-radius:16px;box-shadow:0 6px 24px rgba(18,38,63,0.08);">
          <tr>
            <td align="center" style="padding:28px 24px 8px 24px;">
              <img src="%s" width="88" height="88" alt="logo" 
                   style="display:block;border-radius:50%%;width:88px;height:88px;border:2px solid #eef2f7;object-fit:cover;">
            </td>
          </tr>
          <tr>
            <td align="center" style="padding:0 24px 8px 24px;">
              <div style="font-family:system-ui,-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica,Arial,'Noto Sans',sans-serif;
                          font-size:22px;font-weight:700;color:#111827;line-height:1.3;">%s</div>
            </td>
          </tr>
          <tr>
            <td style="padding:8px 24px 0 24px;">
              <div style="height:1px;background:linear-gradient(90deg,#e5e7eb,#f3f4f6,#e5e7eb);"></div>
            </td>
          </tr>
          <tr>
            <td style="padding:18px 24px 8px 24px;">
              <div style="font-family:system-ui,-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica,Arial,'Noto Sans',sans-serif;
                          font-size:15px;color:#374151;line-height:1.8;">
                %s
              </div>
            </td>
          </tr>
          <tr>
            <!--<td style="padding:8px 24px 28px 24px;" align="center">
              <a href="#" 
                 style="display:inline-block;padding:10px 18px;border-radius:999px;text-decoration:none;
                        font-family:system-ui,-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica,Arial,'Noto Sans',sans-serif;
                        font-size:14px;font-weight:600;background:#111827;color:#ffffff;">
                查看详情
              </a>
            </td>-->
          </tr>
        </table>

        <table role="presentation" cellpadding="0" cellspacing="0" width="600" style="max-width:600px;">
          <tr><td align="center" style="padding:14px 8px 0 8px;color:#6b7280;
              font-family:system-ui,-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica,Arial,'Noto Sans',sans-serif;
              font-size:12px;line-height:1.6;">
            这是一封系统通知邮件，请勿直接回复。
          </td></tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`,
		html.EscapeString(title), // <title> 内建议转义
		logoURL,
		html.EscapeString(title),
		contentHTML,
	)
}
