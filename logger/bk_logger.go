package logger

import (
	"fmt"
	"github.com/spf13/cast"
	"go.uber.org/zap/zapcore"
	"math/rand"
	"time"
)

const maxLogMessageLen int = 10000

var shopIdsDebug map[int64]bool

func init() {
	rand.Seed(time.Now().Unix())
}

func precheckMsgLen(template string, fmtArgs []interface{}) []string {
	msg := template
	if msg == "" && len(fmtArgs) > 0 {
		msg = fmt.Sprint(fmtArgs...)
	} else if msg != "" && len(fmtArgs) > 0 {
		msg = fmt.Sprintf(template, fmtArgs...)
	}

	var splitMsg = make([]string, 0)
	var l = len(msg)
	var id = rand.Int31()
	for i := 0; i <= len(msg)/maxLogMessageLen; i++ {
		var chunkLen = (i + 1) * maxLogMessageLen
		if l < chunkLen {
			chunkLen = l
		}

		var chunkMsg = msg[i*maxLogMessageLen : chunkLen]
		if i == 0 || len(chunkMsg) > 0 {
			if l > maxLogMessageLen {
				chunkMsg = fmt.Sprint(id, ":", i, " ", chunkMsg)
			}

			splitMsg = append(splitMsg, chunkMsg)
		}
	}

	return splitMsg
}

// Debug uses fmt.Sprint to construct and log a message.
func (b *BkLogger) Debug(args ...interface{}) {
	chunks := precheckMsgLen("", args)
	for _, c := range chunks {
		b.SugaredLogger.Debug(c)
	}
}

// Info uses fmt.Sprint to construct and log a message.
func (b *BkLogger) Info(args ...interface{}) {
	chunks := precheckMsgLen("", args)
	for _, c := range chunks {
		b.SugaredLogger.Info(c)
	}
}

// Warn uses fmt.Sprint to construct and log a message.
func (b *BkLogger) Warn(args ...interface{}) {
	chunks := precheckMsgLen("", args)
	for _, c := range chunks {
		b.SugaredLogger.Warn(c)
	}
}

// Error uses fmt.Sprint to construct and log a message.
func (b *BkLogger) Error(args ...interface{}) {
	//b.SugaredLogger.Error(args...)
	chunks := precheckMsgLen("", args)
	for _, c := range chunks {
		b.SugaredLogger.Error(c)
	}
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (b *BkLogger) DPanic(args ...interface{}) {
	b.SugaredLogger.DPanic(args...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func (b *BkLogger) Panic(args ...interface{}) {
	b.SugaredLogger.Panic(args...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (b *BkLogger) Fatal(args ...interface{}) {
	b.SugaredLogger.Fatal(args...)
}

// DebugWithoutShopTestf - uses fmt.Sprintf to log a templated message.
func (b *BkLogger) DebugCsf(shopID int64, template string, args ...interface{}) {
	chunks := precheckMsgLen(template, args)
	if shopIdsDebug == nil {
		for _, c := range chunks {
			b.SugaredLogger.Debug(c)
		}
	}
	if isShopDebug, ok := shopIdsDebug[shopID]; isShopDebug && ok {
		for _, c := range chunks {
			b.SugaredLogger.Info(c)
		}
	} else {
		for _, c := range chunks {
			b.SugaredLogger.Debug(c)
		}
	}
}

// Debugf uses fmt.Sprintf to log a templated message.
func (b *BkLogger) Debugf(template string, args ...interface{}) {
	chunks := precheckMsgLen(template, args)
	for _, c := range chunks {
		b.SugaredLogger.Debug(c)
	}
}

// Infof uses fmt.Sprintf to log a templated message.
func (b *BkLogger) Infof(template string, args ...interface{}) {
	chunks := precheckMsgLen(template, args)
	for _, c := range chunks {
		b.SugaredLogger.Info(c)
	}
}

// Warnf uses fmt.Sprintf to log a templated message.
func (b *BkLogger) Warnf(template string, args ...interface{}) {
	chunks := precheckMsgLen(template, args)
	for _, c := range chunks {
		b.SugaredLogger.Warn(c)
	}
}

// Errorf uses fmt.Sprintf to log a templated message.
func (b *BkLogger) Errorf(template string, args ...interface{}) {
	//b.SugaredLogger.Errorf(template, args...)
	chunks := precheckMsgLen(template, args)
	for _, c := range chunks {
		b.SugaredLogger.Error(c)
	}
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (b *BkLogger) DPanicf(template string, args ...interface{}) {
	b.SugaredLogger.DPanicf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func (b *BkLogger) Panicf(template string, args ...interface{}) {
	b.SugaredLogger.Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (b *BkLogger) Fatalf(template string, args ...interface{}) {
	b.SugaredLogger.Fatalf(template, args)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//  s.With(keysAndValues).Debug(msg)
func (b *BkLogger) Debugw(msg string, keysAndValues ...interface{}) {
	b.SugaredLogger.Debugw(msg, keysAndValues...)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (b *BkLogger) Infow(msg string, keysAndValues ...interface{}) {
	b.SugaredLogger.Infow(msg, keysAndValues...)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (b *BkLogger) Warnw(msg string, keysAndValues ...interface{}) {
	b.SugaredLogger.Warnw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (b *BkLogger) Errorw(msg string, keysAndValues ...interface{}) {
	b.SugaredLogger.Errorw(msg, keysAndValues...)
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func (b *BkLogger) DPanicw(msg string, keysAndValues ...interface{}) {
	b.SugaredLogger.DPanicw(msg, keysAndValues...)
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func (b *BkLogger) Panicw(msg string, keysAndValues ...interface{}) {
	b.SugaredLogger.Panicw(msg, keysAndValues...)
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func (b *BkLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	b.SugaredLogger.Fatalw(msg, keysAndValues...)
}

func (b *BkLogger) SetLevel(l zapcore.Level) {
	b.logLevel.SetLevel(l)
}

func (b *BkLogger) GetLevel() string {
	return b.logLevel.Level().String()
}

func (b *BkLogger) InitShopDebug(shopIdsStr []string) {
	shopIdsDebug = make(map[int64]bool)
	for _, shopIdStr := range shopIdsStr {
		shopIdsDebug[cast.ToInt64(shopIdStr)] = true
	}
}
