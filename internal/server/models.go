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

	// PreviewCmd est la commande de recette manuelle du dépôt, lancée par
	// Sillage dans le worktree d'un chantier ou d'une tâche (voir
	// docs/SPEC-RECETTE.md). Vide = pas de recette pour ce dépôt.
	PreviewCmd string `json:"previewCmd"`

	// PreviewURL est l'adresse à ouvrir quand la commande tourne (optionnelle).
	// Elle accepte les mêmes variables que la commande.
	PreviewURL string `json:"previewUrl"`
}

// Link est une URL épinglée sur un projet (site, dépôt, dashboard...). Title
// est fourni par l'utilisateur ou récupéré best-effort depuis la page
// (voir links.go), avec le nom d'hôte en repli.
type Link struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// Delivery définit ce que « livrer » veut dire pour un projet (voir
// docs/SPEC-LIVRAISON.md). Le fournisseur (github, gitlab) n'est jamais
// persisté : il est redéduit du remote origin à chaque opération.
type Delivery struct {
	// Mode est l'un des quatre comportements de livraison :
	//
	//	pr         pousse la branche du chantier, puis ouvre la pull/merge request
	//	push       pousse la branche du chantier, sans rien ouvrir
	//	merge      fusionne la branche du chantier dans Target, en local, sans jamais pousser
	//	merge-push fusionne dans Target puis pousse Target
	//
	// Les deux modes de fusion sont fast-forward uniquement (voir MergeLocal).
	Mode string `json:"mode"` // pr|push|merge|merge-push

	// Target est la branche de destination : base de la PR en mode "pr",
	// branche fusionnée en modes "merge" et "merge-push". Vide = branche par
	// défaut du dépôt (branche courante au moment de la création de la branche
	// de chantier).
	Target string `json:"target"`

	// StackedPrs réserve l'option « une PR par tâche, empilées » (lot 3, voir
	// docs/SPEC-LIVRAISON.md) : persistée mais ignorée pour l'instant.
	StackedPrs bool `json:"stackedPrs"`
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

	// AllowedTools accorde aux agents claude de ce projet des outils en plus du
	// socle du binaire (typiquement la chaîne du langage : "Bash(pytest:*)").
	// Saisi par l'humain dans les réglages, jamais lu depuis un fichier du dépôt
	// (invariant 5 de CONTRIBUTING.md) : ces fichiers sont écrits par les agents.
	// Le refus figé (claudeDeniedTools) l'emporte sur toute entrée d'ici.
	// Ignoré par codex (sandbox) et par fake.
	AllowedTools []string `json:"allowedTools"`

	// Delivery définit ce que livrer veut dire pour ce projet (voir Delivery).
	Delivery Delivery `json:"delivery"`
}

// ProjectOut est la représentation d'un Project exposée par l'API : les mêmes
// champs que Project (via embedding), plus DeliveryWarning. Ce dernier est
// calculé à la volée (binaire gh/glab absent, remote origin manquant) et
// jamais persisté dans state.json : Project lui-même n'a pas ce champ, donc
// rien à vider à l'écriture disque. Même pattern qu'AgentOut.
type ProjectOut struct {
	Project
	DeliveryWarning string `json:"deliveryWarning"`
}

// CardBranch est la branche de feature d'un chantier sur un dépôt donné :
// un chantier qui touche deux dépôts a deux CardBranch, et livrera deux pull
// requests. Créée à la première tâche du chantier sur ce dépôt, avec son
// worktree dédié (les tâches partent de cette branche et y sont fusionnées à
// l'acceptation). PrURL et ShippedAt sont renseignés par la livraison.
type CardBranch struct {
	RepoName    string     `json:"repoName"`
	Branch      string     `json:"branch"`
	Base        string     `json:"base"`
	WorktreeDir string     `json:"worktreeDir"`
	PrURL       string     `json:"prUrl"`
	ShippedAt   *time.Time `json:"shippedAt"`
}

