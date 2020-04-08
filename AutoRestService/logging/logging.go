package logging

import (
	"fmt"
	"log"

	golf "github.com/aphistic/golf"
)

/*
ServiceLogger main type for logging
*/
type ServiceLogger struct {
	GelfURL    string
	GelfPort   int
	SystemID   string
	Attrs      map[string]interface{}
	gelfActive bool
	c          *golf.Client
}

/*
InitGelf initialise gelf logging
*/
func (s *ServiceLogger) InitGelf() {
	s.gelfActive = false
	if s.GelfURL != "" {
		s.c, _ = golf.NewClient()
		s.c.Dial(fmt.Sprintf("udp://%s:%d", s.GelfURL, s.GelfPort))

		l, _ := s.c.NewLogger()

		golf.DefaultLogger(l)
		for key, value := range s.Attrs {
			l.SetAttr(key, value)
		}
		l.SetAttr("system_id", s.SystemID)
		s.gelfActive = true
	}
}

/*
Debug log this maeesage at debug level
*/
func (s *ServiceLogger) Debug(msg string) {
	if s.gelfActive {
		golf.Info(msg)
	}
	log.Println(msg)
}

/*
Debugf log this maeesage at debug level with formatting
*/
func (s *ServiceLogger) Debugf(format string, va ...interface{}) {
	if s.gelfActive {
		golf.Infof(format, va...)
	}
	log.Printf(format+"\n", va...)
}

/*
Info log this maeesage at info level
*/
func (s *ServiceLogger) Info(msg string) {
	if s.gelfActive {
		golf.Info(msg)
	}
	log.Println(msg)
}

/*
Infof log this maeesage at info level with formatting
*/
func (s *ServiceLogger) Infof(format string, va ...interface{}) {
	if s.gelfActive {
		golf.Infof(format, va...)
	}
	log.Printf(format+"\n", va...)
}

/*
Alert log this maeesage at alert level
*/
func (s *ServiceLogger) Alert(msg string) {
	if s.gelfActive {
		golf.Alert(msg)
	}
	log.Printf("Alert: %s\n", msg)
}

/*
Alertf log this maeesage at alert level with formatting
*/
func (s *ServiceLogger) Alertf(format string, va ...interface{}) {
	if s.gelfActive {
		golf.Alertf(format, va...)
	}
	log.Printf("Alert: %s\n", fmt.Sprintf(format, va...))
}

// Fatal logs a message at level Fatal on the standard logger.
func (s *ServiceLogger) Fatal(msg string) {
	if s.gelfActive {
		golf.Crit(msg)
	}
	log.Printf("Fatal: %s\n", msg)
}

// Fatalf logs a message at level Fatal on the standard logger.
func (s *ServiceLogger) Fatalf(format string, va ...interface{}) {
	if s.gelfActive {
		golf.Critf(format, va...)
	}
	log.Printf("Fatal: %s\n", fmt.Sprintf(format, va...))
}

/*
Close this logging client
*/
func (s *ServiceLogger) Close() {
	if s.gelfActive {
		s.c.Close()
	}
}
