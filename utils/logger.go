package utils

import (
	"log"
	"os"
)

// ErrorLog writes error messages to stderr
var ErrorLog = log.New(os.Stderr, "ERROR: ", log.LstdFlags)

// InfoLog writes informational messages to stdout
var InfoLog = log.New(os.Stdout, "", log.LstdFlags)
