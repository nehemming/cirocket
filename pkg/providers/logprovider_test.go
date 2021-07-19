package providers

import (
	"context"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestNewLogProviderSeverityError(t *testing.T) {
	log := stdlog.New()

	lp := NewLogProvider(log, LogError)

	writer, err := lp.OpenWrite(context.Background())
	if err != nil {
		t.Error("error unexpected", err)
		return
	}
	defer writer.Close()

	_, err = writer.Write([]byte("hello"))
	if err != nil {
		t.Error("error unexpected", err)
	}
}

func TestNewLogProviderSeverityWarn(t *testing.T) {
	log := stdlog.New()

	lp := NewLogProvider(log, LogWarn)

	writer, err := lp.OpenWrite(context.Background())
	if err != nil {
		t.Error("error unexpected", err)
		return
	}
	defer writer.Close()

	_, err = writer.Write([]byte("hello"))
	if err != nil {
		t.Error("error unexpected", err)
	}
}

func TestNewLogProviderSeverityInfo(t *testing.T) {
	log := stdlog.New()

	lp := NewLogProvider(log, LogInfo)

	writer, err := lp.OpenWrite(context.Background())
	if err != nil {
		t.Error("error unexpected", err)
		return
	}
	defer writer.Close()

	_, err = writer.Write([]byte("hello"))
	if err != nil {
		t.Error("error unexpected", err)
	}
}

func TestNewLogProviderSeverityUndefined(t *testing.T) {
	log := stdlog.New()

	lp := NewLogProvider(log, Severity(99))

	writer, err := lp.OpenWrite(context.Background())
	if err != nil {
		t.Error("error unexpected", err)
		return
	}
	defer writer.Close()

	_, err = writer.Write([]byte("hello"))
	if err != nil {
		t.Error("error unexpected", err)
	}
}
