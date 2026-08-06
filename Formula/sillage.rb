# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.13.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.13.0/sillage_darwin_arm64"
      sha256 "2b883d4a2ecfda1f5ce95864e34aeb2162569a77074bf8bd34c3bd6ecba36402"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.13.0/sillage_darwin_amd64"
      sha256 "a9b3debd92527b1613977dd5d8019e343383e19dd1d4c67c46bb41e558d5ada8"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.13.0/sillage_linux_arm64"
      sha256 "89bf3f6266d8289a3619b0a7e2f9ef1c5be8788b20d272a37639ac2b0823d379"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.13.0/sillage_linux_amd64"
      sha256 "eaabf6e56f739258137bfe1661a4c6942ef0fd95e926e231debd2fb8b696845c"
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
