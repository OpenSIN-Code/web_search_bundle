# Purpose: Homebrew formula for sin-websearch release binaries.
# Docs: scripts/homebrew/sin-websearch.rb.doc.md
class SinWebsearch < Formula
  desc "Unified Intelligence Gateway for OpenSIN"
  homepage "https://github.com/OpenSIN-Code/web_search_bundle"
  version "0.4.1"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.4.1/sin-websearch-darwin-amd64"
      sha256 "1cf022af56fb73614a8dfcb9a856b737ba8cdbe5e988f303e26a436af08886fd"
    end
    on_arm do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.4.1/sin-websearch-darwin-arm64"
      sha256 "3ea56bc85d6ff9ba3140f86f130bde277f4ef0f6f9b0eee7a5734d2eaf637a9d"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.4.1/sin-websearch-linux-amd64"
      sha256 "36e7516ffd8d440fd3ef03a926619b60e1dc9e79be9379efe5937cc51843822d"
    end
    on_arm do
      url "https://github.com/OpenSIN-Code/web_search_bundle/releases/download/v0.4.1/sin-websearch-linux-arm64"
      sha256 "0639999d73214c9eca0dbf79b89337caadf0f0f79826c1a375b16978b8c51e46"
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
