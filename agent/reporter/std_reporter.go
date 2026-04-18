package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/Notailab/go-agent/agent/agent"
)

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

func (r *StdoutReporter) BeforeLLM(ctx agent.HookContext) {
	r.mu.Lock()
	r.streamBuffer.Reset()
	r.streamInCodeBlock = false
	r.mu.Unlock()

	if ctx.Stream {
		r.printAssistantHeader()
	}
}

func (r *StdoutReporter) OnLLM(ctx agent.HookContext) {
	if strings.TrimSpace(ctx.Delta) == "" {
		return
	}

	r.mu.Lock()
	r.streamBuffer.WriteString(ctx.Delta)
	r.flushStreamLocked(false)
	r.mu.Unlock()
}

func (r *StdoutReporter) AfterLLM(ctx agent.HookContext) {
	if ctx.Error != nil {
		r.Errorf("LLM error: %v", ctx.Error)
		return
	}

	if ctx.Stream {
		r.mu.Lock()
		r.flushStreamLocked(true)
		r.mu.Unlock()
		fmt.Println()
		return
	}

	content := ""
	if len(ctx.Result.Choices) > 0 {
		content = strings.TrimSpace(ctx.Result.Choices[0].Message.Content)
	}
	if content == "" {
		return
	}

	r.printAssistantHeader()
	fmt.Println(renderMarkdownBlock(content))
}

func (r *StdoutReporter) BeforeTool(ctx agent.HookContext) {
	formattedArgs := summarizeInlineArgs(ctx.ToolCall.Function.Arguments)
	if formattedArgs == "" {
		formattedArgs = "(no arguments)"
	}

	fmt.Printf("%s● Tool%s %s%s%s %s%s%s\n", colorCyanBold, colorReset, colorYellowBold, ctx.ToolCall.Function.Name, colorReset, colorMagenta, formattedArgs, colorReset)
}

func (r *StdoutReporter) AfterTool(ctx agent.HookContext) {
	if ctx.Error != nil {
		r.Errorf("tool %s failed: %v", ctx.ToolCall.Function.Name, ctx.Error)
	}
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
		rendered, keepOpen := renderMarkdownLine(line, r.streamInCodeBlock)
		if rendered == "" && !keepOpen {
			builder.WriteString("\n")
			continue
		}
		builder.WriteString(rendered)
		if !isLast || final {
			builder.WriteString("\n")
		}
		r.streamInCodeBlock = keepOpen
	}

	return builder.String()
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
		rendered, keepOpen := renderMarkdownLine(line, inCodeBlock)
		if rendered == "" && !keepOpen {
			builder.WriteString("\n")
			continue
		}
		builder.WriteString(rendered)
		builder.WriteString("\n")
		inCodeBlock = keepOpen
	}

	return strings.TrimRight(builder.String(), "\n")
}

func renderMarkdownLine(line string, inCodeBlock bool) (string, bool) {
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "```") {
		if inCodeBlock {
			return renderCodeBlockLine(line, true), false
		}
		return renderCodeBlockLine(line, true), true
	}

	if inCodeBlock {
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
