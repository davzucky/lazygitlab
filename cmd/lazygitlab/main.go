package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/davzucky/lazygitlab/internal/app"
)

func main() {
	var projectOverride string
	var debug bool

	flag.StringVar(&projectOverride, "project", "", "GitLab project path (group/project)")
	flag.BoolVar(&debug, "debug", false, "Enable verbose debug logging")
	flag.Parse()

	if err := app.Run(context.Background(), app.Options{
		ProjectOverride: projectOverride,
		Debug:           debug,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "lazygitlab: %v\n", err)
		os.Exit(1)
	}
}
