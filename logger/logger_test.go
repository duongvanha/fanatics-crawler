package logger

import (
	"go.uber.org/zap"
	"testing"
)

func TestLogger(t *testing.T) {
	bkl, err := NewLogger("./log/out/log.log", 1, 10, 10)
	if err != nil {
		t.Error("Cannot create logger")
	}

	for i := 0; i < 1; i++ {
		bkl.Info("Log to file")
		bkl.Errorf("Testing error data %s", "test")
		bkl.Debug("Test debug log")
		bkl.SetLevel(zap.DebugLevel)
		bkl.Debug("Test dynamic debug log")
	}
	bkl.Close()
}
