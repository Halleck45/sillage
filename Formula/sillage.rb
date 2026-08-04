# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.6.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.6.0/sillage_darwin_arm64"
      sha256 "d0c0f8915b0823870ce07139a58f9ddaade21e25623b6c83433f85146cf6ff2a"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.6.0/sillage_darwin_amd64"
      sha256 "741aea2d29610364a1828f32410fcadbfc0ecb58b5a790f136199f6ffe487637"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.6.0/sillage_linux_arm64"
      sha256 "ad64e8ea5e4f287f4398ce1fc1abb46516759cbeb1eec03d216a37e3f9bdecf0"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.6.0/sillage_linux_amd64"
      sha256 "1a32d5f65411013669b649e1c55f2a63de0028e467b0aaf34e8400c56afca2e0"
    end
  end

  def install
    bin.install Dir["sillage_*"].first => "sillage"
  end

  test do
    assert_match "listen address", shell_output("#{bin}/sillage --help 2>&1")
  end
end
