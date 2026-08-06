# Formule Homebrew de Sillage (binaires précompilés).
# Mise à jour automatiquement par le workflow de release à chaque tag.
class Sillage < Formula
  desc "Calm web dashboard to pilot AI coding agents across your projects"
  homepage "https://github.com/Halleck45/sillage"
  version "0.12.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.12.0/sillage_darwin_arm64"
      sha256 "6ad1370300b7db08bc8a49920588a5c0e62576657770a2a034bb7aa22f23daf5"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.12.0/sillage_darwin_amd64"
      sha256 "e561f1823d2fc1a1c0002cb66c120c67e966d11ee7b13d286c3476f5aa8d87a1"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Halleck45/sillage/releases/download/v0.12.0/sillage_linux_arm64"
      sha256 "160473cfc9d4406bdb786e02a3181a909d22e2b1417ddeef4e57e3aecdee22f8"
    end
    on_intel do
      url "https://github.com/Halleck45/sillage/releases/download/v0.12.0/sillage_linux_amd64"
      sha256 "0af2bde29455266cfe584633ff2920e05a1ffdea9e86fcae5e269e9305994af2"
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
