package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Pièces jointes du composeur : des images collées (Ctrl+V) ou choisies avec
// le bouton joindre. Elles arrivent en base64 dans le corps JSON du message
// (le multipart est exclu : toute mutation exige Content-Type
// application/json, c'est la protection CSRF), puis sont écrites dans
// <dataDir>/attachments/<taskId>/. state.json n'en garde que le descripteur :
// y stocker des mégaoctets de base64 gonflerait chaque sauvegarde et chaque
// commit de l'espace de travail.
//
// Ce qui part à l'agent est le chemin absolu du fichier, une seule fois, dans
// le texte transmis au CLI (voir withAttachmentPaths) : aucun des CLI
// supportés ne prend d'image sur son entrée standard, tous savent lire un
// fichier local.

const (
	// maxAttachmentBytes plafonne une image décodée. Une capture d'écran pèse
	// quelques centaines de kilooctets ; 8 Mo laisse de la marge sans ouvrir
	// la porte à un state.json d'espace de travail ingérable.
	maxAttachmentBytes = 8 << 20
	// maxAttachmentsPerMessage plafonne le nombre d'images d'un même message.
	maxAttachmentsPerMessage = 6
)

// attachmentTypes est la liste fermée des formats acceptés, avec l'extension
// posée sur le fichier. Le type déclaré par le navigateur n'est pas cru sur
// parole : ce qui n'est pas dans cette table est refusé.
var attachmentTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// attachmentInput est une image reçue dans le corps d'un POST de message.
type attachmentInput struct {
	Name string `json:"name"`
	Mime string `json:"mime"`
	Data string `json:"data"` // base64, sans préfixe data:
}

func attachmentsDir(dataDir, taskID string) string {
	return filepath.Join(dataDir, "attachments", taskID)
}

// removeTaskAttachments efface les images d'une tâche supprimée. Best-effort :
// une pièce jointe orpheline ne casse rien, elle occupe du disque.
func removeTaskAttachments(dataDir, taskID string) {
	if dataDir == "" || taskID == "" {
		return
	}
	os.RemoveAll(attachmentsDir(dataDir, taskID))
}

// saveAttachments écrit les images d'un message et retourne leurs
// descripteurs. Si l'une échoue, celles déjà écrites sont retirées : un
// message n'est jamais accepté avec une pièce jointe manquante.
func saveAttachments(dataDir, taskID string, inputs []attachmentInput) ([]Attachment, error) {
	if len(inputs) == 0 {
		return nil, nil
	}
	if len(inputs) > maxAttachmentsPerMessage {
		return nil, fmt.Errorf("too many attachments")
	}
	dir := attachmentsDir(dataDir, taskID)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("attachments directory: %w", err)
	}
	out := make([]Attachment, 0, len(inputs))
	for _, in := range inputs {
		att, err := saveAttachment(dir, in)
		if err != nil {
			for _, done := range out {
				os.Remove(done.Path)
			}
			return nil, err
		}
		out = append(out, att)
	}
	return out, nil
}

func saveAttachment(dir string, in attachmentInput) (Attachment, error) {
	mime := strings.ToLower(strings.TrimSpace(in.Mime))
	ext, ok := attachmentTypes[mime]
	if !ok {
		return Attachment{}, fmt.Errorf("unsupported image type")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(in.Data))
	if err != nil {
		return Attachment{}, fmt.Errorf("invalid image data")
	}
	if len(raw) == 0 {
		return Attachment{}, fmt.Errorf("empty image")
	}
	if len(raw) > maxAttachmentBytes {
		return Attachment{}, fmt.Errorf("image too large")
	}
	id, err := attachmentID()
	if err != nil {
		return Attachment{}, err
	}
	path := filepath.Join(dir, id+ext)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return Attachment{}, fmt.Errorf("write attachment: %w", err)
	}
	return Attachment{
		ID:   id,
		Name: attachmentName(in.Name, ext),
		Mime: mime,
		Size: int64(len(raw)),
		Path: path,
	}, nil
}

func attachmentID() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("attachment id: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// attachmentName nettoie le nom d'origine, qui n'est qu'une étiquette
// affichée : jamais un nom de fichier sur le disque (celui-ci est l'id).
func attachmentName(name, ext string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, name)
	if name == "" || name == "." || name == ".." || name == string(filepath.Separator) {
		return "image" + ext
	}
	if len(name) > 80 {
		name = name[:80]
	}
	return name
}

// handleGetAttachment sert l'image d'une pièce jointe au frontend
// (l'authentification est déjà faite par le middleware).
func (s *Server) handleGetAttachment(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	att, ok := s.store.FindAttachment(taskID, r.PathValue("attId"))
	if !ok {
		writeError(w, http.StatusNotFound, "attachment not found")
		return
	}
	// Le chemin vient du store, mais tout ce qui sort du répertoire des pièces
	// jointes est refusé : un state.json trafiqué ne doit pas faire de Sillage
	// un lecteur de fichiers arbitraires.
	dir := attachmentsDir(s.dataDir, taskID)
	path := filepath.Clean(att.Path)
	if !strings.HasPrefix(path, dir+string(filepath.Separator)) {
		writeError(w, http.StatusNotFound, "attachment not found")
		return
	}
	f, err := os.Open(path)
	if err != nil {
		writeError(w, http.StatusNotFound, "attachment not found")
		return
	}
	defer f.Close()
	modTime := time.Time{}
	if info, err := f.Stat(); err == nil {
		modTime = info.ModTime()
	}
	w.Header().Set("Content-Type", att.Mime)
	// Le contenu d'un id ne change jamais : le navigateur peut le garder, mais
	// la réponse reste privée (elle a demandé une session).
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, filepath.Base(path), modTime, f)
}

// withAttachmentPaths ajoute au texte transmis au CLI la liste des images
// jointes, par leur chemin absolu. Le message stocké, lui, garde le texte de
// l'utilisateur tel quel : ces lignes sont une instruction pour l'agent, pas
// un contenu de conversation.
func withAttachmentPaths(text string, atts []Attachment) string {
	if len(atts) == 0 {
		return text
	}
	var b strings.Builder
	b.WriteString(strings.TrimSpace(text))
	if b.Len() > 0 {
		b.WriteString("\n\n")
	}
	b.WriteString("Attached images (local files, read them):")
	for _, a := range atts {
		b.WriteString("\n- ")
		b.WriteString(a.Path)
	}
	return b.String()
}
