// Package server implémente le backend HTTP/JSON/SSE de l'application Atelier.
package server

import "time"

// Tokens représente une consommation de tokens LLM et son coût estimé.
type Tokens struct {
	Input   int     `json:"input"`
	Output  int     `json:"output"`
	CostUsd float64 `json:"costUsd"`
}

// Project est un dépôt git suivi par Atelier.
type Project struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Unread int    `json:"unread"`
	Tokens Tokens `json:"tokens"`

	// CheckCmd est la commande de vérification lancée après chaque tâche
	// (ex : "go build ./..."). Non exposée par l'API v1.
	CheckCmd string `json:"-"`
}

// Card est une colonne du kanban d'un projet.
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
	Branch        string    `json:"branch"`
	Status        string    `json:"status"` // running|review|ready|shipped
	MessagesCount int       `json:"messagesCount"`
	FilesCount    int       `json:"filesCount"`
	DocsCount     int       `json:"docsCount"`
	Checks        []Check   `json:"checks"`
	LiveActivity  *string   `json:"liveActivity"`
	Unread        bool      `json:"unread"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Tokens        Tokens    `json:"tokens"`

	// Champs internes, non exposés par l'API.
	Base        string `json:"-"` // branche de base au moment de la création
	WorktreeDir string `json:"-"`
	SessionID   string `json:"-"` // session claude, pour --resume
}

// Message est un message échangé dans le fil d'une tâche.
type Message struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"taskId"`
	Author    string    `json:"author"` // user|agent
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

// State est la réponse de GET /api/state.
type State struct {
	Projects []Project `json:"projects"`
	Cards    []Card    `json:"cards"`
	Tasks    []Task    `json:"tasks"`
	Agents   []Agent   `json:"agents"`
	Tokens   struct {
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

// ShipResponse est la réponse de POST /api/tasks/{id}/ship.
type ShipResponse struct {
	Task   Task   `json:"task"`
	Output string `json:"output"`
}
