// Code generated by go-enum DO NOT EDIT.
// Version: 0.6.0
// Revision: 919e61c0174b91303753ee3898569a01abb32c97
// Build Date: 2023-12-18T15:54:43Z
// Built By: goreleaser

package logging

import (
	"fmt"
	"strings"
)

const (
	// LogFormatHuman is a LogFormat of type human.
	LogFormatHuman LogFormat = "human"
	// LogFormatJson is a LogFormat of type json.
	LogFormatJson LogFormat = "json"
)

var ErrInvalidLogFormat = fmt.Errorf("not a valid LogFormat, try [%s]", strings.Join(_LogFormatNames, ", "))

var _LogFormatNames = []string{
	string(LogFormatHuman),
	string(LogFormatJson),
}

// LogFormatNames returns a list of possible string values of LogFormat.
func LogFormatNames() []string {
	tmp := make([]string, len(_LogFormatNames))
	copy(tmp, _LogFormatNames)
	return tmp
}

// String implements the Stringer interface.
func (x LogFormat) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x LogFormat) IsValid() bool {
	_, err := ParseLogFormat(string(x))
	return err == nil
}

var _LogFormatValue = map[string]LogFormat{
	"human": LogFormatHuman,
	"json":  LogFormatJson,
}

// ParseLogFormat attempts to convert a string to a LogFormat.
func ParseLogFormat(name string) (LogFormat, error) {
	if x, ok := _LogFormatValue[name]; ok {
		return x, nil
	}
	return LogFormat(""), fmt.Errorf("%s is %w", name, ErrInvalidLogFormat)
}

// MarshalText implements the text marshaller method.
func (x LogFormat) MarshalText() ([]byte, error) {
	return []byte(string(x)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *LogFormat) UnmarshalText(text []byte) error {
	tmp, err := ParseLogFormat(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

const (
	// LogLevelWarn is a LogLevel of type warn.
	LogLevelWarn LogLevel = "warn"
	// LogLevelInfo is a LogLevel of type info.
	LogLevelInfo LogLevel = "info"
	// LogLevelDebug is a LogLevel of type debug.
	LogLevelDebug LogLevel = "debug"
	// LogLevelTrace is a LogLevel of type trace.
	LogLevelTrace LogLevel = "trace"
)

var ErrInvalidLogLevel = fmt.Errorf("not a valid LogLevel, try [%s]", strings.Join(_LogLevelNames, ", "))

var _LogLevelNames = []string{
	string(LogLevelWarn),
	string(LogLevelInfo),
	string(LogLevelDebug),
	string(LogLevelTrace),
}

// LogLevelNames returns a list of possible string values of LogLevel.
func LogLevelNames() []string {
	tmp := make([]string, len(_LogLevelNames))
	copy(tmp, _LogLevelNames)
	return tmp
}

// String implements the Stringer interface.
func (x LogLevel) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x LogLevel) IsValid() bool {
	_, err := ParseLogLevel(string(x))
	return err == nil
}

var _LogLevelValue = map[string]LogLevel{
	"warn":  LogLevelWarn,
	"info":  LogLevelInfo,
	"debug": LogLevelDebug,
	"trace": LogLevelTrace,
}

// ParseLogLevel attempts to convert a string to a LogLevel.
func ParseLogLevel(name string) (LogLevel, error) {
	if x, ok := _LogLevelValue[name]; ok {
		return x, nil
	}
	return LogLevel(""), fmt.Errorf("%s is %w", name, ErrInvalidLogLevel)
}

// MarshalText implements the text marshaller method.
func (x LogLevel) MarshalText() ([]byte, error) {
	return []byte(string(x)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *LogLevel) UnmarshalText(text []byte) error {
	tmp, err := ParseLogLevel(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}
