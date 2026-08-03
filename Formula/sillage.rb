# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.4.2"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.4.2/sillage_darwin_arm64"
      sha256 "d7300880ff216600ed50de18e2d0d519d17df0ba4eb76d578b714f46352e73e2"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.4.2/sillage_darwin_amd64"
      sha256 "9860a9f46bf28ef027bc761358947113f89b49061326cb7df45c7cfbae9cb21f"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.4.2/sillage_linux_arm64"
      sha256 "d522d80ef2607233fd1b21764712172bf1f09a1bfe0629483379ea3a01212d00"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.4.2/sillage_linux_amd64"
      sha256 "c6622157605904dfb5c82cb0831c8433b82c1bb454e1a34890b1bb5afb3f8a21"
    end
  end

  def install
    bin.install Dir["sillage_*"].first => "sillage"
  end

  test do
    assert_match "listen address", shell_output("#{bin}/sillage --help 2>&1")
  end
end
