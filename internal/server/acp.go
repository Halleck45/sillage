package server

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
)

// Client ACP (Agent Client Protocol, https://agentclientprotocol.com) : du
// JSON-RPC 2.0 sur stdio, un objet par ligne. Sillage y tient le rôle de
// l'éditeur : il ouvre une session, envoie un prompt, lit les mises à jour au
// fil de l'eau et répond aux demandes de l'agent.
//
// Ce que le client ne propose pas, volontairement : ni accès fichiers
// (fs/read_text_file, fs/write_text_file) ni terminal. L'agent garde ses
// propres outils, bornés par ses règles à lui ; toute méthode non servie
// reçoit « method not found », ce que le protocole prévoit pour une capacité
// non annoncée.

const acpProtocolVersion = 1

// errACPClosed : le flux de l'agent s'est fermé avant la réponse attendue.
var errACPClosed = errors.New("acp: agent closed the connection")

// acpError est l'erreur JSON-RPC renvoyée par l'agent.
type acpError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *acpError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("acp error %d", e.Code)
	}
	return e.Message
}

// acpAuthRequired est le code que le protocole réserve à un agent non connecté.
const acpAuthRequired = -32000

type acpEnvelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *acpError       `json:"error,omitempty"`
}

// acpUpdate est une notification session/update. Selon `sessionUpdate`, seuls
// certains champs sont remplis. Content reste brut : c'est un bloc de contenu
// pour un fragment de message, une liste pour un appel d'outil.
type acpUpdate struct {
	Kind       string          `json:"sessionUpdate"`
	Content    json.RawMessage `json:"content,omitempty"`
	ToolCallID string          `json:"toolCallId,omitempty"`
	Title      string          `json:"title,omitempty"`
	ToolKind   string          `json:"kind,omitempty"`
	Status     string          `json:"status,omitempty"`
	RawInput   json.RawMessage `json:"rawInput,omitempty"`
	RawOutput  json.RawMessage `json:"rawOutput,omitempty"`
	Locations  []acpLocation   `json:"locations,omitempty"`
}

type acpLocation struct {
	Path string `json:"path"`
}

// text rend le texte d'un fragment de message (bloc de contenu de type text).
func (u acpUpdate) text() string {
	var block struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(u.Content, &block) != nil || block.Type != "text" {
		return ""
	}
	return block.Text
}

// acpPermissionOption est un choix proposé par l'agent pour une demande
// d'autorisation. Kind vaut allow_once, allow_always, reject_once ou
// reject_always.
type acpPermissionOption struct {
	ID   string `json:"optionId"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type acpPermissionRequest struct {
	SessionID string                `json:"sessionId"`
	ToolCall  acpUpdate             `json:"toolCall"`
	Options   []acpPermissionOption `json:"options"`
}

// acpPermissionPolicy choisit l'option retenue pour une demande. Une chaîne
// vide annule la demande (outcome "cancelled").
type acpPermissionPolicy func(acpPermissionRequest) string

// acpRejectAll refuse toute demande. C'est la politique des agents qui
// tournent seuls : ce qu'ils ont le droit de faire est fixé d'avance par leurs
// propres règles, donc une demande qui remonte jusqu'ici sort de ce cadre et
// personne n'est là pour en juger. Sillage n'accorde jamais rien en cours de
// route (invariant 1 : aucune porte vers un push ne s'ouvre par ce chemin).
func acpRejectAll(req acpPermissionRequest) string {
	for _, kind := range []string{"reject_once", "reject_always"} {
		for _, option := range req.Options {
			if option.Kind == kind {
				return option.ID
			}
		}
	}
	return ""
}

type acpInitializeResult struct {
	ProtocolVersion   int `json:"protocolVersion"`
	AgentCapabilities struct {
		LoadSession bool `json:"loadSession"`
	} `json:"agentCapabilities"`
	AgentInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"agentInfo"`
}

type acpUsage struct {
	InputTokens       int `json:"inputTokens"`
	OutputTokens      int `json:"outputTokens"`
	CachedReadTokens  int `json:"cachedReadTokens"`
	CachedWriteTokens int `json:"cachedWriteTokens"`
}

type acpPromptResult struct {
	StopReason string    `json:"stopReason"`
	Usage      *acpUsage `json:"usage"`
}

// acpHandlers reçoit ce que l'agent envoie de lui-même. Les appels arrivent
// dans l'ordre du flux, depuis la goroutine de lecture : un handler ne doit
// pas attendre une réponse de l'agent, il bloquerait la lecture de celle-ci.
type acpHandlers struct {
	Update       func(sessionID string, update acpUpdate)
	Permission   acpPermissionPolicy
	Notification func(method string, params json.RawMessage) // extensions propres à un agent
}

