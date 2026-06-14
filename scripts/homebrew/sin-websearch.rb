# Purpose: Homebrew formula for sin-websearch release binaries.
# Docs: scripts/homebrew/sin-websearch.rb.doc.md
class SinWebsearch < Formula
  desc "Unified Intelligence Gateway for OpenSIN"
  homepage "https://github.com/OpenSIN-Code/web_search_bundle"
  version "0.4.2"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.4.2/sin-websearch-darwin-amd64"
      sha256 "c00495c9b118e02a3c3ad7f299edc6de66d6bd8ff6d4eb1c6253898f66fce341"
    end
    on_arm do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.4.2/sin-websearch-darwin-arm64"
      sha256 "3ce87023ae63e8a484156ef838772178dac74e517d30e26057320b1fc31a497a"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.4.2/sin-websearch-linux-amd64"
      sha256 "3fcf113d4dbb97d01444662a496b67efd216810eef5b0e8506957316fbd11eca"
    end
    on_arm do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.4.2/sin-websearch-linux-arm64"
      sha256 "1b356e8f6034e974948d79cb2aef19e2a750167db5a3dce507fb5b999d90832e"
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
