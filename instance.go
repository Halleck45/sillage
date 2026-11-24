// Une seule instance par dossier de données. Deux serveurs sur le même
// state.json s'écrasent mutuellement à chaque sauvegarde et se disputent le
// dépôt git de l'espace de travail ; c'est donc le dossier, pas le port, qui
// identifie une instance.
//
// Le verrou est un flock sur <dataDir>/sillage.lock, tenu toute la vie du
// process. Le noyau le libère à la mort du process : un crash ne laisse pas de
// verrou périmé. Le fichier est ouvert en O_CLOEXEC (défaut de Go) : ni les
// agents ni les recettes ne l'héritent, et restartInPlace le relâche à l'exec
// pour que le nouveau binaire le reprenne aussitôt.
//
// Le fichier porte aussi l'adresse de l'instance, écrite une fois le port lié :
// c'est ainsi que `sillage -app` retrouve un serveur lancé sur un autre port
// que celui de -addr, au lieu d'en démarrer un second.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const instanceLockName = "sillage.lock"

// instanceInfo : ce que l'instance qui tient le verrou publie d'elle-même.
type instanceInfo struct {
	PID     int    `json:"pid"`
	Addr    string `json:"addr"`
	Version string `json:"version"`
}

// errInstanceRunning : une autre instance tient déjà le dossier de données.
var errInstanceRunning = errors.New("another Sillage instance holds this data directory")

// acquireInstanceLock prend le verrou du dossier de données. Le fichier rendu
// doit rester ouvert jusqu'à l'arrêt : le fermer relâche le verrou.
func acquireInstanceLock(dataDir string) (*os.File, error) {
	f, err := os.OpenFile(filepath.Join(dataDir, instanceLockName), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, errInstanceRunning
		}
		return nil, fmt.Errorf("cannot lock %s: %w", instanceLockName, err)
	}
	return f, nil
}

// publishInstance écrit l'adresse de l'instance dans le fichier verrouillé.
func publishInstance(lock *os.File, info instanceInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	if err := lock.Truncate(0); err != nil {
		return err
	}
	_, err = lock.WriteAt(append(data, '\n'), 0)
	return err
}

// readInstance lit ce qu'a publié l'instance qui tient le verrou. Elle prend le
// verrou avant de charger l'état et n'écrit son adresse qu'une fois le port
// lié : entre les deux, le fichier est vide. On attend donc jusqu'à `wait`.
func readInstance(dataDir string, wait time.Duration) (instanceInfo, error) {
	path := filepath.Join(dataDir, instanceLockName)
	deadline := time.Now().Add(wait)
	for {
		var info instanceInfo
		f, err := os.Open(path)
		if err == nil {
			data, _ := io.ReadAll(io.LimitReader(f, 4096))
			f.Close()
			if json.Unmarshal(data, &info) == nil && info.Addr != "" {
				return info, nil
			}
		}
		if time.Now().After(deadline) {
			return info, fmt.Errorf("the running instance has not published its address in %s", path)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
