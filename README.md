# go-agent

A small interactive AI Agent written in Go.

## Features

- LLM-driven agent core
- Built-in tools:
  - `Bash`
  - `Read_file`
  - `Edit_file`
  - `Write_file`
- Persistent memory stored in `.memory/HISTORY.jsonl`
- Automatic skill discovery and registration from the `skills` directory
- Demo CLI entry point in `main.go`

---

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
- restore memory from `.memory/HISTORY.jsonl`
- start an interactive prompt
- save memory again when you exit

Type `exit` to quit.

---

## Tools

To add your own tool, implement the `core.Tool` interface and pass it to `agent.WithTools()`.

The interface requires these methods:

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

## Skills

Skills are registered through `agent.WithSkills()`. Pass one or more skill directories, and the agent will discover skills automatically.

Example:

```go
agent.WithSkills("skills")
```

Each skill is a folder that contains a `SKILL.md` file, such as `skills/weather/SKILL.md`.
The skill loader scans the provided directories, reads each `SKILL.md` frontmatter, and includes that metadata in the system prompt.

## Memory

To use a custom memory backend, implement the `core.MemoryStore` interface and pass it to `core.NewMemoryWithStore()`.

The interface requires these methods:

- `Save(messages []core.ChatMessage) error`
- `Load(messages *[]core.ChatMessage) error`

Minimal example:

```go
type CustomStore struct{}

func (s *CustomStore) Save(messages []core.ChatMessage) error {
    return nil
}

func (s *CustomStore) Load(messages *[]core.ChatMessage) error {
    *messages = []core.ChatMessage{}
    return nil
}
```

Use it like this:

```go
memory := core.NewMemoryWithStore(&CustomStore{})
```

By default, the demo CLI uses `core.NewFileBackedMemory(".memory/HISTORY.jsonl")`, and it loads and saves memory automatically.
