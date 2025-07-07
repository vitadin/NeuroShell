// Package embedded provides access to embedded theme configuration files.
package embedded

import _ "embed"

// DefaultThemeData contains the embedded default theme YAML data.
//
//go:embed themes/default.yaml
var DefaultThemeData []byte

// DarkThemeData contains the embedded dark theme YAML data.
//
//go:embed themes/dark.yaml
var DarkThemeData []byte

// LightThemeData contains the embedded light theme YAML data.
//
//go:embed themes/light.yaml
var LightThemeData []byte

// PlainThemeData contains the embedded plain theme YAML data.
//
//go:embed themes/plain.yaml
var PlainThemeData []byte
