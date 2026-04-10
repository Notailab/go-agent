package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Reporter interface {
	ToolCall(name, args string)
	AssistantStart()
	AssistantDelta(content string)
	AssistantEnd()
	AssistantMessage(content string)
	Errorf(format string, args ...any)
}

type NoopReporter struct{}

func (NoopReporter) ToolCall(name, args string) {}

func (NoopReporter) AssistantStart() {}

func (NoopReporter) AssistantDelta(content string) {}

func (NoopReporter) AssistantEnd() {}

func (NoopReporter) AssistantMessage(content string) {}

func (NoopReporter) Errorf(format string, args ...any) {}

type StdoutReporter struct {
	mu                sync.Mutex
	streamBuffer      strings.Builder
	streamInCodeBlock bool
	assistantPrinted  bool
}

func (r *StdoutReporter) ResetDialog() {
	r.mu.Lock()
	r.streamBuffer.Reset()
	r.streamInCodeBlock = false
	r.assistantPrinted = false
	r.mu.Unlock()
}

func (r *StdoutReporter) ToolCall(name, args string) {
	formattedArgs := summarizeInlineArgs(args)
	if formattedArgs == "" {
		formattedArgs = "(no arguments)"
	}

	fmt.Printf("%s● Tool%s %s%s%s %s%s%s\n", colorCyanBold, colorReset, colorYellowBold, name, colorReset, colorMagenta, formattedArgs, colorReset)
}

func (r *StdoutReporter) AssistantStart() {
	r.printAssistantHeader()
}

func (r *StdoutReporter) AssistantDelta(content string) {
	if strings.TrimSpace(content) == "" {
		return
	}

	r.mu.Lock()
	r.streamBuffer.WriteString(content)
	r.flushStreamLocked(false)
	r.mu.Unlock()
}

func (r *StdoutReporter) AssistantEnd() {
	r.mu.Lock()
	r.flushStreamLocked(true)
	r.mu.Unlock()
	fmt.Println()
}

func (r *StdoutReporter) AssistantMessage(content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}

	r.printAssistantHeader()
	fmt.Println(renderMarkdownBlock(content))
}

func (r *StdoutReporter) Errorf(format string, args ...any) {
	fmt.Printf("\n%s✗ %s%s\n", colorRedBold, fmt.Sprintf(format, args...), colorReset)
}

func (r *StdoutReporter) flushStreamLocked(final bool) {
	data := r.streamBuffer.String()
	if data == "" {
		return
	}

	lastBreak := strings.LastIndex(data, "\n")
	if !final && lastBreak < 0 {
		return
	}

	var chunk string
	if final || lastBreak == len(data)-1 {
		chunk = data
		r.streamBuffer.Reset()
	} else {
		chunk = data[:lastBreak+1]
		remaining := data[lastBreak+1:]
		r.streamBuffer.Reset()
		r.streamBuffer.WriteString(remaining)
	}

	if chunk == "" {
		return
	}

	fmt.Print(r.renderStreamMarkdown(chunk, final))
}

func (r *StdoutReporter) renderStreamMarkdown(content string, final bool) string {
	var builder strings.Builder
	lines := strings.Split(content, "\n")
	for idx, rawLine := range lines {
		isLast := idx == len(lines)-1
		line := strings.TrimRight(rawLine, "\r")
		if !final && isLast && line == "" {
			continue
		}
		if line == "" && !r.streamInCodeBlock {
			builder.WriteString("\n")
			continue
		}

		rendered, keepOpen := r.renderStreamLine(line)
		builder.WriteString(rendered)
		if !isLast || final {
			builder.WriteString("\n")
		}
		r.streamInCodeBlock = keepOpen
	}

	return builder.String()
}

func (r *StdoutReporter) renderStreamLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "```") {
		if r.streamInCodeBlock {
			return renderCodeBlockLine(line, true), false
		}
		return renderCodeBlockLine(line, true), true
	}

	if r.streamInCodeBlock {
		return renderCodeBlockLine(line, false), true
	}

	if trimmed == "" {
		return "", false
	}

	if strings.HasPrefix(trimmed, "#") {
		level := 0
		for level < len(trimmed) && trimmed[level] == '#' {
			level++
		}
		headline := strings.TrimSpace(trimmed[level:])
		return colorCyanBold + strings.Repeat("#", level) + " " + formatInlineMarkdown(headline) + colorReset, false
	}

	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		return colorYellowBold + "• " + colorReset + formatInlineMarkdown(strings.TrimSpace(trimmed[2:])), false
	}

	if strings.HasPrefix(trimmed, ">") {
		return colorDim + "│ " + colorReset + formatInlineMarkdown(strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))), false
	}

	return formatInlineMarkdown(line), false
}

const (
	colorReset      = "\x1b[0m"
	colorBold       = "\x1b[1m"
	colorDim        = "\x1b[2m"
	colorRedBold    = "\x1b[1;31m"
	colorGreenBold  = "\x1b[1;32m"
	colorYellowBold = "\x1b[1;33m"
	colorBlueBold   = "\x1b[1;34m"
	colorCyanBold   = "\x1b[1;36m"
	colorMagenta    = "\x1b[35m"
	colorCodeBg     = "\x1b[48;5;236m"
	colorCodeFg     = "\x1b[38;5;251m"
)