type acpClient struct {
	handlers acpHandlers

	writeMu sync.Mutex
	w       io.Writer

	mu      sync.Mutex
	nextID  int64
	pending map[string]chan acpEnvelope
	closed  bool
	done    chan struct{}
}

// newACPClient démarre la lecture de r. Elle s'arrête à la fermeture du flux,
// ce qui débloque tous les appels en attente avec errACPClosed.
func newACPClient(r io.Reader, w io.Writer, handlers acpHandlers) *acpClient {
	c := &acpClient{
		handlers: handlers,
		w:        w,
		pending:  map[string]chan acpEnvelope{},
		done:     make(chan struct{}),
	}
	go c.readLoop(r)
	return c
}

func (c *acpClient) readLoop(r io.Reader) {
	defer func() {
		c.mu.Lock()
		c.closed = true
		for id, ch := range c.pending {
			close(ch)
			delete(c.pending, id)
		}
		c.mu.Unlock()
		close(c.done)
	}()
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 32*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var env acpEnvelope
		// Une ligne hors JSON (bannière, trace) n'est pas du protocole.
		if json.Unmarshal(line, &env) != nil {
			continue
		}
		switch {
		case env.Method != "" && len(env.ID) > 0:
			c.handleRequest(env)
		case env.Method != "":
			c.handleNotification(env)
		case len(env.ID) > 0:
			c.mu.Lock()
			ch := c.pending[string(env.ID)]
			delete(c.pending, string(env.ID))
			c.mu.Unlock()
			if ch != nil {
				ch <- env
			}
		}
	}
}

func (c *acpClient) handleNotification(env acpEnvelope) {
	if env.Method == "session/update" {
		var params struct {
			SessionID string    `json:"sessionId"`
			Update    acpUpdate `json:"update"`
		}
		if json.Unmarshal(env.Params, &params) == nil && c.handlers.Update != nil {
			c.handlers.Update(params.SessionID, params.Update)
		}
		return
	}
	if c.handlers.Notification != nil {
		c.handlers.Notification(env.Method, env.Params)
	}
}

func (c *acpClient) handleRequest(env acpEnvelope) {
	if env.Method != "session/request_permission" {
		c.reply(env.ID, nil, &acpError{Code: -32601, Message: "method not found: " + env.Method})
		return
	}
	var req acpPermissionRequest
	if err := json.Unmarshal(env.Params, &req); err != nil {
		c.reply(env.ID, nil, &acpError{Code: -32602, Message: "invalid params"})
		return
	}
	optionID := ""
	if c.handlers.Permission != nil {
		optionID = c.handlers.Permission(req)
	}
	outcome := map[string]any{"outcome": "cancelled"}
	if optionID != "" {
		outcome = map[string]any{"outcome": "selected", "optionId": optionID}
	}
	c.reply(env.ID, map[string]any{"outcome": outcome}, nil)
}

func (c *acpClient) write(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_, err = c.w.Write(append(data, '\n'))
	return err
}

func (c *acpClient) reply(id json.RawMessage, result any, rpcErr *acpError) {
	msg := map[string]any{"jsonrpc": "2.0", "id": id}
	if rpcErr != nil {
		msg["error"] = rpcErr
	} else {
		msg["result"] = result
	}
	_ = c.write(msg)
}

// notify envoie une notification : aucune réponse n'est attendue.
func (c *acpClient) notify(method string, params any) error {
	return c.write(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

// call envoie une requête et attend sa réponse, la fin du flux ou celle de ctx.
func (c *acpClient) call(ctx context.Context, method string, params, out any) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return errACPClosed
	}
	c.nextID++
	id := c.nextID
	key := fmt.Sprintf("%d", id)
	ch := make(chan acpEnvelope, 1)
	c.pending[key] = ch
	c.mu.Unlock()

	if err := c.write(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params}); err != nil {
		c.mu.Lock()
		delete(c.pending, key)
		c.mu.Unlock()
		return err
	}
	select {
	case env, ok := <-ch:
		if !ok {
			return errACPClosed
		}
		if env.Error != nil {
			return env.Error
		}
		if out == nil || len(env.Result) == 0 {
			return nil
		}
		return json.Unmarshal(env.Result, out)
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, key)
		c.mu.Unlock()
		return ctx.Err()
	}
}

