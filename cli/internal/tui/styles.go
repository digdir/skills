package tui

import "fmt"

// ANSI color helpers
func bold(s string) string      { return "\033[1m" + s + "\033[0m" }
func dim(s string) string       { return "\033[2m" + s + "\033[0m" }
func green(s string) string     { return "\033[32m" + s + "\033[0m" }
func red(s string) string       { return "\033[31m" + s + "\033[0m" }
func yellow(s string) string    { return "\033[33m" + s + "\033[0m" }
func blue(s string) string      { return "\033[34m" + s + "\033[0m" }
func cyan(s string) string      { return "\033[36m" + s + "\033[0m" }
func boldGreen(s string) string { return "\033[1;32m" + s + "\033[0m" }
func boldBlue(s string) string  { return "\033[1;34m" + s + "\033[0m" }

// Screen control
func clearScreen()    { fmt.Print("\033[2J\033[H") }
func hideCursor()     { fmt.Print("\033[?25l") }
func showCursor()     { fmt.Print("\033[?25h") }
func moveTo(r, c int) { fmt.Printf("\033[%d;%dH", r, c) }