var (
	boldPattern = regexp.MustCompile(`\*\*(.+?)\*\*`)
	codePattern = regexp.MustCompile("`([^`]+)`")
)

func prettyJSON(input string) (string, bool) {
	var buffer bytes.Buffer
	if err := json.Indent(&buffer, []byte(input), "", "  "); err != nil {
		return input, false
	}
	return buffer.String(), true
}

func summarizeInlineArgs(args string) string {
	trimmed := strings.TrimSpace(args)
	if trimmed == "" {
		return ""
	}

	if pretty, ok := prettyJSON(trimmed); ok {
		trimmed = pretty
	}

	if snippet := extractCommandSnippet(trimmed); snippet != "" {
		trimmed = snippet
	}

	trimmed = strings.Join(strings.Fields(trimmed), " ")
	return truncateWithEllipsis(trimmed, 72)
}

func extractCommandSnippet(input string) string {
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(input), &payload); err != nil {
		return ""
	}

	for _, key := range []string{"command", "path", "file_path", "filePath", "content", "query"} {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

func truncateWithEllipsis(input string, max int) string {
	runes := []rune(input)
	if len(runes) <= max {
		return input
	}
	if max <= 1 {
		return "…"
	}
	return string(runes[:max-1]) + "…"
}

func (r *StdoutReporter) printAssistantHeader() {
	r.mu.Lock()
	shouldPrint := !r.assistantPrinted
	r.streamBuffer.Reset()
	r.streamInCodeBlock = false
	if shouldPrint {
		r.assistantPrinted = true
	}
	r.mu.Unlock()

	if shouldPrint {
		fmt.Printf("\n%sAssistant%s\n", colorGreenBold, colorReset)
	}
}

func renderMarkdownBlock(content string) string {
	var builder strings.Builder
	lines := strings.Split(content, "\n")
	inCodeBlock := false

	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r")
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "```"):
			if inCodeBlock {
				builder.WriteString(renderCodeBlockLine(line, true))
				builder.WriteString("\n")
				inCodeBlock = false
				continue
			}
			inCodeBlock = true
			builder.WriteString(renderCodeBlockLine(line, true))
			builder.WriteString("\n")
			continue
		case inCodeBlock:
			builder.WriteString(renderCodeBlockLine(line, false))
			builder.WriteString("\n")
			continue
		case trimmed == "":
			builder.WriteString("\n")
			continue
		case strings.HasPrefix(trimmed, "#"):
			level := 0
			for level < len(trimmed) && trimmed[level] == '#' {
				level++
			}
			headline := strings.TrimSpace(trimmed[level:])
			builder.WriteString(colorCyanBold)
			builder.WriteString(strings.Repeat("#", level))
			builder.WriteString(" ")
			builder.WriteString(formatInlineMarkdown(headline))
			builder.WriteString(colorReset)
			builder.WriteString("\n")
			continue
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			builder.WriteString(colorYellowBold)
			builder.WriteString("• ")
			builder.WriteString(colorReset)
			builder.WriteString(formatInlineMarkdown(strings.TrimSpace(trimmed[2:])))
			builder.WriteString("\n")
			continue
		case strings.HasPrefix(trimmed, ">"):
			builder.WriteString(colorDim)
			builder.WriteString("│ ")
			builder.WriteString(colorReset)
			builder.WriteString(formatInlineMarkdown(strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))))
			builder.WriteString("\n")
			continue
		default:
			builder.WriteString(formatInlineMarkdown(line))
			builder.WriteString("\n")
		}
	}

	return strings.TrimRight(builder.String(), "\n")
}

func formatInlineMarkdown(text string) string {
	formatted := codePattern.ReplaceAllStringFunc(text, func(match string) string {
		sub := codePattern.FindStringSubmatch(match)
		if len(sub) != 2 {
			return match
		}
		return colorMagenta + "`" + sub[1] + "`" + colorReset
	})

	formatted = boldPattern.ReplaceAllStringFunc(formatted, func(match string) string {
		sub := boldPattern.FindStringSubmatch(match)
		if len(sub) != 2 {
			return match
		}
		return colorBold + sub[1] + colorReset
	})

	return formatted
}

func renderCodeBlockLine(line string, fence bool) string {
	width := terminalWidth()
	inner := line
	if fence {
		inner = colorBlueBold + line
	} else {
		inner = colorCodeFg + line
	}

	padding := width - visibleWidth(line) - 2
	if padding < 0 {
		padding = 0
	}

	return colorCodeBg + "  " + inner + strings.Repeat(" ", padding) + colorReset
}

func terminalWidth() int {
	if raw := strings.TrimSpace(os.Getenv("COLUMNS")); raw != "" {
		if width, err := strconv.Atoi(raw); err == nil && width > 0 {
			return width
		}
	}
	return 80
}

func visibleWidth(text string) int {
	return len(text)
}