// Card est un chantier (vocabulaire produit) : regroupe des tâches dans une
// colonne du kanban d'un projet. Le nom technique Card/cards ne change pas.
type Card struct {
	ID            string  `json:"id"`
	ProjectID     string  `json:"projectId"`
	Ref           int     `json:"ref"` // référence courte du projet, utilisée dans le nom de branche
	Column        string  `json:"column"`
	Title         string  `json:"title"`
	TasksTotal    int     `json:"tasksTotal"`
	TasksDone     int     `json:"tasksDone"`
	DocsCount     int     `json:"docsCount"`
	MessagesCount int     `json:"messagesCount"`
	ReviewCount   int     `json:"reviewCount"`
	Progress      int     `json:"progress"`
	LiveActivity  *string `json:"liveActivity"`

	// Branches porte une entrée par dépôt touché par le chantier (voir CardBranch).
	Branches []CardBranch `json:"branches"`

	// ShipReady et ShipBlocker sont dérivés (recalculés dans recomputeCard) :
	// le bouton de livraison n'est actif que si le chantier a au moins une
	// tâche, qu'au moins une est acceptée, et qu'une branche de chantier
	// existe. Des tâches encore en cours ou à relire ne bloquent pas : on livre
	// le travail accepté et on relivre le reste ensuite. ShipBlocker nomme la
	// raison du blocage, vide sinon.
	ShipReady   bool   `json:"shipReady"`
	ShipBlocker string `json:"shipBlocker"` // ""|no-tasks|nothing-accepted|nothing-to-ship

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
	Warning string      `json:"warning"`
	Quota   *AgentQuota `json:"quota,omitempty"`
}

// AgentQuotaWindow est une fenêtre glissante de consommation de quota chez le
// fournisseur du cli (ex : 5h, hebdomadaire).
type AgentQuotaWindow struct {
	Label       string    `json:"label"` // "5h"|"week"|"<n>m" (fenêtre inattendue)
	UsedPercent float64   `json:"usedPercent"`
	ResetsAt    time.Time `json:"resetsAt"`
}

// AgentQuota est le dernier instantané connu de consommation de quota pour un
// fournisseur cli. Seul codex publie cette information (voir runner.go
// readCodexRateLimits, lue dans le fichier de session codex après chaque
// exécution : le flux `codex exec --json` ne la porte pas). C'est un quota de
// compte OpenAI, donc partagé par tous les agents cli=codex, jamais calculé
// par agent. claude et fake n'ont pas de source : AgentOut.Quota reste nil.
type AgentQuota struct {
	UpdatedAt time.Time          `json:"updatedAt"`
	Windows   []AgentQuotaWindow `json:"windows"`
}

// Check est le résultat d'une vérification projet (ex : go test) pour une tâche.
type Check struct {
	Label string `json:"label"`
	Ok    bool   `json:"ok"`
}

// CommandLogEntry est une commande jouée par l'agent (tool_use), horodatée.
// Alimente l'onglet « Historique » du panneau de tâche (débogage de ce que
// l'agent a réellement exécuté) ; contrairement à Task.LiveActivity, persiste
// au-delà de l'exécution en cours.
type CommandLogEntry struct {
	Text string    `json:"text"`
	At   time.Time `json:"at"`
}

// Task est une tâche assignée à un agent, exécutée dans un worktree git dédié.
type Task struct {
	ID            string  `json:"id"`
	CardID        string  `json:"cardId"`
	ProjectID     string  `json:"projectId"`
	Ref           int     `json:"ref"`
	Title         string  `json:"title"`
	AgentID       string  `json:"agentId"`
	RepoName      string  `json:"repoName"` // dépôt du projet utilisé pour le worktree
	Branch        string  `json:"branch"`
	Status        string  `json:"status"` // running|review|accepted|cancelled
	MessagesCount int     `json:"messagesCount"`
	FilesCount    int     `json:"filesCount"`
	DocsCount     int     `json:"docsCount"`
	CommitsCount  int     `json:"commitsCount"` // commits de base..branche, recalculé à chaque fin d'exécution
	Checks        []Check `json:"checks"`
	LiveActivity  *string `json:"liveActivity"`

	// CommandLog conserve les commandes jouées par l'agent (tool_use), les plus
	// récentes en dernier, plafonné à commandLogLimit entrées (runner.go) : un
	// historique de débogage, pas un audit exhaustif.
	CommandLog []CommandLogEntry `json:"commandLog"`

	Unread    bool      `json:"unread"`
	UpdatedAt time.Time `json:"updatedAt"`
	Tokens    Tokens    `json:"tokens"`

	// Rebasing indique qu'un rebase automatique de cette tâche est en cours sur
	// la branche de son chantier (voir Server.rebaseSiblingTasks) : le frontend
	// affiche un fuseau à la place de l'icône d'état. État volatile, remis à
	// faux au chargement de state.json (resetTransientTaskFlags) pour qu'un
	// arrêt brutal en plein rebase ne laisse pas un fuseau tourner sans fin.
	Rebasing bool `json:"rebasing"`

	// Champs internes ; persistés dans state.json (sinon perdus au redémarrage)
	// et visibles par le client authentifié, ce qui est sans enjeu en mono-utilisateur.
	Base        string `json:"base"`        // branche de base : la branche du chantier
	WorktreeDir string `json:"worktreeDir"` //
	SessionID   string `json:"sessionId"`   // session claude, pour --resume
}

