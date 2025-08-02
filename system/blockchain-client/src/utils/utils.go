package utils

import "pkg-common/logger"

func FailOnError(err error, msg string) {
	if err != nil {
		logger.Default().Fatal(err, msg)
	}
}
