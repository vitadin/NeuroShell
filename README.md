# NeuroShell

[![CircleCI](https://dl.circleci.com/status-badge/img/circleci/RjkGUoMoHBKh13iJmkaXTF/Pfh8uXKBx5881azXtzYpio/tree/main.svg?style=svg&circle-token=CCIPRJ_UA4CCuNuLUnf978JvYtAzb_299edc3497615f5f30d9653830038654df0c471b)](https://dl.circleci.com/status-badge/redirect/circleci/RjkGUoMoHBKh13iJmkaXTF/Pfh8uXKBx5881azXtzYpio/tree/main)
[![Go Report Card](https://goreportcard.com/badge/github.com/vitadin/NeuroShell)](https://goreportcard.com/report/github.com/vitadin/NeuroShell)
[![Release](https://github.com/vitadin/NeuroShell/actions/workflows/release.yml/badge.svg)](https://github.com/vitadin/NeuroShell/actions/workflows/release.yml)
[![License: LGPL v3](https://img.shields.io/badge/License-LGPL_v3-blue.svg)](https://www.gnu.org/licenses/lgpl-3.0)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vitadin/NeuroShell)](https://github.com/vitadin/NeuroShell)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/vitadin/NeuroShell)](https://github.com/vitadin/NeuroShell/releases/latest)

> [!IMPORTANT]
> **⚠️ Early Development Warning**
>
> NeuroShell is currently in early development (version < 1.0). While we strive to minimize disruption, **breaking changes may occur between releases**. We recommend:
>
> - 🚫 **Not recommended for production use** - use for development/experimentation only
> - 📖 **Check release notes** before updating
> - 💾 **Backup your `.neuro` scripts** and session data
> - 🐛 **Report issues** on [GitHub Issues](https://github.com/vitadin/NeuroShell/issues)
>
> Follow our [releases](https://github.com/vitadin/NeuroShell/releases) for updates and migration guides.

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

#### Homebrew (macOS/Linux) - Recommended

```bash
# Add the tap and install
brew tap vitadin/neuroshell
brew install neuroshell

# Run NeuroShell
neuroshell
```

#### From Source

```bash
# Clone and build
git clone https://github.com/vitadin/NeuroShell.git
cd NeuroShell
just build

# Run NeuroShell
./bin/neuro
```

### Getting Version

```
\version
```

### Getting Help

```
\help                    
\help session-new        
\model-catalog           
\provider-catalog        
```

Commands explained:
- `\help` - List all commands
- `\help session-new` - Command-specific help  
- `\model-catalog` - Available LLM models
- `\provider-catalog` - Supported providers

### Getting LLM Models

```
\model-catalog
```

### Basic Usage

Start with simple LLM interaction:
```
\model-new[catalog_id="O4MC"] gpt-model
\send Hello, can you help me with Python programming?
```

Create a specialized session:
```
\model-new[catalog_id="O4MC"] gpt-model
\session-new[system="You are a code reviewer"] review-session
```

Use variables for context:
```
\model-new[catalog_id="O4MC"] gpt-model
\set[project="MyApp"]
\send I'm working on ${project}. Can you help me optimize this code?
```

Execute system commands and use their output:
```
\model-new[catalog_id="O4MC"] gpt-model
\bash git status
\send Based on this git status: ${_output}, what should I do next?
```

Save your work:
```
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
\model-new[catalog_id="CS4", thinking_budget=1000] claude-sonnet
\send Hello, Claude!
```

### OpenAI GPT
```bash
\model-new[catalog_id="O4MC"] gpt-model
\send Hello, GPT!
```

### Google Gemini
```bash
\model-new[catalog_id="GM25F", thinking_budget=2048] gemini-model
\send Hello, Gemini!
```

## Variable Types

- **User Variables**: `${name}`, `${project}` - Your custom variables
- **Message History**: `${1}` (latest response), `${2}` (previous user message)
- **Command Output**: `${_output}`, `${_error}`, `${_status}`
- **System Info**: `${@user}`, `${@date}`, `${@pwd}`
- **Session Metadata**: `${#session_name}`, `${#active_model_name}`

## Comments

Use `%%` to comment entire lines in NeuroShell scripts:

```
%% This is a comment line
\send Hello world
%% Another comment
\set[name="value"]
```

## Script Files (.neuro)

Create reproducible workflows in `.neuro` files:

Example `analysis.neuro`:
```
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
```
\session-new[system="You are a data analyst"] analysis
\bash head -5 data.csv
\send What insights can you find in this data: ${_output}
\bash python analyze.py > report.txt
\send Summarize this analysis: ${_output}
```

### Code Review
```
\session-new[system="You are a senior software engineer"] review
\bash git diff HEAD~1
\send Please review these changes: ${_output}
\send What improvements would you suggest?
```


## License

LGPL v3 - see [LICENSE](LICENSE) for details.

---

*NeuroShell: Where traditional shell power meets AI intelligence.*
