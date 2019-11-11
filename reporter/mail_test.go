package reporter

import (
	"fmt"
	"testing"
)

func TestEmail_Send(t *testing.T) {
	user := "account"
	password := "password"
	name := "test mail"
	host := "smtp.gmail.com:465"
	to := "sun75626253@outlook.com"
	subject := "测试邮箱"
	body := `<span style="color:red">this is a test</span>`
	err := SendToMail(user, password, name, host, to, subject, body, true)
	fmt.Println(err)
}
