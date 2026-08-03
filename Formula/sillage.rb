# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.5.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.5.0/sillage_darwin_arm64"
      sha256 "bc1d08ec1e5ee7b51bc5887e794680dfa9a433f91709b5b6d80090b822dd7d79"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.5.0/sillage_darwin_amd64"
      sha256 "bb66ac3716a7e3e17e12352e31d2173d3ffc1b343b9d9d4b72e6c2592a2896d6"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.5.0/sillage_linux_arm64"
      sha256 "a551fed9198aec7e5070c6c502bacd59894e1e608692e21e5772619fe8ece4bb"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.5.0/sillage_linux_amd64"
      sha256 "c9b65fe5a8ee781e245b20f6a74b74dc3b6360ec0a51da3a2ca5e46ef3da09cb"
    end
  end

  def install
    bin.install Dir["sillage_*"].first => "sillage"
  end

  test do
    assert_match "listen address", shell_output("#{bin}/sillage --help 2>&1")
  end
end
