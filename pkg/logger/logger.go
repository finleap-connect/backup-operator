/*
Copyright 2020 Backup Operator Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logger

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

type Logger = logr.Logger

var log Logger

func init() {
	logMode := os.Getenv("LOG_MODE")
	var (
		zapLog *zap.Logger
		err    error
	)
	if logMode == "" || logMode == "dev" {
		zapLog, err = zap.NewDevelopment()
	} else if logMode == "prod" {
		zapLog, err = zap.NewProduction()
	} else {
		zapLog = zap.NewNop()
	}
	if err != nil {
		panic(fmt.Sprintf("log setup failed: %v", err))
	}
	log = zapr.NewLogger(zapLog)
}

func WithName(name string) Logger {
	return log.WithName(name)
}
