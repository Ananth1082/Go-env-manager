package env_manager

const (
	LOW = iota + 1
	MED
	HIGH
)

func (e *EnvManager) Log(level int, msg string, args ...any) {
	if e.logMode > level {
		return
	}
	e.logger.Printf(msg+"\n", args...)
}

func (e *EnvManager) LogFatal(level int, msg string, args ...any) {
	if e.logMode > level {
		return
	}
	e.logger.Fatalf(msg+"\n", args...)
}
