package message

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"syslogmonitor/conf"
	"syslogmonitor/reporter"
)

type Pool struct {
	list             [][]byte
	mutex            sync.Mutex
	lastSendMailTime time.Time
	lastSendSMSTime  time.Time
}

func (m *Pool) Push(data []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.list = append(m.list, data)
}

func (m *Pool) Length() int {
	return len(m.list)
}

func (m *Pool) PopAll() [][]byte {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	list := m.list
	m.list = [][]byte{}
	return list
}

func (m *Pool) Monitor() {
	for {
		n := m.Length()
		if n > 0 {
			go m.SendSMS(n)
			go m.SendMail()
		}
		time.Sleep(time.Second)
	}
}

func (m *Pool) SendSMS(n int) {
	if !conf.Config.SMS.Send {
		return
	}
	if time.Now().Sub(m.lastSendSMSTime) < 15*time.Minute {
		return
	}
	sms := reporter.SMS{
		AppKey: conf.Config.SMS.AppKey,
	}
	data := map[string]interface{}{}
	data["number"] = n
	data["time"] = time.Now().Format("2006-01-02 15:04:05")
	for _, v := range conf.Config.SMS.Phones {
		go sms.SMSSend(v, conf.Config.SMS.LogTemplateId, data)
	}
	m.lastSendSMSTime = time.Now()
}

func (m *Pool) SendMail() {
	// 一分钟发一次
	if time.Now().Sub(m.lastSendMailTime) < time.Minute {
		return
	}
	hostname, _ := os.Hostname()
	list := m.PopAll()
	if len(list) > 0 {
		var body string
		body += "<div><ul>"
		count := len(list)
		if count > 20 {
			list = list[0:20]
		}
		for _, v := range list {
			s := string(v)
			s = strings.Replace(s, "@|@", "<br />", -1)
			s = strings.Replace(s, "\n", "<br />", -1)
			s = strings.Replace(s, "\t", "&nbsp;&nbsp;&nbsp;&nbsp;", -1)
			body += fmt.Sprintf(`<li style="list-style: decimal;"><div style="padding-left: 10px;"><pre>%s</pre></div></li>`, s)
		}
		body += "</ul></div>"
		body += fmt.Sprintf("<div>共计%d条错误信息</div>", count)
		go func(body string) {
			if conf.Config.Mail.Host != "" {
				if err := reporter.SendToMail(
					conf.Config.Mail.User,
					conf.Config.Mail.Password,
					conf.Config.Mail.Name,
					conf.Config.Mail.Host,
					conf.Config.Mail.To,
					fmt.Sprintf("%s: %s", conf.Config.Mail.Subject, hostname),
					body,
					true,
				); err != nil {
					log.Println(err)
				}
			}

		}(body)
		m.lastSendMailTime = time.Now()
	}
}
