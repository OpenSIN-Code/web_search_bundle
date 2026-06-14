# Purpose: Homebrew formula for sin-websearch release binaries.
# Docs: scripts/homebrew/sin-websearch.rb.doc.md
class SinWebsearch < Formula
  desc "Unified Intelligence Gateway for OpenSIN"
  homepage "https://github.com/OpenSIN-Code/web_search_bundle"
  version "0.4.0"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.4.0/sin-websearch-darwin-amd64"
      sha256 "b32f5de61bc4d2dfe5a69e4a9084a9e8309f40aadd313382b2f226eee97e4304"
    end
    on_arm do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.4.0/sin-websearch-darwin-arm64"
      sha256 "ca7295ef9fe3b43f40a98b4d70e64f390acffd4d18d23ed789abcf2725974e94"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.4.0/sin-websearch-linux-amd64"
      sha256 "f5c48b2dc1c028c42ebaac04612fd777e62f4f01628bedce18cffe29dd3cea0b"
    end
    on_arm do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.4.0/sin-websearch-linux-arm64"
      sha256 "d6eff912eee61ba4c0d58ffe22e30108055cacace42fac1b6212b52760f09ca5"
    end
  end

  def install
    bin.install Dir["sin-websearch-*"].first => "sin-websearch"
    chmod 0755, bin/"sin-websearch"
  end

  test do
    system "#{bin}/sin-websearch", "--version"
  end
end
