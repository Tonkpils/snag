package vow

import "fmt"

func red(s string) string {
	return fmt.Sprintf("\x1b[0;31m%s\x1b[0m", s)
}

func yellow(s string) string {
	return fmt.Sprintf("\x1b[1;33m%s\x1b[0m", s)
}

func green(s string) string {
	return fmt.Sprintf("\x1b[0;32m%s\x1b[0m", s)
}