// Message est un message échangé dans le fil d'une tâche. AuthorName porte
// le nom de l'agent pour author="agent" ; pour author="user", c'est le
// displayName des Settings (vide si non renseigné, le frontend affiche alors
// "Vous"/"You").
// PreviewRun est une exécution de recette manuelle : la commande d'un dépôt
// lancée dans le worktree d'un chantier ou d'une tâche. Jamais persisté (le
// champ ne figure pas dans Store) : les runs et leur journal vivent en mémoire
// le temps de la session du serveur.
type PreviewRun struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	CardID    string `json:"cardId"`
	TaskID    string `json:"taskId"` // vide pour un run de chantier
	RepoName  string `json:"repoName"`

	Cmd string `json:"cmd"` // la commande telle qu'elle a été lancée
	URL string `json:"url"` // previewUrl du dépôt, variables substituées
	Dir string `json:"dir"` // worktree d'exécution, jamais le dépôt du projet

	Status   string `json:"status"` // running|exited|stopped|failed
	ExitCode int    `json:"exitCode"`
	Error    string `json:"error"` // renseigné quand le lancement a échoué

	StartedAt time.Time  `json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt"`
}

// PreviewLogEvent est le contenu de l'événement SSE "previewLog" : une ligne de
// sortie d'un run de recette.
type PreviewLogEvent struct {
	RunID string `json:"runId"`
	Line  string `json:"line"`
}

