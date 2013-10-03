package main

import (
	"fmt"
	"os"
	"strings"
)

const info = `
ps 0: list processes

Usage: hk ps

Command ps lists running processes.
`

func maybePrintInfo() {
	if os.Getenv("HKPLUGINMODE") == "info" {
		fmt.Println(strings.TrimSpace(info))
		os.Exit(0)
	}
}
