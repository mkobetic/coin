package assert

import (
	"fmt"
	"testing"
)

func Equal(t *testing.T, actual, expected interface{}, args ...interface{}) bool {
	t.Helper()
	if actual == expected {
		return true
	}
	msg := msgFromArgs(args, "not equal")
	t.Errorf("%s\nexpected: %v\nactual:   %v\n", msg, expected, actual)
	return false
}

func NotNil(t *testing.T, v interface{}, args ...interface{}) bool {
	t.Helper()
	if v != nil {
		return true
	}
	msg := msgFromArgs(args, "is nil")
	t.Error(msg)
	return false
}

func True(t *testing.T, v bool, args ...interface{}) bool {
	t.Helper()
	if v {
		return true
	}
	msg := msgFromArgs(args, "not true")
	t.Error(msg)
	return false
}

func NoError(t *testing.T, err error, args ...interface{}) bool {
	t.Helper()
	if err == nil {
		return true
	}
	msg := msgFromArgs(args, err.Error())
	t.Errorf("%s\nerror: %s\n", msg, err)
	return false
}

func msgFromArgs(args []interface{}, defaultMsg string) string {
	if len(args) == 0 {
		return defaultMsg
	}
	msg := args[0].(string)
	if len(args) == 1 {
		return msg
	}
	return fmt.Sprintf(msg, args[1:]...)
}
