# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.15.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.15.0/sillage_darwin_arm64"
      sha256 "f836540e358b0d600e1438568104756bea82c5f5ed5fe62ea4a4ea0bd9664ada"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.15.0/sillage_darwin_amd64"
      sha256 "e8d3c5583f1a98dae6ce4ed8a0f021988ec33bedfc3fa456e22e0c04dba771ed"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.15.0/sillage_linux_arm64"
      sha256 "ca0385b40ccf490315482ff0591e8edd85292d307e3c74d52bab31c94d1ec4cd"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.15.0/sillage_linux_amd64"
      sha256 "da9d48ff59a73a4f63e991df031c7aaea71ac630fa66b12fcaaf912da27029ec"
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
