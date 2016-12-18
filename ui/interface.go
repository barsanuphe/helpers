package ui

// UserInterface deals with user input, output and logging.
type UserInterface interface {
	// input
	GetInput() (string, error)
	Accept(string) bool
	UpdateValue(string, string, string, bool) (string, error)
	SelectOption(string, string, []string, bool) (string, error)
	Edit(string) (string, error)
	// output
	Title(string, ...interface{})
	SubTitle(string, ...interface{})
	SubPart(string, ...interface{})
	Choice(string, ...interface{})
	Display(string)
	Tag(string, bool) string
	// log
	InitLogger(string) error
	CloseLog()
	Error(string)
	Errorf(string, ...interface{})
	Warning(string)
	Warningf(string, ...interface{})
	Info(string)
	Infof(string, ...interface{})
	Debug(string)
	Debugf(string, ...interface{})
}
