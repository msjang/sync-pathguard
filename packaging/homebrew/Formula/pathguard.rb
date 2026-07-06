# Homebrew Formula for the pathguard CLI (installs the prebuilt binary).
# Copy into a tap repo (e.g. msjang/homebrew-tap → Formula/pathguard.rb) and
# fill in `version` + the four `sha256` values from the GitHub release assets.
class Pathguard < Formula
  desc "CLI that flags NFD filename/path byte-lengths risking cloud/NAS sync"
  homepage "https://github.com/msjang/sync-pathguard"
  version "0.1.0"

  on_macos do
    on_arm do
      url "https://github.com/msjang/sync-pathguard/releases/download/v#{version}/pathguard-macos-arm64.zip"
      sha256 "REPLACE_WITH_macos_arm64_sha256"
    end
    on_intel do
      url "https://github.com/msjang/sync-pathguard/releases/download/v#{version}/pathguard-macos-amd64.zip"
      sha256 "REPLACE_WITH_macos_amd64_sha256"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/msjang/sync-pathguard/releases/download/v#{version}/pathguard-linux-arm64.tar.gz"
      sha256 "REPLACE_WITH_linux_arm64_sha256"
    end
    on_intel do
      url "https://github.com/msjang/sync-pathguard/releases/download/v#{version}/pathguard-linux-amd64.tar.gz"
      sha256 "REPLACE_WITH_linux_amd64_sha256"
    end
  end

  def install
    bin.install "pathguard"
  end

  test do
    system bin/"pathguard", "--root", testpath, "--json"
  end
end
