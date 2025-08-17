# Homebrew Setup for NeuroShell

This document explains how to set up Homebrew installation for NeuroShell.

## Option 1: Custom Tap (Recommended)

### Step 1: Create a Homebrew Tap Repository

1. Create a new repository named `homebrew-neuroshell` on GitHub
2. The repository should be public
3. Copy the `Formula/neuroshell.rb` file to the new repository

### Step 2: Repository Structure

Your `homebrew-neuroshell` repository should have this structure:
```
homebrew-neuroshell/
├── README.md
├── Formula/
│   └── neuroshell.rb
```

### Step 3: Users Install Via Tap

Once set up, users can install NeuroShell with:

```bash
# Add the tap
brew tap vitadin/neuroshell

# Install NeuroShell
brew install neuroshell

# Or in one command
brew install vitadin/neuroshell/neuroshell
```

## Option 2: Direct Formula Installation

Users can also install directly from the formula:

```bash
brew install https://raw.githubusercontent.com/vitadin/homebrew-neuroshell/main/Formula/neuroshell.rb
```

## Formula Details

The formula (`Formula/neuroshell.rb`):
- Downloads source code from GitHub releases
- Uses `just` and `go` to build the project
- Installs the `neuro` binary as `neuroshell`
- Includes basic version test

## Release Process

When you create new releases:

1. Update the `url` in the formula to point to the new release tag
2. Remove or update the `sha256` (Homebrew will calculate it automatically)
3. Commit and push the formula changes

## Making It Public

Before users can install via Homebrew:
1. Make the main NeuroShell repository public
2. Create and make the `homebrew-neuroshell` repository public
3. Ensure releases are properly tagged (like `v0.2.4`)

## Testing

Test the formula locally:
```bash
brew install --build-from-source ./Formula/neuroshell.rb
brew test neuroshell
brew uninstall neuroshell
```