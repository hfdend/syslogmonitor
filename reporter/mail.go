package reporter

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"path/filepath"
	"strings"
	"time"
)

type EmailContentType string

type Attachment struct {
	Name string
	Body []byte
}

func SendToMail(user, password, name, addr, to, subject, body string, isHtml bool, attachments ...Attachment) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return err
	}
	// create new SMTP client
	smtpClient, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	auth := smtp.PlainAuth("", user, password, host)
	err = smtpClient.Auth(auth)
	if err != nil {
		return err
	}
	from := mail.Address{Name: name, Address: user}
	if err := smtpClient.Mail(from.Address); err != nil {
		return err
	}
	for _, v := range strings.Split(to, ";") {
		if err := smtpClient.Rcpt(strings.TrimSpace(v)); err != nil {
			return err
		}
	}

	writer, err := smtpClient.Data()
	if err != nil {
		return err
	}
	var contentType string
	if isHtml {
		contentType = "text/html;\r\n\tcharset=utf-8"
	} else {
		contentType = "text/plain;\r\n\tcharset=utf-8"
	}

	boundary := "----THIS_IS_BOUNDARY_JUST_MAKE_YOURS_MIXED"

	buffer := bytes.NewBuffer(nil)

	header := fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: multipart/mixed;\r\n\tBoundary=\"%s\"\r\n"+
		"Mime-Version: 1.0\r\n"+
		"Date: %s\r\n\r\n", to, user, subject, boundary, time.Now().String())
	buffer.WriteString(header)
	buffer.WriteString("This is a multi-part message in MIME format.\r\n\r\n")

	// 正文
	if len(body) > 0 {
		bodyBoundary := "----THIS_IS_BOUNDARY_JUST_MAKE_YOURS_BODY"
		buffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buffer.WriteString(fmt.Sprintf("Content-Type: multipart/alternative;\r\n\tBoundary=\"%s\"\r\n\r\n", bodyBoundary))

		buffer.WriteString(fmt.Sprintf("--%s\r\n", bodyBoundary))
		buffer.WriteString(fmt.Sprintf("Content-Type: %s\r\n", contentType))
		buffer.WriteString(fmt.Sprintf("Content-Transfer-Encoding: base64\r\n\r\n"))
		buffer.WriteString(fmt.Sprintf("%s\r\n\r\n", base64.StdEncoding.EncodeToString([]byte(body))))
		buffer.WriteString(fmt.Sprintf("--%s--\r\n", bodyBoundary))

	}
	for _, attachment := range attachments {
		t := mime.TypeByExtension(filepath.Ext(attachment.Name))
		if t == "" {
			t = "application/octet-stream"
		}
		buffer.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
		buffer.WriteString(fmt.Sprintf("Content-Transfer-Encoding: base64\r\n"))
		buffer.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n\r\n", t, attachment.Name))
		buffer.WriteString(fmt.Sprintf("%s\r\n\r\n", base64.StdEncoding.EncodeToString(attachment.Body)))
	}

	buffer.WriteString("\r\n\r\n--" + boundary + "--")
	_, err = writer.Write(buffer.Bytes())
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	return smtpClient.Quit()
}
