package providers

import "github.com/nehemming/cirocket/pkg/loggee"

const (
	// LogInfo is an informational level of logging.
	LogInfo = Severity(iota)
	// LogWarn is a warning level of logging.
	LogWarn
	// LogError is an error level of logging.
	LogError
)

// Severity is the level to log messages at in a log provider.
type Severity int

// NewLogProvider creates provider that writes to a log.
func NewLogProvider(log loggee.Logger, severity Severity) ResourceProvider {
	var logFunc LineFunc

	switch severity {
	case LogInfo:
		logFunc = log.Info
	case LogWarn:
		logFunc = log.Warn
	case LogError:
		logFunc = log.Error
	default:
		logFunc = func(_ string) {}
	}

	return logFunc
}
