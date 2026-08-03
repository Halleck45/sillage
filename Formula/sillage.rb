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
      sha256 "14f8112f163658fe9e44765ad23a035c0832a310a417fde91473f556333bc02f"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.5.0/sillage_darwin_amd64"
      sha256 "4623b387a37d5782592f6abc86487dc11988f9d1107ef0cd6a16d97fbc5502d6"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.5.0/sillage_linux_arm64"
      sha256 "39d34ea0ed03032b74406105f584a742e44db230cd2528d0d57cb081af7afb4b"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.5.0/sillage_linux_amd64"
      sha256 "ba51b85ea73077f88d45c1823051a019e4f16a4dd3ce5785082e5aa279701ddf"
    end
  end

  def install
    bin.install Dir["sillage_*"].first => "sillage"
  end

  test do
    assert_match "listen address", shell_output("#{bin}/sillage --help 2>&1")
  end
end
