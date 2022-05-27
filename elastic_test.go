package logger

import "testing"

func TestElasticLogger_CreateIndex(t *testing.T) {
	es := new(elasticLogger)
	es.Addr = "http://119.28.70.87:9200"
	es.Index = "chainlogs"
	es.HC = NewHttpClient(0, 0, 3)
	err := es.CreateIndex()
	if err != nil {
		t.Fatal("create index err:", err)
	}
}

func TestElasticLogger_CheckIndex(t *testing.T) {
	es := new(elasticLogger)
	es.Addr = "http://119.28.70.87:9200"
	es.Index = "tie-logsa"
	es.HC = NewHttpClient(0, 0, 3)
	ok := es.CheckIndex()
	t.Log(ok)
}
