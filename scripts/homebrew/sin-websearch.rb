# Purpose: Homebrew formula for sin-websearch release binaries.
# Docs: scripts/homebrew/sin-websearch.rb.doc.md
class SinWebsearch < Formula
  desc "Unified Intelligence Gateway for OpenSIN"
  homepage "https://github.com/OpenSIN-Code/web_search_bundle"
  version "0.3.0"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.3.0/sin-websearch-darwin-amd64"
      sha256 "8f0cd8470ab7d047519d79e481d380f69769c3f35719e84e528efeb1ec9f56ee"
    end
    on_arm do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.3.0/sin-websearch-darwin-arm64"
      sha256 "6d3e528e80c49d84fbe9575f80766f325fbd61e65662d251428a42e66b90c476"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.3.0/sin-websearch-linux-amd64"
      sha256 "6fa60e479b39b5e44bd9eaff7d6c2f229c57c171d4cdb8e0477a9e344f914a45"
    end
    on_arm do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.3.0/sin-websearch-linux-arm64"
      sha256 "59a97ea16eda3c0f885a63cfc0f4e3d04440e195383ba6fd6e87e8fc523c9415"
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
