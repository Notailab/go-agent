package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Notailab/go-agent/agent/agent"
	"github.com/Notailab/go-agent/agent/core"
	"github.com/Notailab/go-agent/agent/storage"
	"github.com/Notailab/go-agent/agent/tools"
)

func LoadEnvFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		os.Setenv(key, value)
	}

	return scanner.Err()
}

func main() {
	if err := LoadEnvFile(".env"); err != nil {
		fmt.Printf("warning: failed to load .env: %v\n", err)
	}

	baseURL := os.Getenv("LLM_BASE_URL")
	apiKey := os.Getenv("LLM_API_KEY")
	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "Qwen/Qwen3.5-35B-A3B"
	}

	if baseURL == "" || apiKey == "" {
		fmt.Println("LLM_BASE_URL or LLM_API_KEY is not set")
		return
	}

	reporter := &agent.StdoutReporter{}

	memory := core.NewMemory(
		storage.NewFileChatStore(".memory/HISTORY.jsonl"),
		storage.NewFileLongStore(".memory/MEMORY.md"),
	)

	agent := agent.NewReactAgent(
		agent.WithLLM(baseURL, model, apiKey),
		agent.WithTools(
			&tools.BashTool{},
			&tools.ReadFileTool{},
			&tools.EditFileTool{},
			&tools.WriteFileTool{},
			tools.NewLongMemoryTool(memory),
		),
		agent.WithMemory(memory),
		agent.WithSkills("skills"),
		agent.WithReporter(reporter),
		agent.WithMaxTokens(12000),
	)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("\n(Cur %v tokens) You: ", agent.CurTokens())
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if strings.ToLower(input) == "exit" {
			fmt.Println("退出程序")
			break
		}

		_, err := agent.StreamRun(context.Background(), input)
		if err != nil {
			reporter.Errorf("Error running agent: %v", err)
			continue
		}
	}
}
