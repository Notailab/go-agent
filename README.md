# go-agent

A small interactive AI agent written in Go.


## Features

- LLM-driven agent core
- Built-in tools:
  - `Bash`
  - `Read_file`
  - `Edit_file`
  - `Write_file`
  - `Memory` for long-term memory entries
- Two-layer memory system:
  - chat history for the ongoing conversation
  - long-term memory for durable facts, preferences, and decisions
- Automatic skill discovery from the `skills` directory
- Interactive CLI entry point in [main.go](main.go)

## Quick Start

### 1. Requirements

- Go 1.20 or newer is recommended
- An LLM service with a compatible API
  - `LLM_BASE_URL`
  - `LLM_API_KEY`
  - `LLM_MODEL`

### 2. Configuration

Create a `.env` file in the project root. You can start from `.env.example`:

```env
LLM_BASE_URL=YOUR_LLM_BASE_URL
LLM_API_KEY=YOUR_LLM_API_KEY
LLM_MODEL=YOUR_LLM_MODEL
```

### 3. Run Demo

Run the demo CLI from the project root:

```bash
go run .
```

The demo will:

- load `.env` if present
- restore chat history from `.memory/HISTORY.jsonl`
- restore long-term memory from `.memory/MEMORY.md`
- start an interactive prompt
- persist both memory stores as the session continues

Type `exit` to quit.


## Tools

To add your own tool, implement the `core.Tool` interface and pass it to `agent.WithTools()`.

Required methods:

- `Name() string`
- `Description() string`
- `Parameters() core.Parameters`
- `Execute(input string) (string, error)`

Minimal example:

```go
type EchoTool struct{}

func (t *EchoTool) Name() string { return "Echo" }

func (t *EchoTool) Description() string { return "Return the input text." }

func (t *EchoTool) Parameters() core.Parameters {
    return core.Parameters{
        Type: "object",
        Properties: map[string]core.Param{
            "text": {Type: "string", Description: "Text to echo back"},
        },
        Required: []string{"text"},
    }
}

func (t *EchoTool) Execute(input string) (string, error) {
    params, err := core.ParseParams(input, "text")
    if err != nil {
        return "", err
    }
    return params["text"].(string), nil
}
```

Register it like this:

```go
agent.WithTools(
    &tools.BashTool{},
    &tools.ReadFileTool{},
    &EchoTool{},
)
```

## Memory

To add your own memory backend, implement `core.ChatMemoryStore` and `core.LongMemoryStore`, then pass them to `core.NewMemory()`.

The chat store requires these methods:

- `Get(int) (core.ChatMessage, error)`
- `Append(core.ChatMessage) error`
- `Update(int, core.ChatMessage) error`
- `Replace(int, int, []core.ChatMessage) error`
- `Delete(int) error`
- `List() ([]core.ChatMessage, error)`
- `Count() (int, error)`
- `Clear() error`
- `Clone() core.ChatMemoryStore`

The long memory store uses the same pattern for strings:

- `Get(int) (string, error)`
- `Append(string) error`
- `Update(int, string) error`
- `Replace(int, int, []string) error`
- `Delete(int) error`
- `List() ([]string, error)`
- `Count() (int, error)`
- `Clear() error`
- `Clone() core.LongMemoryStore`

Example:

```go
memory := core.NewMemory(
    storage.NewFileChatStore(".memory/HISTORY.jsonl"),
    storage.NewFileLongStore(".memory/MEMORY.md"),
)
```

The built-in `Memory` tool manages long-term memory entries with `create`, `update`, and `delete` operations.

## Skills

Skills are registered through `agent.WithSkills()`. Pass one or more skill directories, and the agent will discover skills automatically.

Example:

```go
agent.WithSkills("skills")
```

Each skill is a folder that contains a `SKILL.md` file, such as `skills/weather/SKILL.md`.
The skill loader scans the provided directories, reads each `SKILL.md` frontmatter, and includes that metadata in the system prompt.
