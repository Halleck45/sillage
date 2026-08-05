# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.10.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.10.0/sillage_darwin_arm64"
      sha256 "5b11513218fcc98a0bf331e07a5b1a31c50047f63fd43e8f1502a11c74af0c6d"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.10.0/sillage_darwin_amd64"
      sha256 "361b20fc4881719145f3bb1a5cb87d424aea626d178cddf5d0fb8cd0d2c47458"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.10.0/sillage_linux_arm64"
      sha256 "edae644630bb62ac681b093590c76cc4aa17f088fbf87a2fbdf1ea1155c77bea"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.10.0/sillage_linux_amd64"
      sha256 "4df83bafb0fd4b372bad8387719e952f905f006d8b4387970081390757c8244e"
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
