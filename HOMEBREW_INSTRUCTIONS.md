# How to Set Up Homebrew Installation for NeuroShell

Follow these steps to enable `brew install` for NeuroShell:

## Step 1: Create the Homebrew Tap Repository

1. Go to GitHub and create a new repository
2. **Repository name**: `homebrew-neuroshell` (must start with `homebrew-`)
3. **Visibility**: Public
4. **Description**: "Homebrew tap for NeuroShell - A specialized shell environment for LLM agents"
5. Initialize with README

## Step 2: Set Up the Tap Repository

1. Clone your new `homebrew-neuroshell` repository:
   ```bash
   git clone https://github.com/vitadin/homebrew-neuroshell.git
   cd homebrew-neuroshell
   ```

2. Create the Formula directory and copy the formula:
   ```bash
   mkdir Formula
   # Copy the neuroshell.rb file from this repository
   cp /path/to/NeuroShell/Formula/neuroshell.rb Formula/
   ```

3. Replace the README with the tap-specific one:
   ```bash
   # Copy the homebrew-README.md file from this repository
   cp /path/to/NeuroShell/homebrew-README.md README.md
   ```

4. Commit and push:
   ```bash
   git add .
   git commit -m "Add NeuroShell formula for Homebrew"
   git push origin main
   ```

## Step 3: Make NeuroShell Repository Public

1. Go to your main NeuroShell repository settings
2. Scroll down to "Danger Zone"
3. Click "Change repository visibility"
4. Select "Make public"
5. Confirm the change

## Step 4: Test the Installation

Once both repositories are public, test the installation:

```bash
# Test the tap installation
brew tap vitadin/neuroshell

# Test the package installation  
brew install neuroshell

# Test that it works
neuroshell --version

# Clean up for further testing
brew uninstall neuroshell
brew untap vitadin/neuroshell
```

## Step 5: Update Documentation

Add installation instructions to your main README.md:

```markdown
## Installation

### Homebrew (macOS/Linux)

```bash
brew tap vitadin/neuroshell
brew install neuroshell
```

### From Source

```bash
git clone https://github.com/vitadin/NeuroShell.git
cd NeuroShell
just build
./bin/neuro
```
```

## Step 6: For Future Releases

When you release new versions:

1. The formula will automatically use the latest release from your main repository
2. If needed, you can update the formula in the `homebrew-neuroshell` repository
3. Homebrew will handle version updates automatically

## Files Created

This setup created the following files in your NeuroShell repository:

- `Formula/neuroshell.rb` - The Homebrew formula
- `BREW_SETUP.md` - Detailed setup documentation
- `homebrew-README.md` - README for the tap repository
- `HOMEBREW_INSTRUCTIONS.md` - These step-by-step instructions

## Final Result

Users will be able to install NeuroShell with:

```bash
brew install vitadin/neuroshell/neuroshell
```

Or after adding the tap:

```bash
brew tap vitadin/neuroshell
brew install neuroshell
```