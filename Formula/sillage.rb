# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.4.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.4.0/sillage_darwin_arm64"
      sha256 "c61f3386bea7ad8dea698d87061abd615a9eba25751db40239852aa49fa0ade4"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.4.0/sillage_darwin_amd64"
      sha256 "c17a4d804de7158eb237cb71e06f5e10524a241b136dbaa503d49a8632641ba4"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.4.0/sillage_linux_arm64"
      sha256 "377cf3cd671e0766ddbbf2032002c750cd3f3462982928341918a629d04bcb0e"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.4.0/sillage_linux_amd64"
      sha256 "de762998035fd732ae248a404de6e537dbef1eec3a1d054c52551fcd6ef887b7"
    end
  end

  def install
    bin.install Dir["sillage_*"].first => "sillage"
  end

  test do
    assert_match "listen address", shell_output("#{bin}/sillage --help 2>&1")
  end
end