// initialize négocie la version et annonce ce que Sillage sait faire : rien
// côté fichiers ni terminal. meta porte les extensions propres à un agent
// (clientCapabilities._meta).
func (c *acpClient) initialize(ctx context.Context, meta map[string]any) (acpInitializeResult, error) {
	capabilities := map[string]any{
		"fs":       map[string]any{"readTextFile": false, "writeTextFile": false},
		"terminal": false,
	}
	if len(meta) > 0 {
		capabilities["_meta"] = meta
	}
	var result acpInitializeResult
	err := c.call(ctx, "initialize", map[string]any{
		"protocolVersion":    acpProtocolVersion,
		"clientCapabilities": capabilities,
		"clientInfo":         map[string]any{"name": "sillage", "title": "Sillage", "version": CurrentVersion()},
	}, &result)
	if err == nil && result.ProtocolVersion != acpProtocolVersion {
		err = fmt.Errorf("unsupported ACP protocol version %d", result.ProtocolVersion)
	}
	return result, err
}

func (c *acpClient) newSession(ctx context.Context, cwd string) (string, error) {
	var result struct {
		SessionID string `json:"sessionId"`
	}
	err := c.call(ctx, "session/new", map[string]any{"cwd": cwd, "mcpServers": []any{}}, &result)
	if err == nil && result.SessionID == "" {
		err = errors.New("acp: session/new returned no session id")
	}
	return result.SessionID, err
}

// loadSession reprend une session. L'agent rejoue alors tout son historique en
// session/update avant de répondre : à l'appelant de ne pas le prendre pour du
// travail neuf.
func (c *acpClient) loadSession(ctx context.Context, sessionID, cwd string) error {
	return c.call(ctx, "session/load", map[string]any{"sessionId": sessionID, "cwd": cwd, "mcpServers": []any{}}, nil)
}

// prompt envoie un tour et rend la main quand l'agent a fini d'agir dessus.
func (c *acpClient) prompt(ctx context.Context, sessionID, text string) (acpPromptResult, error) {
	var result acpPromptResult
	err := c.call(ctx, "session/prompt", map[string]any{
		"sessionId": sessionID,
		"prompt":    []any{map[string]any{"type": "text", "text": text}},
	}, &result)
	return result, err
}

// cancel demande l'arrêt du tour en cours ; le prompt répond alors avec
// stopReason "cancelled".
func (c *acpClient) cancel(sessionID string) {
	_ = c.notify("session/cancel", map[string]any{"sessionId": sessionID})
}

// --- Lecture d'un tour ---

// acpToolCall est l'état connu d'un appel d'outil, reconstitué à partir de
// tool_call puis de ses tool_call_update, qui ne portent que ce qui change.
type acpToolCall struct {
	Title    string
	Kind     string
	RawInput json.RawMessage
	Path     string
	line     string
}

// acpToolLine résume un appel d'outil sur une ligne, au format des autres
// adaptateurs (« outil · détail »). Le détail vient de l'entrée brute quand
// elle porte une clé connue, sinon du titre fourni par l'agent. Les chemins
// sous cwd sont rendus relatifs : le worktree d'une tâche a un chemin long qui
// mangerait toute la ligne.
func acpToolLine(call acpToolCall, cwd string) string {
	label := call.Kind
	if label == "" || label == "other" {
		label = "tool"
	}
	detail := ""
	var input map[string]any
	if json.Unmarshal(call.RawInput, &input) == nil {
		for _, key := range []string{"command", "file_path", "path", "fileName", "pattern", "query", "url"} {
			if v, ok := input[key].(string); ok && v != "" {
				detail = v
				break
			}
		}
	}
	if detail == "" {
		detail = call.Path
	}
	if detail == "" {
		detail = call.Title
	}
	if cwd != "" {
		if rel, err := filepath.Rel(cwd, detail); err == nil && filepath.IsAbs(detail) && !strings.HasPrefix(rel, "..") {
			detail = rel
		}
	}
	detail = strings.Join(strings.Fields(detail), " ")
	if detail == "" {
		return label
	}
	return label + " · " + truncate(detail, 80)
}

// acpDenied reconnaît un outil refusé dans la sortie brute d'un appel en
// échec : les agents qui appliquent leurs propres règles le signalent par un
// code "denied".
func acpDenied(rawOutput json.RawMessage) bool {
	var out struct {
		Code string `json:"code"`
	}
	return json.Unmarshal(rawOutput, &out) == nil && out.Code == "denied"
}

