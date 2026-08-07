# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.14.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.14.0/sillage_darwin_arm64"
      sha256 "12e1a083e884ba04130e38cf95cb992383c1b783f3eee286bd5b9fca1dba1043"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.14.0/sillage_darwin_amd64"
      sha256 "183d444c26fd7a513e2f30b880c292a4c4b74861471a3f0ecb0c8ac308939d46"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.14.0/sillage_linux_arm64"
      sha256 "1727d7564bbccf533d522d4560eba3128d0d8529774d23ba3ec2d658cbf5b007"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.14.0/sillage_linux_amd64"
      sha256 "32c1fe8d200362af7e6ef4e82f678494bbae5b2bebe728b7f4928fa4f2dae50c"
    end
  end

  def install
    bin.install Dir["sillage_*"].first => "sillage"
  end

  # `brew services start sillage` : lancement à l'ouverture de session
  # (launchd sous macOS, systemd --user sous Linux).
  service do
    run [opt_bin/"sillage"]
    keep_alive true
    log_path var/"log/sillage.log"
    error_log_path var/"log/sillage.log"
    # Sillage lance des CLI (claude, codex, copilot, agy, kiro-cli, git, gh, glab). Sans PATH explicite,
    # un service ne voit qu'un PATH minimal et tous les agents seraient signalés
    # « binaire absent du PATH ». On ajoute ~/.local/bin, où atterrissent la
    # plupart des installeurs de CLI d'agents.
    environment_variables PATH: "#{std_service_path_env}:#{Dir.home}/.local/bin"
  end

  test do
    assert_match "listen address", shell_output("#{bin}/sillage --help 2>&1")
  end
end
