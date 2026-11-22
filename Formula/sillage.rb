# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.16.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.16.0/sillage_darwin_arm64"
      sha256 "b8740955390c59eb63ce07d9840d55172bcbb133a7e116bee93295ab76a244e1"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.16.0/sillage_darwin_amd64"
      sha256 "00352e8caba6ddf79aad22877fdeb42acc985a408388505b63ec8a40fce4e4ea"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.16.0/sillage_linux_arm64"
      sha256 "f3de4bba95818fe19c531af9afa62fc49ff33bd48a5492d3f29b6c879f474db9"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.16.0/sillage_linux_amd64"
      sha256 "47b3431a59005edf298696ac74455ada48b262c1a6f52e356d9340ac54713330"
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
