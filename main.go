package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"syslogmonitor/cli"
	"syslogmonitor/conf"
	"syslogmonitor/healthy"
	"syslogmonitor/message"
	"time"

	"github.com/olivere/elastic/v7"
)

var (
	messagePool *message.Pool
)

func main() {
	cli.Init()
	messagePool = new(message.Pool)
	go messagePool.Monitor()
	go checkHealthy()
	go pullES()
	select {}
}

func pullES() {
	if conf.Config.ES.Addr == "" {
		return
	}
	var opt []elastic.ClientOptionFunc
	opt = append(opt, elastic.SetURL(conf.Config.ES.Addr), elastic.SetSniff(false))
	if conf.Config.ES.User != "" {
		opt = append(opt, elastic.SetBasicAuth(conf.Config.ES.User, conf.Config.ES.Password))
	}
	es, err := elastic.NewClient(opt...)
	if err != nil {
		messagePool.Push([]byte(fmt.Sprintf("connect elasticsearch error: %s addr: %s, user: %s, password: %s", err, conf.Config.ES.Addr, conf.Config.ES.User, conf.Config.ES.Password)))
		return
	}
	size := 5
	idMap := map[string]interface{}{}
	for {
		boolQuery := elastic.NewBoolQuery()
		boolQuery.Must(
			elastic.NewMatchQuery("level", "error"),
			elastic.NewRangeQuery("@timestamp").Gt(time.Now().Add(-5*time.Minute)),
		)
		ctx, _ := context.WithTimeout(context.Background(), time.Minute)
		result, err := es.Search(conf.Config.ES.Index).Size(size).Query(boolQuery).Sort("@timestamp", false).Do(ctx)
		if err != nil {
			log.Fatalln(err)
		}
		for _, v := range result.Hits.Hits {
			if _, ok := idMap[v.Id]; !ok {
				idMap[v.Id] = nil
				info, _ := json.MarshalIndent(v.Source, "", "    ")
				messagePool.Push(info)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func checkHealthy() {
	for _, v := range conf.Config.CheckServices {
		go func(host, port, name string) {
			for {
				ok := healthy.TcpCheck(host, port, time.Second)
				if !ok {
					msg := []byte(fmt.Sprintf("service: %s, address: %s:%s dial failed", name, host, port))
					messagePool.Push(msg)
				}
				time.Sleep(time.Second)
			}
		}(v.Host, v.Port, v.Name)
	}
}
