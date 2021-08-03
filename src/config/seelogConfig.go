package config

import "github.com/cihub/seelog"

func SetupLogger() {
	logger, err := seelog.LoggerFromConfigAsFile("resource/seelog.xml")
	if err != nil {
		return
	}
	seelog.ReplaceLogger(logger)
}
