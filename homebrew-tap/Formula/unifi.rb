class Unifi < Formula
  desc "UniFi Controller CLI tool for local network management"
  homepage "https://github.com/dl-alexandre/Local-UniFi-CLI"
  version "1.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/dl-alexandre/Local-UniFi-CLI/releases/download/v#{version}/unifi_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_ARM64"
    else
      url "https://github.com/dl-alexandre/Local-UniFi-CLI/releases/download/v#{version}/unifi_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/dl-alexandre/Local-UniFi-CLI/releases/download/v#{version}/unifi_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    else
      url "https://github.com/dl-alexandre/Local-UniFi-CLI/releases/download/v#{version}/unifi_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end
  end

  def install
    bin.install "unifi"
    
    # Install shell completions
    bash_completion.install "completions/unifi.bash" => "unifi" if File.exist?("completions/unifi.bash")
    zsh_completion.install "completions/_unifi" => "_unifi" if File.exist?("completions/_unifi")
    fish_completion.install "completions/unifi.fish" if File.exist?("completions/unifi.fish")
  end

  test do
    system "#{bin}/unifi", "--version"
  end
end
