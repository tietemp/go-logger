package logger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"strconv"
	"strings"
	"sync"
	"time"
)

type elasticLogger struct {
	Addr     string `json:"addr"`
	Index    string `json:"index"`
	Level    string `json:"level"`
	Owner    string `json:"owner"`
	Open     bool   `json:"open"`
	LogLevel int
	Es       *elasticsearch.Client
	Mu       sync.RWMutex
	HC       *HttpClient
}

type MsgBody struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Path    string `json:"path"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type ElasticLogBody struct {
	TimeStamp int64  `json:"timestamp"`
	Level     int64  `json:"level"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	Owner     string `json:"owner"`
}

var logIndexBody = `{
    "mappings": {
        "properties": {
            "timestamp": {
                "type": "date"
            },
            "level": {
                "type": "short"
            },
            "name": {
                "type": "keyword"
            },
			"owner": {
                "type": "keyword"
            },
            "content": {
                "type": "keyword"
            }
        }
    },
    "settings": {
        "index": {
            "max_result_window": 2000000000,
            "number_of_shards": 1,
            "number_of_replicas": 1
        }
    }
}`

func init() {
	Register(AdapterElastic, &elasticLogger{LogLevel: LevelTrace})
}

// Init 初始化
func (e *elasticLogger) Init(jsonConfig string) error {
	if len(jsonConfig) == 0 {
		return nil
	}

	err := json.Unmarshal([]byte(jsonConfig), &e)
	if err != nil {
		fmt.Println("unmarshal es err", err, jsonConfig)
		return err
	}

	if e.Open == false {
		return nil
	}

	if lv, ok := LevelMap[e.Level]; ok {
		e.LogLevel = lv
	}
	err = e.getClient()
	if err != nil {
		return err
	}
	e.HC = NewHttpClient(0, 0, 3)
	return e.CreateIndex()
}

// LogWrite
func (e *elasticLogger) LogWrite(when time.Time, msgText interface{}, level int) error {

	if level > e.LogLevel {
		return nil
	}

	msg, ok := msgText.(string)
	if !ok {
		return nil
	}

	if e.Es == nil {
		err := e.getClient()
		if err != nil {
			return err
		}
	}

	body := new(MsgBody)
	err := json.Unmarshal([]byte(msg), &body)
	if err != nil {
		return err
	}

	esBody := new(ElasticLogBody)
	esBody.TimeStamp = time.Now().UnixMicro() / 1000
	esBody.Name = body.Name
	esBody.Level = e.getLevelNum(body.Level)
	esBody.Content = strings.Replace(body.Content, "   ", "", -1)
	esBody.Path = body.Path
	esBody.Owner = e.Owner
	esByte, _ := json.Marshal(esBody)
	go e.saveMessage(string(esByte))
	return nil
}

// Destroy
func (e *elasticLogger) Destroy() {
	e.Es = nil
}

// getClient get elastic client
func (e *elasticLogger) getClient() (err error) {
	cfg := elasticsearch.Config{Addresses: []string{e.Addr}}
	e.Es, err = elasticsearch.NewClient(cfg)
	if err != nil {
		return errors.New(fmt.Sprintf("Get elastic client error %v", err))
	}
	return nil
}

// saveMessage save message to esdb
func (e *elasticLogger) saveMessage(msg string) error {
	dateTime := strconv.FormatInt(time.Now().UnixNano(), 10)
	req := esapi.IndexRequest{
		Index:      e.Index,
		DocumentID: dateTime,
		Body:       strings.NewReader(msg),
		Refresh:    "true",
	}
	res, err := req.Do(context.Background(), e.Es)
	if err != nil {
		fmt.Println("do err", err)
		return err
	}
	res.Body.Close()
	return nil
}

// CreateIndex create logger elastic index
func (e *elasticLogger) CreateIndex() error {
	if len(e.Index) == 0 {
		return fmt.Errorf("please set es index")
	}

	if e.CheckIndex() == true {
		return nil
	}

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	_, err := e.HC.Request(e.Addr+"/"+e.Index, "PUT", logIndexBody, header)
	return err
}

// CheckIndex Check if the index exists
func (e *elasticLogger) CheckIndex() bool {
	res, _ := e.HC.Request(e.Addr+"/"+e.Index, "GET", "", nil)
	if strings.Count(res, "index_not_found_exception") > 0 {
		return false
	}
	return true
}

func (e *elasticLogger) getLevelNum(levelStr string) int64 {
	switch levelStr {
	case "DEBG":
		return 10
	case "INFO":
		return 20
	case "WARN":
		return 30
	case "EROR":
		return 40
	default:
		return 0
	}
}
