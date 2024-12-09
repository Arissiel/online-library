package logger

import (
	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func InitLogger() {
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	Log.SetLevel(logrus.DebugLevel)

	Log.SetOutput(logrus.StandardLogger().Out)
}
