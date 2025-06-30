# NeuroShell

[![CircleCI](https://dl.circleci.com/status-badge/img/circleci/RjkGUoMoHBKh13iJmkaXTF/Pfh8uXKBx5881azXtzYpio/tree/main.svg?style=svg&circle-token=CCIPRJ_UA4CCuNuLUnf978JvYtAzb_299edc3497615f5f30d9653830038654df0c471b)](https://dl.circleci.com/status-badge/redirect/circleci/RjkGUoMoHBKh13iJmkaXTF/Pfh8uXKBx5881azXtzYpio/tree/main)

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

# Send a message to the LLM
\send Hello, how can you help me today?

# Set a variable
\set[name="Alice"] 

# Use variables in messages
\send My name is ${name}, can you remember that?

# Execute system commands
\bash ls -la

# Access command output
\send The current directory contains: ${_output}

# Save your session
\save[name="my_session"]
```

## Command Reference

### Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `\send` | Send message to LLM agent | `\send Explain quantum computing` |
| `\set[var=value]` | Set variable | `\set[project="MyApp"]` |
| `\get[var]` | Get variable value | `\get[project]` |
| `\bash` | Execute system command | `\bash python script.py` |
| `\history[n=5]` | View recent exchanges | `\history[n=10]` |
| `\list` | List all variables | `\list` |
| `\save[name="session"]` | Save session state | `\save[name="analysis"]` |
| `\load[name="session"]` | Load saved session | `\load[name="analysis"]` |
| `\clear` | Clear current session | `\clear` |
| `\help[command]` | Show command help | `\help[set]` |

### Command Syntax

```bash
\command[option1, option2, key=value] message text
```

## Variable System

NeuroShell supports multiple types of variables with different prefixes:

### Variable Types

- **User Variables**: `${name}`, `${project}` - Your custom variables
- **Message History**: `${1}` (latest agent response), `${2}` (previous user message)
- **Command Outputs**: `${_output}`, `${_error}`, `${_status}`, `${_elapsed}`
- **System Variables**: `${@pwd}`, `${@user}`, `${@home}`, `${@date}`, `${@os}`
- **Metadata**: `${#session_id}`, `${#message_count}`, `${#test_mode}`

### Variable Interpolation

Variables are automatically interpolated in commands and messages:

```bash
\set[data_file="sales_2024.csv"]
\set[model="gpt-4"]
\bash python analyze.py ${data_file}
\send[model=${model}] Analyze this data: ${_output}
```

## Scripting with .neuro Files

Create reproducible workflows by saving commands in `.neuro` script files:

```bash
# analysis.neuro
\set[data="sales_2024.csv"]
\set[output_dir="results_${@date}"]

# Preprocess data
\bash mkdir -p ${output_dir}
\bash python preprocess.py ${data} --output ${output_dir}/clean_data.csv

# Analyze with LLM
\send Analyze the processed sales data in ${_output}. Focus on trends and anomalies.

# Save analysis
\set[analysis="${1}"]
\bash echo "${analysis}" > ${output_dir}/analysis.txt

# Save session for later
\save[name="sales_analysis_${@date}"]
```

Run scripts with:
```bash
./bin/neuro script analysis.neuro
```

## Example Workflows

### Data Analysis Pipeline

```bash
# Set up data sources
\set[raw_data="customer_data.csv"]
\set[analysis_date="${@date}"]

# Process data
\bash python clean_data.py ${raw_data}
\send Please analyze this cleaned dataset: ${_output}

# Generate report
\set[insights="${1}"]
\bash echo "# Analysis Report - ${analysis_date}\n\n${insights}" > report.md

# Save session
\save[name="customer_analysis_${analysis_date}"]
```

### Code Review Workflow

```bash
# Get recent changes
\bash git diff HEAD~1

# Review with LLM
\send Please review these code changes: ${_output}

# Save feedback
\set[review="${1}"]
\bash echo "${review}" > code_review.txt

# Create summary
\send Summarize the key issues from your review
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