type Message struct {
	ID         string    `json:"id"`
	TaskID     string    `json:"taskId"`
	Author     string    `json:"author"` // user|agent
	AuthorName string    `json:"authorName"`
	Text       string    `json:"text"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Workspace est l'état de synchronisation git persisté de l'espace de
// données (dataDir), stocké dans state.json. AutoSync active la
// synchronisation périodique automatique (voir Server.autoSyncTick) ;
// l'activer exige que git soit initialisé et qu'un remote soit défini.
type Workspace struct {
	SetupDone  bool       `json:"setupDone"`
	SyncRemote string     `json:"syncRemote"`
	LastSyncAt *time.Time `json:"lastSyncAt"`
	AutoSync   bool       `json:"autoSync"`
}

// WorkspaceStatus est la réponse de GET /api/workspace, le contenu de
// l'événement SSE "workspace" et le champ State.Workspace : elle combine
// l'état persisté (Workspace) et des faits git calculés à la volée.
// LastSyncError n'est jamais persisté (état en mémoire du serveur, remis à
// zéro à chaque redémarrage) : dernier message d'échec de la synchronisation
// automatique, vide si le dernier tick (ou la dernière sync manuelle) a
// réussi.
type WorkspaceStatus struct {
	SetupDone     bool       `json:"setupDone"`
	GitEnabled    bool       `json:"gitEnabled"`
	Remote        string     `json:"remote"`
	Dirty         bool       `json:"dirty"`
	LastCommitAt  *time.Time `json:"lastCommitAt"`
	LastSyncAt    *time.Time `json:"lastSyncAt"`
	AutoSync      bool       `json:"autoSync"`
	LastSyncError string     `json:"lastSyncError"`
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
	// UpdateCheck : vérification périodique des mises à jour. Pointeur pour
	// que "absent" (state.json écrit avant cette fonctionnalité) veuille dire
	// "activé", sans migration.
	UpdateCheck *bool `json:"updateCheck"`
}

// UpdateStatus décrit les mises à jour de Sillage (GET /api/update, champ
// `update` de GET /api/state, événement SSE "update"). Jamais persisté : ce
// n'est pas un état du produit mais une observation du binaire courant.
type UpdateStatus struct {
	Current   string `json:"current"`   // version en cours d'exécution, "dev" en local
	Latest    string `json:"latest"`    // dernière version publiée, "" si inconnue
	Available bool   `json:"available"` // latest > current
	// Method : "brew" | "binary" | "go" | "unknown" | "dev"
	Method        string     `json:"method"`
	Path          string     `json:"path"`          // binaire en cours d'exécution
	SelfUpdatable bool       `json:"selfUpdatable"` // POST /api/update/apply peut aboutir
	Command       string     `json:"command"`       // la mise à jour à la main, toujours renseignée
	ReleaseURL    string     `json:"releaseUrl"`
	CheckEnabled  bool       `json:"checkEnabled"`
	CheckedAt     *time.Time `json:"checkedAt"`
	Error         string     `json:"error"` // échec de la dernière vérification
	Applying      bool       `json:"applying"`
	Blocker       string     `json:"blocker"` // "" si un clic suffit, sinon la raison
}

// UpdateApplyResponse est la réponse de POST /api/update/apply.
type UpdateApplyResponse struct {
	Output     string `json:"output"`
	Version    string `json:"version"`
	Restarting bool   `json:"restarting"` // faux : mise à jour posée, redémarrage à faire à la main
	Note       string `json:"note"`
}

// State est la réponse de GET /api/state.
type State struct {
	Projects  []ProjectOut    `json:"projects"`
	Cards     []Card          `json:"cards"`
	Tasks     []Task          `json:"tasks"`
	Agents    []AgentOut      `json:"agents"`
	Workspace WorkspaceStatus `json:"workspace"`
	Settings  Settings        `json:"settings"`
	Previews  []PreviewRun    `json:"previews"`
	Update    UpdateStatus    `json:"update"`
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

// TaskDeletedEvent est le contenu de l'événement SSE "taskDeleted".
type TaskDeletedEvent struct {
	TaskID    string `json:"taskId"`
	CardID    string `json:"cardId"`
	ProjectID string `json:"projectId"`
}

// CardDeletedEvent est le contenu de l'événement SSE "cardDeleted".
type CardDeletedEvent struct {
	CardID    string `json:"cardId"`
	ProjectID string `json:"projectId"`
}

// ProjectDeletedEvent est le contenu de l'événement SSE "projectDeleted".
type ProjectDeletedEvent struct {
	ProjectID string `json:"projectId"`
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

// AcceptResponse est la réponse de POST /api/tasks/{id}/accept : la tâche mise
// à jour et la branche de chantier dans laquelle son travail a été fusionné.
type AcceptResponse struct {
	Task              Task   `json:"task"`
	WorkstreamBranch  string `json:"workstreamBranch"`
	Output            string `json:"output"`
	ConflictFilePaths string `json:"conflictFilePaths,omitempty"`
}

// CatchUpRepoResult est le résultat, sur un dépôt, d'un rattrapage de la
// branche de destination dans la branche du chantier (voir handleCardCatchUp).
type CatchUpRepoResult struct {
	RepoName string `json:"repoName"`
	Target   string `json:"target"`

	Merged   bool `json:"merged"`   // la destination a été fusionnée dans le chantier
	UpToDate bool `json:"upToDate"` // rien à rattraper, la destination y était déjà

	// ConflictFilePaths liste les fichiers en conflit, séparés par des espaces.
	// La fusion est alors annulée : le worktree du chantier revient intact.
	ConflictFilePaths string `json:"conflictFilePaths,omitempty"`
	Output            string `json:"output"`
	Error             string `json:"error"`
}

// CatchUpResponse est la réponse de POST /api/cards/{id}/catch-up. Un dépôt en
// échec n'annule pas les autres : chaque ligne porte son propre résultat.
type CatchUpResponse struct {
	Card  Card                `json:"card"`
	Repos []CatchUpRepoResult `json:"repos"`
}

// DeliveryRepoPreview décrit ce que la livraison ferait sur un dépôt donné.
// Commits/Files (contenu de la livraison, base..branche) et Pending (ce qui
// reste réellement à livrer) sont calculés à la volée par git.
type DeliveryRepoPreview struct {
	RepoName string `json:"repoName"`
	Branch   string `json:"branch"`
	Base     string `json:"base"`
	Commits  int    `json:"commits"`
	Files    int    `json:"files"`

	// Pending est le nombre de commits pas encore livrés : non poussés en mode
	// "pr" (origin/<branche>..<branche>), pas encore fusionnés dans la branche
	// de destination en mode "merge". Zéro partout signifie « rien à livrer ».
	Pending int `json:"pending"`

	// Behind est le nombre de commits que la base a et que la branche du
	// chantier n'a pas (branche..base) : le chantier est en retard sur la
	// release et devra être rebasé avant d'être livré proprement.
	Behind int `json:"behind"`

	// MergedIntoTarget dit que la branche du chantier est entièrement contenue
	// dans la branche de destination : tout est arrivé à destination, il n'y a
	// plus rien à livrer, que ce soit par Sillage ou à la main. L'UI remplace
	// alors le bouton par « Déjà sur <destination> ».
	MergedIntoTarget bool `json:"mergedIntoTarget"`

	// FastForwardable dit qu'une fusion fast-forward de la branche du chantier
	// dans la destination est possible (la destination est un ancêtre de la
	// branche). Faux avec Behind > 0 : les deux modes de fusion refuseraient,
	// l'UI désactive donc le bouton plutôt que d'annoncer un échec certain.
	FastForwardable bool `json:"fastForwardable"`

	PrURL     string     `json:"prUrl"`
	ShippedAt *time.Time `json:"shippedAt"`
}

// DeliveryPreview est la réponse de GET /api/cards/{id}/delivery : tout ce
// qu'il faut pour annoncer la livraison AVANT de la déclencher (sous-texte du
// bouton et panneau de récapitulatif). Lecture seule, ne pousse jamais rien.
type DeliveryPreview struct {
	Mode     string                `json:"mode"`
	Target   string                `json:"target"`
	Provider string                `json:"provider"` // github|gitlab|"" (déduit du remote origin)
	Ready    bool                  `json:"ready"`
	Blocker  string                `json:"blocker"`
	Warnings []string              `json:"warnings"`
	Counts   DeliveryCounts        `json:"counts"`
	Repos    []DeliveryRepoPreview `json:"repos"`

	// Behind indique, par identifiant de tâche encore vivante (running ou
	// review), le nombre de commits que la branche du chantier a et que la
	// branche de la tâche n'a pas : le retard qui produira un conflit à
	// l'acceptation. Une tâche à jour n'apparaît pas dans la table.
	Behind map[string]int `json:"behind"`
}

// DeliveryCounts compte les tâches du chantier par état, pour la ligne d'état
// affichée au-dessus du bouton de livraison.
type DeliveryCounts struct {
	Accepted int `json:"accepted"`
	Refused  int `json:"refused"`
	Pending  int `json:"pending"`
}

// ShipRepoResult est le résultat de la livraison d'un dépôt. Un dépôt en échec
// n'annule pas les autres : chaque ligne porte sa propre erreur.
type ShipRepoResult struct {
	RepoName string `json:"repoName"`
	Branch   string `json:"branch"`
	Base     string `json:"base"`
	Pushed   bool   `json:"pushed"`
	Merged   bool   `json:"merged"`
	Skipped  bool   `json:"skipped"` // rien à livrer (aucun commit sur la branche)
	PrURL    string `json:"prUrl"`
	Output   string `json:"output"`
	Error    string `json:"error"`
}

// ShipResponse est la réponse de POST /api/cards/{id}/ship.
type ShipResponse struct {
	Card  Card             `json:"card"`
	Mode  string           `json:"mode"`
	Repos []ShipRepoResult `json:"repos"`
}
