# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.4.1"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.4.1/sillage_darwin_arm64"
      sha256 "c8fabc4df48891456e9a0bfc5d8e8318dec6f59f97396936a0a05dab64164808"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.4.1/sillage_darwin_amd64"
      sha256 "f847fe2a3c4f8d188b90284864ef9f3673e411ab87ec3934e32977d014dd720e"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.4.1/sillage_linux_arm64"
      sha256 "bbe32298038f3ec4e855880d7ae3be26864e2ac0fa2d91f066d8353028500f27"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.4.1/sillage_linux_amd64"
      sha256 "21b26e3414e50856acfa13cdcaf4263bf998a772f6fd2d6664023cad4d5ebf82"
    end
  end

  def install
    bin.install Dir["sillage_*"].first => "sillage"
  end

  test do
    assert_match "listen address", shell_output("#{bin}/sillage --help 2>&1")
  end
end
