# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.9.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.9.0/sillage_darwin_arm64"
      sha256 "bc677a7adf40f6b073f20f27d9331bc1c40bf7795a93fc71dad234d886f98504"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.9.0/sillage_darwin_amd64"
      sha256 "664747339514b0f614bb2120549472686a14c0c20eb400847b7284910732c077"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.9.0/sillage_linux_arm64"
      sha256 "b24d3f666d0599ea9274b3478c8c1c9a9b7cbe6eb305485d16146ea0969f1425"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.9.0/sillage_linux_amd64"
      sha256 "ff745d6d80db03b37f9ec646f8ff40ae5fa61eb02ba5775c3f1776c2defcfa51"
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
    # Sillage lance des CLI (claude, codex, git, gh, glab). Sans PATH explicite,
    # un service ne voit qu'un PATH minimal et tous les agents seraient signalés
    # « binaire absent du PATH ». On ajoute ~/.local/bin, où atterrissent la
    # plupart des installeurs de CLI d'agents.
    environment_variables PATH: "#{std_service_path_env}:#{Dir.home}/.local/bin"
  end

  test do
    assert_match "listen address", shell_output("#{bin}/sillage --help 2>&1")
  end
end
