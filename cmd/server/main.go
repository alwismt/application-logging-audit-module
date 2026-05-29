package main

import (
	"fmt"
	"os"

	"github.com/alwismt/application-logging-audit-module/pkg/loggingaudit"
)

func main() {
	mod, err := loggingaudit.NewFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "init module: %v\n", err)
		os.Exit(1)
	}

	if err := mod.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "run app: %v\n", err)
		os.Exit(1)
	}
}
