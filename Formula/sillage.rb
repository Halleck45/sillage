# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.7.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.7.0/sillage_darwin_arm64"
      sha256 "905cbe3eab5303055301284c91d62e4ebac14fb34b73b8fe7190cb9d965eae6c"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.7.0/sillage_darwin_amd64"
      sha256 "c8ca5bc4cc93d51a1def7a661b741a725eaaba07f604f2125e43e5da031e8bf4"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.7.0/sillage_linux_arm64"
      sha256 "6f019fe0382ec18b48c5030b4f8b8c0e450c608d8e14db015ece34cd00e736c9"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.7.0/sillage_linux_amd64"
      sha256 "2e32bdc0d617ffd6c5ac8b48fe06bbbc6e043c677f5f55b358c209a42c4131e0"
    end
  end

  def install
    bin.install Dir["sillage_*"].first => "sillage"
  end

  test do
    assert_match "listen address", shell_output("#{bin}/sillage --help 2>&1")
  end
end
