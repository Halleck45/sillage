// Package server implémente le backend HTTP/JSON/SSE de l'application Sillage.
package server

import "time"

// Tokens représente une consommation de tokens LLM et son coût estimé.
type Tokens struct {
	Input   int     `json:"input"`
	Output  int     `json:"output"`
	CostUsd float64 `json:"costUsd"`
}

// Repo est un dépôt git parmi ceux d'un projet. Name est court et affichable,
// unique dans le projet ; Path est un chemin absolu vers un dépôt git valide.
type Repo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Link est une URL épinglée sur un projet (site, dépôt, dashboard...). Title
// est fourni par l'utilisateur ou récupéré best-effort depuis la page
// (voir links.go), avec le nom d'hôte en repli.
type Link struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// Project est suivi par Sillage et peut regrouper plusieurs dépôts git (Repos).
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"` // une phrase, affichée sous le nom
	Repos       []Repo `json:"repos"`
	Links       []Link `json:"links"` // liens épinglés (max 12, http(s) uniquement)
	Unread      int    `json:"unread"`
	Tokens      Tokens `json:"tokens"`

	// CheckCmd est la commande de vérification lancée après chaque tâche
	// (ex : "go build ./..."). Exposée et modifiable via PATCH /api/projects/{id}.
	CheckCmd string `json:"checkCmd"`

	// ContextPrompt est un texte libre transmis aux agents (voir runner.go :
	// ajouté au system prompt claude, préfixe du prompt codex, ignoré par fake).
	ContextPrompt string `json:"contextPrompt"`
}

// Card est un chantier (vocabulaire produit) : regroupe des tâches dans une
// colonne du kanban d'un projet. Le nom technique Card/cards ne change pas.
type Card struct {
	ID            string  `json:"id"`
	ProjectID     string  `json:"projectId"`
	Column        string  `json:"column"`
	Title         string  `json:"title"`
	TasksTotal    int     `json:"tasksTotal"`
	TasksDone     int     `json:"tasksDone"`
	DocsCount     int     `json:"docsCount"`
	MessagesCount int     `json:"messagesCount"`
	ReviewCount   int     `json:"reviewCount"`
	Progress      int     `json:"progress"`
	LiveActivity  *string `json:"liveActivity"`

	// ContextPrompt est un texte libre transmis aux agents (voir runner.go :
	// ajouté au system prompt claude, préfixe du prompt codex, ignoré par fake).
	ContextPrompt string `json:"contextPrompt"`
}

// Agent est un profil d'agent IA (claude, codex ou fake).
type Agent struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Emoji         string `json:"emoji"`
	Color         string `json:"color"`
	Model         string `json:"model"`
	Cli           string `json:"cli"`
	ContextPrompt string `json:"contextPrompt"`
	Active        bool   `json:"active"`
}

// AgentOut est la représentation d'un Agent exposée par l'API : les mêmes
// champs qu'Agent (via embedding), plus Warning. Warning est calculé à la
// volée à chaque ListAgents (jamais persisté dans state.json : Agent lui-même
// n'a pas ce champ, donc rien à vider à l'écriture disque).
type AgentOut struct {
	Agent
	Warning string `json:"warning"`
}

// Check est le résultat d'une vérification projet (ex : go test) pour une tâche.
type Check struct {
	Label string `json:"label"`
	Ok    bool   `json:"ok"`
}

