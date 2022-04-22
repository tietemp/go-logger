package logger

import (
	"testing"
	"time"
)

func TestLogOut(t *testing.T) {
	SetLogger(`{
		"Console": {
			"level": "DEBG",
			"color": true
		},
		"File": {
			"filename": "test.log",
			"level": "DEBG",
			"daily": true,
			"maxdays": -1,
			"append": true,
			"permit": "0660"
		}}`)
	Debug("🔨 show log info test", "time", time.Now().Unix())
	Info("🔨 show log info test", "time", time.Now().Unix())
}

func BenchmarkError(b *testing.B) {
	SetLogger("./log.json")
	for i := 0; i < b.N; i++ {
		Info("🔨 show log info test", "time", time.Now().Unix())
	}
}