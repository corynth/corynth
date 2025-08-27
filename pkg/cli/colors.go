package cli

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Color codes for professional CLI output
const (
	// Basic colors
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	
	// Foreground colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	
	// Bright foreground colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"
	
	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// Professional color schemes
var (
	// Status colors
	SuccessColor = BrightGreen
	ErrorColor   = BrightRed
	WarningColor = BrightYellow
	InfoColor    = BrightBlue
	
	// UI element colors
	HeaderColor    = Bold + BrightCyan
	SubHeaderColor = Bold + BrightWhite
	LabelColor     = BrightBlue
	ValueColor     = BrightWhite
	DimColor       = BrightBlack
	
	// Command-specific colors
	CommandColor   = Bold + BrightMagenta
	FlagColor      = BrightCyan
	ExampleColor   = BrightGreen
	
	// Progress colors
	StepColor      = BrightYellow
	CompletedColor = BrightGreen
	FailedColor    = BrightRed
	SkippedColor   = BrightBlack
)

// ColorEnabled determines if colors should be used
var ColorEnabled = true

func init() {
	// Disable colors on Windows by default unless explicitly enabled
	if runtime.GOOS == "windows" {
		if os.Getenv("FORCE_COLOR") == "" && os.Getenv("CORYNTH_COLOR") == "" {
			ColorEnabled = false
		}
	}
	
	// Check environment variables
	if os.Getenv("NO_COLOR") != "" || os.Getenv("CORYNTH_NO_COLOR") != "" {
		ColorEnabled = false
	}
	
	// Check if output is redirected
	if !isTerminal() {
		ColorEnabled = false
	}
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// Colorize applies color to text if colors are enabled
func Colorize(color, text string) string {
	if !ColorEnabled {
		return text
	}
	return color + text + Reset
}

// Professional formatting functions
func Success(text string) string {
	return Colorize(SuccessColor, "✓ "+text)
}

func Error(text string) string {
	return Colorize(ErrorColor, "✗ "+text)
}

func Warning(text string) string {
	return Colorize(WarningColor, "⚠ "+text)
}

func Info(text string) string {
	return Colorize(InfoColor, "ℹ "+text)
}

func Header(text string) string {
	return Colorize(HeaderColor, text)
}

func SubHeader(text string) string {
	return Colorize(SubHeaderColor, text)
}

func Label(text string) string {
	return Colorize(LabelColor, text)
}

func Value(text string) string {
	return Colorize(ValueColor, text)
}

func Command(text string) string {
	return Colorize(CommandColor, text)
}

func Flag(text string) string {
	return Colorize(FlagColor, text)
}

func Example(text string) string {
	return Colorize(ExampleColor, text)
}

func DimText(text string) string {
	return Dim + text + Reset
}


func Step(text string) string {
	return Colorize(StepColor, "• "+text)
}

func Completed(text string) string {
	return Colorize(CompletedColor, "✓ "+text)
}

func Failed(text string) string {
	return Colorize(FailedColor, "✗ "+text)
}

func Skipped(text string) string {
	return Colorize(SkippedColor, "⊘ "+text)
}

// Progress indicator
func Progress(current, total int, operation string) string {
	percentage := float64(current) / float64(total) * 100
	progressBar := generateProgressBar(current, total, 20)
	
	status := fmt.Sprintf("[%d/%d] %.0f%% %s %s",
		current, total, percentage, progressBar, operation)
	
	return Colorize(InfoColor, status)
}

// generateProgressBar creates a visual progress bar
func generateProgressBar(current, total, width int) string {
	if total == 0 {
		return strings.Repeat("█", width)
	}
	
	filled := int(float64(current) / float64(total) * float64(width))
	empty := width - filled
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("[%s]", bar)
}

// Table formatting
func TableHeader(headers ...string) string {
	var colored []string
	for _, header := range headers {
		colored = append(colored, Colorize(HeaderColor, header))
	}
	return strings.Join(colored, " | ")
}

func TableRow(values ...string) string {
	var colored []string
	for i, value := range values {
		if i == 0 {
			// First column in a different color
			colored = append(colored, Colorize(LabelColor, value))
		} else {
			colored = append(colored, Colorize(ValueColor, value))
		}
	}
	return strings.Join(colored, " | ")
}

// Status indicators
func StatusRunning(text string) string {
	return Colorize(StepColor, "⏳ "+text)
}

func StatusSuccess(text string) string {
	return Colorize(SuccessColor, "✅ "+text)
}

func StatusFailed(text string) string {
	return Colorize(ErrorColor, "❌ "+text)
}

func StatusWarning(text string) string {
	return Colorize(WarningColor, "⚠️  "+text)
}

// Logo and branding
func Logo() string {
	logo := `
 ██████╗ ██████╗ ██████╗ ██╗   ██╗███╗   ██╗████████╗██╗  ██╗
██╔════╝██╔═══██╗██╔══██╗╚██╗ ██╔╝████╗  ██║╚══██╔══╝██║  ██║
██║     ██║   ██║██████╔╝ ╚████╔╝ ██╔██╗ ██║   ██║   ███████║
██║     ██║   ██║██╔══██╗  ╚██╔╝  ██║╚██╗██║   ██║   ██╔══██║
╚██████╗╚██████╔╝██║  ██║   ██║   ██║ ╚████║   ██║   ██║  ██║
 ╚═════╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═══╝   ╚═╝   ╚═╝  ╚═╝`
	
	return Colorize(HeaderColor, logo)
}

// Banner creates a professional banner
func Banner(title, subtitle string) string {
	border := strings.Repeat("=", 60)
	
	return fmt.Sprintf("\n%s\n%s\n%s\n%s\n%s\n",
		Colorize(HeaderColor, border),
		Colorize(HeaderColor, fmt.Sprintf("  %s", title)),
		Colorize(SubHeaderColor, fmt.Sprintf("  %s", subtitle)),
		Colorize(HeaderColor, border),
		"")
}

// Section creates a section divider
func Section(title string) string {
	border := strings.Repeat("-", len(title)+4)
	return fmt.Sprintf("\n%s\n%s %s\n%s\n",
		Colorize(DimColor, border),
		Colorize(SubHeaderColor, "►"),
		Colorize(SubHeaderColor, title),
		Colorize(DimColor, border))
}

// Bullet point lists
func BulletPoint(text string) string {
	return Colorize(InfoColor, "  • ") + text
}

func NumberedPoint(num int, text string) string {
	return Colorize(InfoColor, fmt.Sprintf("  %d. ", num)) + text
}

// Code formatting
func Code(text string) string {
	return Colorize(ExampleColor, "`"+text+"`")
}

func CodeBlock(text string) string {
	lines := strings.Split(text, "\n")
	var coloredLines []string
	
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			coloredLines = append(coloredLines, Colorize(ExampleColor, "  "+line))
		} else {
			coloredLines = append(coloredLines, "")
		}
	}
	
	return strings.Join(coloredLines, "\n")
}