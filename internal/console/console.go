package console

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	BgRed   = "\033[41m"
	BgGreen = "\033[42m"
	BgBlue  = "\033[44m"
	BgCyan  = "\033[46m"
)

func init() {
	enableWindowsANSI()
}

// enableWindowsANSI enables ANSI escape code processing on Windows 10+
func enableWindowsANSI() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")

	handle := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	r, _, _ := getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if r == 0 {
		return
	}
	// ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
	mode |= 0x0004
	setConsoleMode.Call(uintptr(handle), uintptr(mode))
}

// Logo prints the APM ASCII art logo
func Logo() {
	logo := fmt.Sprintf(`%s%s
     █████╗ ██████╗ ███╗   ███╗
    ██╔══██╗██╔══██╗████╗ ████║
    ███████║██████╔╝██╔████╔██║
    ██╔══██║██╔═══╝ ██║╚██╔╝██║
    ██║  ██║██║     ██║ ╚═╝ ██║
    ╚═╝  ╚═╝╚═╝     ╚═╝     ╚═╝
    %sAwesome Package Manager%s
%s`, BrightCyan, Bold, BrightYellow, BrightCyan, Reset)
	fmt.Println(logo)
}

// Info prints an info message
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s ℹ %s %s%s\n", BrightBlue, Bold, Reset, msg, Reset)
}

// AskYesNoConsole asks a yes/no question in the console
func AskYesNoConsole(question string) bool {
	fmt.Printf("%s%s ❓ %s%s %s[y/N]%s: ", BrightCyan, Bold, Reset, question, BrightYellow, Reset)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes" || response == "д" || response == "да"
}

// Success prints a success message
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s ✓ %s %s%s\n", BrightGreen, Bold, Reset, msg, Reset)
}

// Warning prints a warning message
func Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s ⚠ %s %s%s\n", BrightYellow, Bold, Reset, msg, Reset)
}

// Error prints an error message
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s ✗ %s %s%s\n", BrightRed, Bold, Reset, msg, Reset)
}

// Step prints a step message with an icon
func Step(icon string, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s %s %s %s%s\n", BrightCyan, Bold, icon, Reset, msg, Reset)
}

// Package prints a package name highlighted
func PackageName(name string) string {
	return fmt.Sprintf("%s%s%s%s", Bold, BrightCyan, name, Reset)
}

// Version prints a version highlighted
func VersionStr(version string) string {
	return fmt.Sprintf("%s%s%s", BrightYellow, version, Reset)
}

// Table prints a formatted table
func Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	fmt.Print("  ")
	for i, h := range headers {
		fmt.Printf("%s%s%-*s%s", Bold, BrightWhite, widths[i]+3, h, Reset)
	}
	fmt.Println()

	// Print separator
	fmt.Print("  ")
	for i := range headers {
		fmt.Printf("%s%s%s", Dim, strings.Repeat("─", widths[i]+3), Reset)
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		fmt.Print("  ")
		for i, cell := range row {
			if i == 0 {
				fmt.Printf("%s%s%-*s%s", Bold, BrightCyan, widths[i]+3, cell, Reset)
			} else if i == len(row)-1 {
				fmt.Printf("%s%-*s%s", BrightYellow, widths[i]+3, cell, Reset)
			} else {
				fmt.Printf("%-*s", widths[i]+3, cell)
			}
		}
		fmt.Println()
	}
}

// ProgressBar renders a progress bar
func ProgressBar(current, total int64, width int) string {
	if total == 0 {
		return ""
	}
	ratio := float64(current) / float64(total)
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}

	bar := fmt.Sprintf("%s%s%s%s",
		BrightGreen,
		strings.Repeat("█", filled),
		strings.Repeat("░", width-filled),
		Reset,
	)

	percent := int(ratio * 100)
	return fmt.Sprintf("%s %d%%", bar, percent)
}

// FormatBytes formats bytes into human-readable format
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