// Task est une tâche assignée à un agent, exécutée dans un worktree git dédié.
type Task struct {
	ID            string    `json:"id"`
	CardID        string    `json:"cardId"`
	ProjectID     string    `json:"projectId"`
	Ref           int       `json:"ref"`
	Title         string    `json:"title"`
	AgentID       string    `json:"agentId"`
	RepoName      string    `json:"repoName"` // dépôt du projet utilisé pour le worktree
	Branch        string    `json:"branch"`
	Status        string    `json:"status"` // running|review|shipped|done|cancelled
	MessagesCount int       `json:"messagesCount"`
	FilesCount    int       `json:"filesCount"`
	DocsCount     int       `json:"docsCount"`
	Checks        []Check   `json:"checks"`
	LiveActivity  *string   `json:"liveActivity"`
	Unread        bool      `json:"unread"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Tokens        Tokens    `json:"tokens"`

	// Champs internes ; persistés dans state.json (sinon perdus au redémarrage)
	// et visibles par le client authentifié, ce qui est sans enjeu en mono-utilisateur.
	Base        string `json:"base"`        // branche de base au moment de la création
	WorktreeDir string `json:"worktreeDir"` //
	SessionID   string `json:"sessionId"`   // session claude, pour --resume
}

// Message est un message échangé dans le fil d'une tâche. AuthorName porte
// le nom de l'agent pour author="agent" ; pour author="user", c'est le
// displayName des Settings (vide si non renseigné, le frontend affiche alors
// "Vous"/"You").
type Message struct {
	ID         string    `json:"id"`
	TaskID     string    `json:"taskId"`
	Author     string    `json:"author"` // user|agent
	AuthorName string    `json:"authorName"`
	Text       string    `json:"text"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Workspace est l'état de synchronisation git persisté de l'espace de
// données (dataDir), stocké dans state.json.
type Workspace struct {
	SetupDone  bool       `json:"setupDone"`
	SyncRemote string     `json:"syncRemote"`
	LastSyncAt *time.Time `json:"lastSyncAt"`
}

// WorkspaceStatus est la réponse de GET /api/workspace, le contenu de
// l'événement SSE "workspace" et le champ State.Workspace : elle combine
// l'état persisté (Workspace) et des faits git calculés à la volée.
type WorkspaceStatus struct {
	SetupDone    bool       `json:"setupDone"`
	GitEnabled   bool       `json:"gitEnabled"`
	Remote       string     `json:"remote"`
	Dirty        bool       `json:"dirty"`
	LastCommitAt *time.Time `json:"lastCommitAt"`
	LastSyncAt   *time.Time `json:"lastSyncAt"`
}

// SyncResponse est la réponse de POST /api/workspace/sync.
type SyncResponse struct {
	Output     string    `json:"output"`
	LastSyncAt time.Time `json:"lastSyncAt"`
}

// Settings sont les préférences globales persistées de l'utilisateur.
type Settings struct {
	DisplayName string `json:"displayName"`
	Lang        string `json:"lang"` // ""|"fr"|"en"
}

// State est la réponse de GET /api/state.
type State struct {
	Projects  []Project       `json:"projects"`
	Cards     []Card          `json:"cards"`
	Tasks     []Task          `json:"tasks"`
	Agents    []AgentOut      `json:"agents"`
	Workspace WorkspaceStatus `json:"workspace"`
	Settings  Settings        `json:"settings"`
	Tokens    struct {
		Global Tokens `json:"global"`
	} `json:"tokens"`
}

// TokensEvent est le contenu de l'événement SSE "tokens".
type TokensEvent struct {
	Global   Tokens            `json:"global"`
	Projects map[string]Tokens `json:"projects"`
	Tasks    map[string]Tokens `json:"tasks"`
}

// DiffLine est une ligne d'un hunk de diff unifié.
type DiffLine struct {
	Type string `json:"type"` // ctx|add|del
	Text string `json:"text"`
}

// Hunk est un bloc de diff unifié (@@ ... @@).
type Hunk struct {
	Header string     `json:"header"`
	Lines  []DiffLine `json:"lines"`
}

// DiffFile est le diff d'un fichier.
type DiffFile struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Hunks     []Hunk `json:"hunks"`
}

// DiffResponse est la réponse de GET /api/tasks/{id}/diff.
type DiffResponse struct {
	Branch string     `json:"branch"`
	Base   string     `json:"base"`
	Files  []DiffFile `json:"files"`
}

// CommitInfo est une ligne de `git log`.
type CommitInfo struct {
	Hash    string
	Subject string
	RelTime string
}

// Item est un livrable (code, doc ou image) affiché dans le détail de tâche.
type Item struct {
	Kind  string `json:"kind"`
	Title string `json:"title"`
	Meta  string `json:"meta"`
	Path  string `json:"path,omitempty"`
}

// DeliverablesResponse est la réponse de GET /api/tasks/{id}/deliverables.
type DeliverablesResponse struct {
	Code   []Item `json:"code"`
	Docs   []Item `json:"docs"`
	Images []Item `json:"images"`
}

// ShipResponse est la réponse de POST /api/tasks/{id}/ship. BranchUrl pointe
// vers la branche sur GitHub (si le remote origin est un dépôt github.com),
// vide sinon : jamais d'erreur pour cette information optionnelle.
type ShipResponse struct {
	Task      Task   `json:"task"`
	Output    string `json:"output"`
	BranchUrl string `json:"branchUrl"`
}

// PRResponse est la réponse de POST /api/tasks/{id}/pr.
type PRResponse struct {
	URL string `json:"url"`
}
