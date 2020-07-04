package lhw

type Logger interface {
	Printf(format string, v ...interface{})
}
