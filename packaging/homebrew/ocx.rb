class Ocx < Formula
  desc "Headless Claude/Gemini-style wrapper for opencode (GPT-5 ready)"
  homepage "https://github.com/YOURUSER/opencode-gpt5-fork"
  version "0.1.0"
  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/YOURUSER/opencode-gpt5-fork/releases/download/v0.1.0/ocx_darwin_arm64.tar.gz"
      sha256 "REPLACE_ME"
    else
      url "https://github.com/YOURUSER/opencode-gpt5-fork/releases/download/v0.1.0/ocx_darwin_amd64.tar.gz"
      sha256 "REPLACE_ME"
    end
  end
  def install
    bin.install "ocx"
  end
end
