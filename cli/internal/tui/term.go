package tui

import (
	"os"

	"golang.org/x/term"
)

// keyEvent represents a parsed key press.
type keyEvent int

const (
	keyNone keyEvent = iota
	keyUp
	keyDown
	keyLeft
	keyRight
	keyEnter
	keySpace
	keyEsc
	keyBackspace
	keyTab
	keyCtrlC
	keyCtrlD
	keyRune // regular character
)

type keyInput struct {
	event keyEvent
	char  rune
}

// rawMode puts the terminal into raw mode and returns a restore function.
func rawMode() (restore func(), err error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	return func() { term.Restore(fd, oldState) }, nil
}

// readKey reads a single key press from stdin (must be in raw mode).
func readKey() keyInput {
	buf := make([]byte, 8)
	n, err := os.Stdin.Read(buf)
	if err != nil || n == 0 {
		return keyInput{event: keyNone}
	}

	b := buf[:n]

	// Escape sequences
	if n >= 3 && b[0] == 0x1b && b[1] == '[' {
		switch b[2] {
		case 'A':
			return keyInput{event: keyUp}
		case 'B':
			return keyInput{event: keyDown}
		case 'C':
			return keyInput{event: keyRight}
		case 'D':
			return keyInput{event: keyLeft}
		}
	}

	// Single byte
	if n == 1 {
		switch b[0] {
		case 3: // ctrl+c
			return keyInput{event: keyCtrlC}
		case 4: // ctrl+d
			return keyInput{event: keyCtrlD}
		case 9: // tab
			return keyInput{event: keyTab}
		case 13: // enter
			return keyInput{event: keyEnter}
		case 27: // esc
			return keyInput{event: keyEsc}
		case 32: // space
			return keyInput{event: keySpace}
		case 127: // backspace
			return keyInput{event: keyBackspace}
		default:
			if b[0] >= 32 && b[0] < 127 {
				return keyInput{event: keyRune, char: rune(b[0])}
			}
		}
	}

	return keyInput{event: keyNone}
}
