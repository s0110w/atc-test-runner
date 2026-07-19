// Package tui provides the task selection UI for `atr new`.
// It drives the terminal directly (raw mode via termios) to stay
// dependency-free.
package tui

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func ioctl(fd uintptr, req uint64, arg *syscall.Termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(req), uintptr(unsafe.Pointer(arg)))
	if errno != 0 {
		return errno
	}
	return nil
}

// makeRaw disables echo, line buffering and signal keys (Ctrl-C is
// handled as a key so the terminal state is always restored).
func makeRaw(fd uintptr) (syscall.Termios, error) {
	var old syscall.Termios
	if err := ioctl(fd, ioctlGet, &old); err != nil {
		return old, fmt.Errorf("stdin is not a terminal (task selection needs an interactive session)")
	}
	raw := old
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	if err := ioctl(fd, ioctlSet, &raw); err != nil {
		return old, err
	}
	return old, nil
}

// SelectTasks shows a checkbox list and returns the chosen indices.
// All items start selected, so plain enter keeps the default behavior.
// It returns an error if the user cancels (q / Ctrl-C) or stdin is not a
// terminal.
func SelectTasks(title string, items []string) ([]int, error) {
	fd := os.Stdin.Fd()
	old, err := makeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer ioctl(fd, ioctlSet, &old)

	selected := make([]bool, len(items))
	for i := range selected {
		selected[i] = true
	}
	cursor := 0

	fmt.Print("\033[?25l")       // hide cursor
	defer fmt.Print("\033[?25h") // show cursor

	fmt.Printf("%s  (↑↓/jk: move, space: toggle, a: all, enter: ok, q: cancel)\n", title)
	draw := func(redraw bool) {
		if redraw {
			fmt.Printf("\033[%dA", len(items)) // move back to the first item line
		}
		for i, item := range items {
			mark, ptr := " ", " "
			if selected[i] {
				mark = "x"
			}
			if i == cursor {
				ptr = ">"
			}
			fmt.Printf("\r%s [%s] %s\033[K\n", ptr, mark, item)
		}
	}
	draw(false)

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return nil, err
		}
		key := buf[0]
		if n == 3 && buf[0] == 0x1b && buf[1] == '[' { // arrow keys
			switch buf[2] {
			case 'A':
				key = 'k'
			case 'B':
				key = 'j'
			}
		}
		switch key {
		case 'k':
			if cursor > 0 {
				cursor--
			}
		case 'j':
			if cursor < len(items)-1 {
				cursor++
			}
		case ' ':
			selected[cursor] = !selected[cursor]
		case 'a':
			all := true
			for _, s := range selected {
				if !s {
					all = false
					break
				}
			}
			for i := range selected {
				selected[i] = !all
			}
		case '\r', '\n':
			var picked []int
			for i, s := range selected {
				if s {
					picked = append(picked, i)
				}
			}
			return picked, nil
		case 'q', 3: // q or Ctrl-C
			return nil, fmt.Errorf("canceled")
		}
		draw(true)
	}
}
