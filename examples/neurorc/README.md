# NeuroShell .neurorc Examples

This directory contains example `.neurorc` files that demonstrate various auto-startup configurations for NeuroShell.

## What is .neurorc?

`.neurorc` is NeuroShell's auto-startup script system. When you start NeuroShell, it automatically looks for and executes `.neurorc` files in this priority order:

1. **Current directory** `.neurorc` (highest priority)
2. **Home directory** `~/.neurorc` (fallback)

## CLI Control Options

You can control `.neurorc` behavior with these CLI flags:

- `--no-rc`: Skip all `.neurorc` files
- `--rc-file=path`: Use a specific startup script instead of `.neurorc`
- `--confirm-rc`: Prompt before executing any `.neurorc` file

Environment variables:
- `NEURO_RC=0`: Disable `.neurorc` processing
- `NEURO_RC_FILE=path`: Use specific startup script

## Example Files

### `minimal.neurorc`
The absolute minimum `.neurorc` setup:
- Creates a basic session
- Simple welcome message

**Use case**: When you want auto-startup but with minimal configuration.

### `basic-startup.neurorc`
Standard personal workspace setup:
- Sets up default variables and theme
- Creates a daily work session
- Configures styling preferences

**Use case**: Personal daily workflow with consistent environment setup.

### `development-setup.neurorc`
Development-focused configuration:
- Creates specialized development session with code review prompts
- Sets up project-specific variables
- Loads Git branch information
- Development-friendly styling

**Use case**: Software development work, code reviews, debugging sessions.

### `research-workflow.neurorc`
Academic and research-focused setup:
- Creates research assistant session
- Sets up citation and analysis preferences
- Research-friendly styling and environment
- Helpful tips for research workflows

**Use case**: Academic research, literature reviews, data analysis, writing.

## Usage

1. Copy any example file to your desired location:
   ```bash
   # For project-specific startup
   cp examples/neurorc/development-setup.neurorc .neurorc
   
   # For global startup
   cp examples/neurorc/basic-startup.neurorc ~/.neurorc
   ```

2. Customize the variables and sessions to match your workflow

3. Start NeuroShell - your `.neurorc` will execute automatically!

## Creating Custom .neurorc Files

Your `.neurorc` file can contain any valid NeuroShell commands:

```neuro
# Set variables
\set[my_var="value"]

# Create sessions
\session-new[system="Custom system prompt"] my_session

# Execute commands
\echo Welcome to my custom setup!

# Load other scripts
\run path/to/other-script.neuro
```

## Testing Your .neurorc

Use the built-in testing tools to verify your `.neurorc` works correctly:

```bash
# Test with confirmation prompt
neuro shell --confirm-rc

# Test with specific file
neuro shell --rc-file=my-custom-startup.neuro

# Disable for troubleshooting
neuro shell --no-rc
```

## Best Practices

1. **Keep it fast**: Avoid slow operations in `.neurorc` to maintain quick startup
2. **Use variables**: Leverage system variables like `${@user}`, `${@date}`, `${@pwd}`
3. **Comment your code**: Document what each section does for future reference
4. **Test regularly**: Ensure your `.neurorc` works after modifications
5. **Version control**: Consider adding `.neurorc` to your project repositories for team consistency