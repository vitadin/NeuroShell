package output

import "io"

// Option is a functional option for configuring Printer instances.
type Option func(*Printer)

// WithStyles configures the printer to use the provided StyleProvider for styling.
// If the provider is nil or not available, the printer will fall back to plain text.
func WithStyles(provider StyleProvider) Option {
	return func(p *Printer) {
		if provider != nil && provider.IsAvailable() {
			p.styleProvider = provider
		}
	}
}

// WithWriter configures the printer to write output to the specified writer.
// Default is os.Stdout if not specified.
func WithWriter(writer io.Writer) Option {
	return func(p *Printer) {
		if writer != nil {
			p.writer = writer
		}
	}
}

// WithMode configures the printer to operate in a specific output mode.
func WithMode(mode Mode) Option {
	return func(p *Printer) {
		p.mode = mode
	}
}

// PlainText forces the printer to use plain text output, ignoring any StyleProvider.
// This is useful for machine-readable output or when styling should be disabled.
func PlainText() Option {
	return func(p *Printer) {
		p.mode = ModePlain
		p.forcePlain = true
	}
}

// JSON configures the printer for structured JSON output.
// This is useful for scripting and automation scenarios.
func JSON() Option {
	return func(p *Printer) {
		p.mode = ModeJSON
	}
}

// TestMode configures the printer for deterministic output in tests.
// This ensures consistent output regardless of terminal capabilities.
func TestMode() Option {
	return func(p *Printer) {
		p.testMode = true
		p.mode = ModePlain
		p.forcePlain = true
	}
}

// Silent configures the printer to suppress all output.
// This is useful when you want to capture output without displaying it.
func Silent() Option {
	return func(p *Printer) {
		p.silent = true
	}
}

// WithPrefix adds a prefix to all output from this printer.
// This is useful for component-specific output identification.
func WithPrefix(prefix string) Option {
	return func(p *Printer) {
		p.prefix = prefix
	}
}
