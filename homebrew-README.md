# Homebrew Tap for NeuroShell

This is the official Homebrew tap for [NeuroShell](https://github.com/vitadin/NeuroShell), a specialized shell environment for seamless interaction with LLM agents.

## Installation

```bash
# Add the tap and install NeuroShell
brew tap vitadin/neuroshell
brew install neuroshell

# Or install directly in one command
brew install vitadin/neuroshell/neuroshell
```

## Usage

After installation, you can run NeuroShell with:

```bash
neuroshell
```

For help and getting started:

```bash
neuroshell --help
```

## What is NeuroShell?

NeuroShell bridges traditional command-line interfaces and modern AI assistants, enabling professionals to create reproducible, automated workflows with LLM agents.

### Key Features

- **Multi-Provider LLM Support**: Anthropic Claude, OpenAI GPT/o1, Google Gemini
- **Advanced Session Management**: Create, activate, copy, edit, export/import conversations  
- **Variable System**: User, system, command output, and metadata variables with interpolation
- **Script Execution**: Reproducible workflows with `.neuro` files
- **Shell Integration**: Execute system commands with output capture

### Quick Start

```bash
# Start NeuroShell
neuroshell

# Get help
\help

# Create a model and start chatting
\model-new[catalog_id="O4MC"] gpt-model
\send Hello, can you help me with Python programming?
```

## Links

- **Main Repository**: https://github.com/vitadin/NeuroShell
- **Issues**: https://github.com/vitadin/NeuroShell/issues
- **Releases**: https://github.com/vitadin/NeuroShell/releases
- **Documentation**: https://github.com/vitadin/NeuroShell#readme

## License

LGPL v3 - see [LICENSE](https://github.com/vitadin/NeuroShell/blob/main/LICENSE) for details.