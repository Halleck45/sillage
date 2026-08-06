# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.11.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.11.0/sillage_darwin_arm64"
      sha256 "9770471d0d21cc244ab87eb6262622ae34ba82c947f1266289dfa6e7ee6b4f6d"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.11.0/sillage_darwin_amd64"
      sha256 "f66ac5e392c1eae7939eba2deb97f6f6f32f14a64cdf4ef32c8f6da2b1e6079b"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.11.0/sillage_linux_arm64"
      sha256 "a4e49dc2d9abc3ee374940046daa8ac6d1668d5e4f92b6e6b5a181e9adcf3b2a"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.11.0/sillage_linux_amd64"
      sha256 "b77934a521bbaf131ab01d9378922d93f7b6dba9ca4343c391d6b5215be72fee"
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