// acpTurn transforme le flux d'un tour en ce que Sillage affiche : des
// messages (le texte de l'agent, coupé à chaque appel d'outil comme dans le
// fil de claude), une ligne d'activité par outil, un marqueur par outil
// refusé. Ses méthodes sont appelées depuis la goroutine de lecture du
// client ; le verrou couvre la bascule faite depuis celle du runner.
type acpTurn struct {
	cwd        string
	onMessage  func(text string)
	onActivity func(line string)
	onDenied   func(line string)

	mu      sync.Mutex
	active  bool // faux hors prompt : rejeu d'historique, bavardage de démarrage
	started bool // l'agent a annoncé le début de son tour (discardPrelude)
	text    strings.Builder
	calls   map[string]*acpToolCall
	denied  map[string]bool
}

func newACPTurn(cwd string) *acpTurn {
	return &acpTurn{cwd: cwd, calls: map[string]*acpToolCall{}, denied: map[string]bool{}}
}

// begin ouvre le tour : ce qui arrive avant n'appartient pas à la réponse.
func (t *acpTurn) begin() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.active = true
	t.started = false
	t.text.Reset()
}

// discardPrelude jette le texte reçu depuis begin, une seule fois par tour.
// Certains agents écrivent des avertissements de démarrage dans le canal des
// messages ; quand l'agent annonce le vrai début de son tour, ce qui précède
// part ici. Les annonces suivantes du même tour (un agent en fait une par
// appel au modèle) ne jettent plus rien : ce serait de la réponse.
func (t *acpTurn) discardPrelude() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active || t.started {
		return
	}
	t.started = true
	t.text.Reset()
}

// end ferme le tour et publie le texte en attente.
func (t *acpTurn) end() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.flushLocked()
	t.active = false
}

// abort ferme un tour annulé sans publier le texte en attente : ce qui suit le
// dernier outil est alors une phrase coupée, ou l'accusé de réception de
// l'annulation (« Operation cancelled by user »), pas une réponse.
func (t *acpTurn) abort() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.text.Reset()
	t.active = false
}

func (t *acpTurn) flushLocked() {
	text := strings.TrimSpace(t.text.String())
	t.text.Reset()
	if text != "" && t.onMessage != nil {
		t.onMessage(text)
	}
}

func (t *acpTurn) handle(update acpUpdate) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return
	}
	switch update.Kind {
	case "agent_message_chunk":
		t.text.WriteString(update.text())
	case "tool_call", "tool_call_update":
		t.toolLocked(update)
	}
}

func (t *acpTurn) toolLocked(update acpUpdate) {
	if update.ToolCallID == "" {
		return
	}
	call, known := t.calls[update.ToolCallID]
	if !known {
		call = &acpToolCall{}
		t.calls[update.ToolCallID] = call
		// Un outil sépare deux prises de parole : ce que l'agent a dit avant
		// devient un message, sans attendre la fin du tour.
		t.flushLocked()
	}
	if update.Title != "" {
		call.Title = update.Title
	}
	if update.ToolKind != "" {
		call.Kind = update.ToolKind
	}
	if len(update.RawInput) > 0 {
		call.RawInput = update.RawInput
	}
	if len(update.Locations) > 0 {
		call.Path = update.Locations[0].Path
	}
	if line := acpToolLine(*call, t.cwd); line != call.line {
		call.line = line
		if !known && t.onActivity != nil {
			t.onActivity(line)
		}
	}
	if update.Status == "failed" && acpDenied(update.RawOutput) {
		t.denyLocked(call.line)
	}
}

// permission applique la politique à une demande et note le refus : un outil
// que Sillage vient de refuser mérite le même marqueur qu'un outil refusé par
// les règles de l'agent.
func (t *acpTurn) permission(policy acpPermissionPolicy, req acpPermissionRequest) string {
	optionID := policy(req)
	granted := false
	for _, option := range req.Options {
		if option.ID == optionID && strings.HasPrefix(option.Kind, "allow") {
			granted = true
		}
	}
	if granted {
		return optionID
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return optionID
	}
	t.toolLocked(req.ToolCall)
	if call := t.calls[req.ToolCall.ToolCallID]; call != nil {
		t.denyLocked(call.line)
	}
	return optionID
}

// denyLocked signale un outil refusé, une seule fois par outil et par tour :
// un agent qui réessaie trois fois inonderait le fil.
func (t *acpTurn) denyLocked(line string) {
	line = strings.ReplaceAll(line, "]", "")
	if line == "" || t.denied[line] {
		return
	}
	t.denied[line] = true
	if t.onDenied != nil {
		t.onDenied(line)
	}
}
