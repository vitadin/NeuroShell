# NeuroShell

[![CircleCI](https://dl.circleci.com/status-badge/img/circleci/RjkGUoMoHBKh13iJmkaXTF/Pfh8uXKBx5881azXtzYpio/tree/main.svg?style=svg&circle-token=CCIPRJ_UA4CCuNuLUnf978JvYtAzb_299edc3497615f5f30d9653830038654df0c471b)](https://dl.circleci.com/status-badge/redirect/circleci/RjkGUoMoHBKh13iJmkaXTF/Pfh8uXKBx5881azXtzYpio/tree/main)
[![Go Report Card](https://goreportcard.com/badge/github.com/vitadin/NeuroShell)](https://goreportcard.com/report/github.com/vitadin/NeuroShell)

![Neuro Logo](./asset/neurologo.svg)

A specialized shell environment designed for seamless interaction with LLM agents. NeuroShell bridges traditional command-line interfaces and modern AI assistants, enabling professionals to create reproducible, automated workflows that combine system operations with LLM capabilities.

## Overview

NeuroShell treats LLM interactions as first-class operations, similar to how traditional shells handle file operations and process management. It's designed for data scientists, software engineers, researchers, and professionals who need reproducible AI-assisted workflows.

## Key Features

- **Command System**: Intuitive `\command[options] message` syntax
- **Variable System**: Session-scoped variables with powerful interpolation
- **Session Management**: Create, save, load, and manage conversation contexts
- **Script Execution**: Run `.neuro` files for reproducible workflows
- **Shell Integration**: Execute system commands seamlessly
- **History Tracking**: Access conversation and command history

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/your-org/NeuroShell.git
cd NeuroShell

# Build the application
just build

# Run NeuroShell
./bin/neuro
```

### Basic Usage

```bash
# Start NeuroShell
./bin/neuro

# Get help with styled formatting
\help[styled=true]

# Create a session with context
\session-new[system="You are a helpful assistant"] my-work

# Send a message to the LLM
\send Hello, how can you help me today?

# Set a variable
\set[name="Alice"] 

# Use variables in messages with styling
\render[style=info] Current user: ${name}
\send My name is ${name}, can you remember that?

# Execute system commands
\bash ls -la

# Access command output with enhanced formatting
\render[style=success] Directory listing complete: ${_output}
\send The current directory contains: ${_output}

# List your variables
\vars[type=user]

# Manage your sessions
\session-list[sort=updated]
```

## Command Reference

### Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `\send` | Send message to LLM agent | `\send Explain quantum computing` |
| `\set[var=value]` | Set variable | `\set[project="MyApp"]` |
| `\get[var]` | Get variable value | `\get[project]` |
| `\bash` | Execute system command | `\bash python script.py` |
| `\vars` | List variables with filtering | `\vars[type=user]` |
| `\help[styled=true] [command]` | Show command help | `\help[styled=true,set]` |

### Session Management

| Command | Description | Example |
|---------|-------------|---------|
| `\session-new[system=prompt]` | Create new chat session | `\session-new[system="You are a code reviewer"] work` |
| `\session-list[sort=name]` | List all sessions | `\session-list[sort=updated]` |
| `\session-delete` | Delete a session | `\session-delete work` |

### Text Processing & Styling

| Command | Description | Example |
|---------|-------------|---------|
| `\echo[to=var] [silent=true]` | Output text with variable expansion | `\echo[to=result] Hello ${name}` |
| `\render[style=bold]` | Style text with colors/formatting | `\render[keywords=[\get,\set]] Use commands` |
| `\editor` | Open external editor | `\editor` |

### Development & Testing

| Command | Description | Example |
|---------|-------------|---------|
| `\run` | Execute .neuro script file | `\run analysis.neuro` |
| `\assert-equal[expect=val]` | Compare values for testing | `\assert-equal[expect=5, actual=${result}]` |
| `\exit` | Exit the shell | `\exit` |

### Command Syntax

```bash
\command[option1, option2, key=value] message text
```

### Getting Help

NeuroShell includes a comprehensive help system with both plain text and styled output:

```bash
# Show all available commands
\help

# Show styled help with colors and formatting
\help[styled=true]

# Get detailed help for a specific command
\help[set]
\help[styled=true,session-new]

# Both bracket and space syntax work
\help[styled=true] render
```

The styled help provides professional formatting with:
- **Color-coded command syntax**
- **Bordered sections** for better readability
- **Keyword highlighting** for NeuroShell commands
- **Comprehensive examples** and usage notes

## Session Management

NeuroShell provides powerful session management capabilities for organizing your LLM conversations:

### Creating Sessions

```bash
# Create a basic session
\session-new work

# Create session with custom system prompt
\session-new[system="You are a helpful coding assistant"] development

# Create session with variable interpolation
\session-new analysis-${@date}

# Session names can include spaces
\session-new "project planning session"
```

### Managing Sessions

```bash
# List all sessions (default: sorted by creation time)
\session-list

# List sessions sorted by name
\session-list[sort=name]

# List sessions sorted by last update
\session-list[sort=updated]

# Show only the currently active session
\session-list[filter=active]

# Delete sessions by name or ID
\session-delete work
\session-delete abc123-uuid
\session-delete[name="project planning"]
```

### Session Variables

When you create or switch sessions, NeuroShell automatically sets metadata variables:

- `${#session_id}` - Unique session identifier
- `${#session_name}` - Human-readable session name
- `${#message_count}` - Number of messages in current session

```bash
# Use session variables in commands
\echo Current session: ${#session_name} (ID: ${#session_id})
\bash mkdir -p "session_${#session_id}"
```

## Variable System

NeuroShell supports multiple types of variables with different prefixes:

### Variable Types

- **User Variables**: `${name}`, `${project}` - Your custom variables
- **Message History**: `${1}` (latest agent response), `${2}` (previous user message)
- **Command Outputs**: `${_output}`, `${_error}`, `${_status}`, `${_elapsed}`
- **System Variables**: `${@pwd}`, `${@user}`, `${@home}`, `${@date}`, `${@os}`
- **Metadata**: `${#session_id}`, `${#session_name}`, `${#message_count}`, `${#test_mode}`

### Variable Interpolation

Variables are automatically interpolated in commands and messages:

```bash
\set[data_file="sales_2024.csv"]
\set[model="gpt-4"]
\bash python analyze.py ${data_file}
\send[model=${model}] Analyze this data: ${_output}
```

## Text Processing & Styling

NeuroShell includes powerful text processing and styling capabilities for enhanced output:

### Text Rendering

The `\render` command provides professional text styling using the lipgloss library:

```bash
# Basic styling
\render[style=bold] Important message
\render[style=success] Operation completed successfully!
\render[style=error] Something went wrong

# Keyword highlighting for NeuroShell commands
\render[keywords=[\get,\set,\send]] Use \get and \set before \send
\render[keywords=[\session-new,\help]] Try \session-new then \help

# Custom colors and themes
\render[color=#FF5733, background=#000000] Custom colored text
\render[theme=dark, style=bold] Dark theme styling

# Store output in variables
\render[to=styled_msg, silent=true] Formatted content
\echo ${styled_msg}
```

### Text Output Options

```bash
# Echo with variable expansion
\echo Hello ${name}, welcome to session ${#session_name}

# Silent output (store in variable without displaying)
\echo[to=result, silent=true] Processed: ${data}

# Raw mode (preserves escape sequences)
\echo[raw=true] Line 1\nLine 2\tTabbed content
```

### External Editor Integration

```bash
# Open your configured editor for multi-line input
\editor

# Configure your preferred editor
\set[@editor="code --wait"]  # VS Code
\set[@editor="nano"]         # Nano
\set[@editor="vim"]          # Vim

# Use editor content
\editor
\send ${_output}  # Send the editor content to LLM
```

## Scripting with .neuro Files

Create reproducible workflows by saving commands in `.neuro` script files:

```bash
%% analysis.neuro - Sales data analysis workflow
%% Comments start with %% and are ignored during execution

\set[data="sales_2024.csv"]
\set[output_dir="results_${@date}"]

%% Create session for this analysis
\session-new[system="You are a data analyst expert"] sales-analysis-${@date}

%% Preprocess data
\bash mkdir -p ${output_dir}
\bash python preprocess.py ${data} --output ${output_dir}/clean_data.csv

%% Show styled progress message
\render[style=success] Data preprocessing completed: ${_output}

%% Analyze with LLM
\send Analyze the processed sales data. Focus on trends, anomalies, and actionable insights.

%% Save analysis results
\set[analysis="${1}"]
\echo[to=report_file, silent=true] ${output_dir}/analysis_report.txt
\bash echo "${analysis}" > ${report_file}

%% Create styled summary
\render[style=bold, keywords=[\session-list]] Analysis saved! Use \session-list to see all sessions.
```

Run scripts with:
```bash
./bin/neuro script analysis.neuro
```

## Example Workflows

### Data Analysis Pipeline with Sessions

```bash
# Create dedicated analysis session
\session-new[system="You are a data scientist specializing in customer analytics"] customer-analysis

# Set up data sources with styled output
\set[raw_data="customer_data.csv"]
\set[analysis_date="${@date}"]
\render[style=info] Starting analysis for ${analysis_date}

# Process data
\bash python clean_data.py ${raw_data}
\render[style=success] Data cleaned: ${_output}

# Analyze with context-aware LLM
\send Please analyze this cleaned customer dataset focusing on behavior patterns and segmentation opportunities: ${_output}

# Generate styled report
\set[insights="${1}"]
\render[to=report_header, silent=true] # Customer Analysis Report - ${analysis_date}
\echo[to=full_report, silent=true] ${report_header}\n\n## Key Insights\n\n${insights}
\bash echo "${full_report}" > customer_analysis_${analysis_date}.md

# Show completion status
\render[style=success, keywords=[\session-list]] Analysis complete! Use \session-list to manage your sessions.
```

### Code Review Workflow with Enhanced Formatting

```bash
# Create code review session
\session-new[system="You are a senior software engineer conducting code reviews"] code-review-${@date}

# Get recent changes with context
\bash git log --oneline -5
\render[style=info] Recent commits: ${_output}

\bash git diff HEAD~1
\render[style=bold] Reviewing changes: ${_output}

# Comprehensive review with LLM
\send Please conduct a thorough code review of these changes, focusing on:
1. Code quality and best practices
2. Potential bugs or security issues  
3. Performance implications
4. Maintainability concerns

Changes to review: ${_output}

# Format and save feedback
\set[review="${1}"]
\render[style=warning, to=review_header, silent=true] ## Code Review - ${@date}
\echo[to=formatted_review, silent=true] ${review_header}\n\n${review}
\bash echo "${formatted_review}" > code_review_${@date}.md

# Create actionable summary
\send Based on your review, create a prioritized action item list for the developer.

\set[action_items="${1}"]
\render[style=success] Review saved to code_review_${@date}.md
\render[keywords=[\help]] Use \help for more commands
```

### Multi-Session Research Workflow

```bash
# Create separate sessions for different research aspects
\session-new[system="You are a technical researcher"] tech-research
\send Research the latest developments in AI model architectures

# Switch context for market analysis
\session-new[system="You are a market analyst"] market-research  
\send Analyze the commercial adoption trends of AI technologies

# Compare insights across sessions
\session-list[sort=updated]
\render[style=info, keywords=[\session-list]] Created ${#message_count} sessions for comprehensive research
```

## Development

### Building

```bash
# Install dependencies
go mod download

# Run tests
just test

# Build binary
just build

# Run development version
just run
```

### Testing

The project has comprehensive test coverage:

```bash
# Run all tests with coverage
just test

# Run specific test suites
just test-unit          # Service and utility tests
just test-parser        # Parser tests (100% coverage)
just test-context       # Context tests (100% coverage)
just test-shell         # Shell integration tests
just test-commands      # Command implementation tests
just test-bench         # Performance benchmarks
```

### Project Structure

```
NeuroShell/
├── cmd/                    # Main applications
│   ├── neuro/             # Primary CLI
│   └── neurotest/         # Test utilities
├── internal/              # Private application code
│   ├── commands/          # Command implementations
│   ├── context/           # Session context management
│   ├── parser/            # Command syntax parsing
│   ├── services/          # Core business logic
│   ├── shell/             # Shell integration
│   └── testutils/         # Testing utilities
├── pkg/                   # Public API
│   └── types/             # Shared types
└── scripts/               # Development scripts
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run the test suite (`just test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Guidelines

- Maintain test coverage above 90%
- Follow Go best practices and conventions
- Add tests for new features and bug fixes
- Update documentation for API changes

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- **Phase 1** (Current): Core shell functionality and command system
- **Phase 2**: Multi-agent support and advanced templating
- **Phase 3**: Plugin system and cloud session synchronization

## Support

- Create an issue for bug reports or feature requests
- Check existing issues before creating new ones
- Provide detailed reproduction steps for bugs

---

*NeuroShell: Where traditional shell power meets AI intelligence.*
