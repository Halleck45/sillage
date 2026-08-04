# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.8.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.8.0/sillage_darwin_arm64"
      sha256 "c6c38174434d6aa8ae359276cfa553993afc4ef1a8989e950a05c018238e2065"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.8.0/sillage_darwin_amd64"
      sha256 "483750a8253a96af55b21337afd77cd1a96fcf9d65b56389331d2d630a3b8fa1"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.8.0/sillage_linux_arm64"
      sha256 "4cc4726c933870465c5f899bd21c96aa8d6459dbf5ac1f16e132e483c4876dd0"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.8.0/sillage_linux_amd64"
      sha256 "fcc396f9f61b939cc1230bd2c3137d10417be1ef7e7c8582cf06f11b68f50de4"
    end
  end

  def install
    bin.install Dir["sillage_*"].first => "sillage"
  end

  test do
    assert_match "listen address", shell_output("#{bin}/sillage --help 2>&1")
  end
end
