package conf

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Model struct {
	Mail struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
		Host     string `yaml:"host"`
		To       string `yaml:"to"`
		Subject  string `yaml:"subject"`
	} `yaml:"mail"`
	ES struct {
		Addr     string `yaml:"addr"`
		User     string `yaml:"user"`
		Index    string `yaml:"index"`
		Password string `yaml:"password"`
	} `yaml:"es"`
	SMS struct {
		Send          bool     `yaml:"send"`
		AppKey        string   `yaml:"app_key"`
		LogTemplateId string   `yaml:"log_template_id"`
		Phones        []string `yaml:"phones"`
	} `yaml:"sms"`
	CheckServices []struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
		Name string `yaml:"name"`
	} `yaml:"check_services"`
}

var (
	Config Model
)

func Init() {
	var (
		configFile string
		configBts  []byte
		err        error
	)
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	flag.StringVar(&configFile, "f", "", "config file")
	flag.Parse()

	if configFile == "" {
		if configBts, err = LRead("config.yml", 2); err != nil {
			log.Fatalln(err)
		}
	} else {
		if configBts, err = ioutil.ReadFile(configFile); err != nil {
			log.Fatalln(err)
		}
	}
	if err = yaml.Unmarshal(configBts, &Config); err != nil {
		log.Fatalln(err)
	}
}

// LRead 向上读取文件
func LRead(name string, level int) (raw []byte, err error) {
	var file *os.File
	for i := 0; i <= level; i++ {
		filePath := fmt.Sprintf("%s%s", strings.Repeat("../", i), name)
		file, err = os.OpenFile(filePath, os.O_RDONLY, 0600)
		if err != nil {
			continue
		} else {
			break
		}
	}
	if err != nil {
		return
	}
	raw, err = ioutil.ReadAll(file)
	return
}
