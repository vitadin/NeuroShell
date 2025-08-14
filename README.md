# NeuroShell

[![CircleCI](https://dl.circleci.com/status-badge/img/circleci/RjkGUoMoHBKh13iJmkaXTF/Pfh8uXKBx5881azXtzYpio/tree/main.svg?style=svg&circle-token=CCIPRJ_UA4CCuNuLUnf978JvYtAzb_299edc3497615f5f30d9653830038654df0c471b)](https://dl.circleci.com/status-badge/redirect/circleci/RjkGUoMoHBKh13iJmkaXTF/Pfh8uXKBx5881azXtzYpio/tree/main)
[![Go Report Card](https://goreportcard.com/badge/github.com/vitadin/NeuroShell)](https://goreportcard.com/report/github.com/vitadin/NeuroShell)

![Neuro Logo](./assets/neurologo.svg)

A specialized shell environment for seamless interaction with LLM agents. NeuroShell bridges traditional command-line interfaces and modern AI assistants, enabling professionals to create reproducible, automated workflows.

## Features

- **Multi-Provider LLM Support**: Anthropic Claude, OpenAI GPT/o1, Google Gemini
- **Advanced Session Management**: Create, activate, copy, edit, export/import conversations
- **Thinking Blocks**: Visual reasoning display for supported models
- **Variable System**: User, system, command output, and metadata variables with interpolation
- **Script Execution**: Reproducible workflows with `.neuro` files
- **Shell Integration**: Execute system commands with output capture

## Quick Start

### Installation

```bash
# Clone and build
git clone https://github.com/your-org/NeuroShell.git
cd NeuroShell
just build

# Run NeuroShell
./bin/neuro
```

### Basic Usage

```bash
# Start with simple LLM interaction
\send Hello, can you help me with Python programming?

# Create a specialized session
\session-new[system="You are a code reviewer"] review-session

# Use variables
\set[project="MyApp"]
\send I'm working on ${project}. Can you help me optimize this code?

# Execute system commands
\bash git status
\send Based on this git status: ${_output}, what should I do next?

# Save your work
\session-export analysis_results.json
```

## Core Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `\send` | Send message to LLM | `\send Explain machine learning` |
| `\session-new` | Create conversation session | `\session-new[system="You are a data scientist"] analysis` |
| `\model-new` | Create and configure LLM model | `\model-new[catalog_id="CS4"] claude-model` |
| `\bash` | Execute system command | `\bash python analyze.py` |
| `\set` / `\get` | Variable management | `\set[data="file.csv"]` |
| `\help` | Show command help | `\help session-new` |

## LLM Provider Setup

### Anthropic Claude
```bash
\anthropic-client-new my-claude
\model-new[catalog_id="CS4", thinking_budget=1000] claude-sonnet
\model-activate claude-sonnet
\send Hello from Claude!
```

### OpenAI GPT
```bash
\openai-client-new my-gpt
\model-new[catalog_id="O4MC"] gpt-model
\model-activate gpt-model
\send Hello from GPT!
```

### Google Gemini
```bash
\gemini-client-new my-gemini
\model-new[catalog_id="GM25F", thinking_budget=2048] gemini-model
\model-activate gemini-model
\send Hello from Gemini!
```

## Variable Types

- **User Variables**: `${name}`, `${project}` - Your custom variables
- **Message History**: `${1}` (latest response), `${2}` (previous user message)
- **Command Output**: `${_output}`, `${_error}`, `${_status}`
- **System Info**: `${@user}`, `${@date}`, `${@pwd}`
- **Session Metadata**: `${#session_name}`, `${#active_model_name}`

## Script Files (.neuro)

Create reproducible workflows:

```bash
%% analysis.neuro
\session-new[system="You are a data analyst"] data-session
\set[data_file="sales.csv"]
\bash python preprocess.py ${data_file}
\send Analyze this data: ${_output}
\session-export results_${@date}.json
```

Run with:
```bash
./bin/neuro batch analysis.neuro
```

## Example Workflows

### Data Analysis
```bash
\session-new[system="You are a data analyst"] analysis
\bash head -5 data.csv
\send What insights can you find in this data: ${_output}
\bash python analyze.py > report.txt
\send Summarize this analysis: ${_output}
```

### Code Review
```bash
\session-new[system="You are a senior software engineer"] review
\bash git diff HEAD~1
\send Please review these changes: ${_output}
\send What improvements would you suggest?
```

### Multi-Model Comparison
```bash
# Ask Claude
\anthropic-client-new claude
\model-new[catalog_id="CS4"] claude-model
\model-activate claude-model
\send Explain quantum computing

# Ask GPT
\openai-client-new gpt
\model-new[catalog_id="O4MC"] gpt-model
\model-activate gpt-model
\send Explain quantum computing

# Compare responses
\echo Claude: ${1}
\echo GPT: ${3}
```

## Getting Help

```bash
\help                    # List all commands
\help session-new        # Command-specific help
\model-catalog           # Available LLM models
\provider-catalog        # Supported providers
```

## Development

```bash
# Build and test
just build
just test

# Run specific tests
just test-e2e           # End-to-end tests
just test-unit          # Unit tests
```

## License

LGPL v3 - see [LICENSE](LICENSE) for details.

---

*NeuroShell: Where traditional shell power meets AI intelligence.*