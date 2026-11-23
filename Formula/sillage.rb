# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.17.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.17.0/sillage_darwin_arm64"
      sha256 "9e6eba0b5a6e34f7b0718fb97a569ae165800ce7b38f28bdebb10976386bad1d"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.17.0/sillage_darwin_amd64"
      sha256 "49a399af80e9dd19196b96e4d70a484c8b8ba89f6aeeefe4748d698cfc4db4b9"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.17.0/sillage_linux_arm64"
      sha256 "0064c76be94c3e105eb2832bc918ed8bee01d33e17fd9813cddbd7c34c5c9b79"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.17.0/sillage_linux_amd64"
      sha256 "0b0138e1fde689ddd731c5352ff53967edede8c3a310125ac5475665abe9f9ab"
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
