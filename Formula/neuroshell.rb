class Neuroshell < Formula
  desc "A specialized shell environment for seamless interaction with LLM agents"
  homepage "https://github.com/vitadin/NeuroShell"
  url "https://github.com/vitadin/NeuroShell/archive/refs/tags/v0.2.4.tar.gz"
  sha256 "" # This will be calculated automatically by Homebrew
  license "LGPL-3.0"
  head "https://github.com/vitadin/NeuroShell.git", branch: "main"

  depends_on "go" => :build

  def install
    # Set version information
    version_info = version.to_s
    git_commit = `git rev-parse --short HEAD 2>/dev/null || echo unknown`.strip
    build_date = Time.now.utc.strftime("%Y-%m-%d")
    
    # Build flags for version injection
    ldflags = [
      "-X 'neuroshell/internal/version.Version=#{version_info}'",
      "-X 'neuroshell/internal/version.GitCommit=#{git_commit}'",
      "-X 'neuroshell/internal/version.BuildDate=#{build_date}'"
    ].join(" ")
    
    # Build the main neuro binary
    system "go", "build", "-ldflags=#{ldflags}", "-o", "neuro", "./cmd/neuro"
    
    # Install the binary
    bin.install "neuro" => "neuroshell"
  end

  test do
    # Test that the binary works
    assert_match "NeuroShell", shell_output("#{bin}/neuroshell --version")
  end
end