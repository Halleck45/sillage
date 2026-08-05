package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStoreRoundtripSaveLoad(t *testing.T) {
	dir := t.TempDir()

	s1, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	project, err := s1.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s1.AddCard(project.ID, "Ma carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	if card.Column != "soon" {
		t.Fatalf("colonne par défaut attendue 'soon', reçue %q", card.Column)
	}

	taskID, ref := s1.ReserveTaskID()
	if ref < 100 {
		t.Fatalf("la référence doit démarrer à 100 ou plus, reçue %d", ref)
	}
	task, err := s1.CreateTask(taskID, ref, card.ID, project.ID, "Titre de tâche", "echo", "sillage/100-titre", "main", "/tmp/worktree", "sillage")
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if task.RepoName != "sillage" {
		t.Fatalf("repoName attendu 'sillage', reçu %q", task.RepoName)
	}
	if task.Status != "running" {
		t.Fatalf("statut initial attendu 'running', reçu %q", task.Status)
	}

	task, err = s1.UpdateTask(taskID, func(tk *Task) {
		tk.Status = "accepted"
		tk.Unread = true
		tk.Tokens = Tokens{Input: 10, Output: 5, CostUsd: 0.01}
		tk.SessionID = "sess-abc"
	})
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	// Mono-utilisateur : authorName reste vide pour les messages "user".
	msg, task, err := s1.AddMessage(taskID, "user", "", "bonjour")
	if err != nil {
		t.Fatalf("AddMessage: %v", err)
	}
	if task.MessagesCount != 1 {
		t.Fatalf("messagesCount attendu 1, reçu %d", task.MessagesCount)
	}
	if msg.Author != "user" || msg.Text != "bonjour" || msg.AuthorName != "" {
		t.Fatalf("message inattendu : %+v", msg)
	}

	// Recharge depuis le disque : tout doit être identique.
	s2, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore (reload): %v", err)
	}

	loadedTask, ok := s2.GetTask(taskID)
	if !ok {
		t.Fatalf("tâche %s introuvable après rechargement", taskID)
	}
	if loadedTask.Status != "accepted" {
		t.Fatalf("statut attendu 'accepted' après rechargement, reçu %q", loadedTask.Status)
	}
	if loadedTask.Tokens.Input != 10 || loadedTask.Tokens.Output != 5 || loadedTask.Tokens.CostUsd != 0.01 {
		t.Fatalf("tokens inattendus après rechargement : %+v", loadedTask.Tokens)
	}
	// Régression : les champs internes doivent survivre au redémarrage,
	// sinon diff et resume sont cassés après relance du serveur.
	if loadedTask.Base != "main" || loadedTask.WorktreeDir != "/tmp/worktree" || loadedTask.SessionID != "sess-abc" {
		t.Fatalf("champs internes perdus après rechargement : base=%q worktree=%q session=%q",
			loadedTask.Base, loadedTask.WorktreeDir, loadedTask.SessionID)
	}
	if loadedTask.MessagesCount != 1 {
		t.Fatalf("messagesCount attendu 1 après rechargement, reçu %d", loadedTask.MessagesCount)
	}

	msgs := s2.GetMessages(taskID)
	if len(msgs) != 1 || msgs[0].Text != "bonjour" {
		t.Fatalf("messages inattendus après rechargement : %+v", msgs)
	}

	loadedProject, ok := s2.GetProject(project.ID)
	if !ok {
		t.Fatalf("projet %s introuvable après rechargement", project.ID)
	}
	if len(loadedProject.Repos) != 1 || loadedProject.Repos[0].Path != "/tmp/sillage" {
		t.Fatalf("repos de projet inattendus après rechargement : %+v", loadedProject.Repos)
	}

	// Les agents seedés doivent survivre au rechargement.
	if _, ok := s2.GetAgent("bolt"); !ok {
		t.Fatalf("l'agent seedé 'bolt' doit être présent après rechargement")
	}
	for id, wantName := range map[string]string{"github-copilot": "Octo", "antigravity": "Astro"} {
		if agent, ok := s2.GetAgent(id); !ok {
			t.Fatalf("seeded agent %q should be present after reload", id)
		} else if agent.Name != wantName {
			t.Fatalf("seeded agent %q name = %q, want %q", id, agent.Name, wantName)
		}
	}
}

func TestAgentSeedMigrationRunsOnce(t *testing.T) {
	dir := t.TempDir()
	legacy := `{
  "FormatVersion": 1,
  "Projects": {}, "Cards": {}, "Tasks": {}, "Messages": {},
  "Agents": {
    "github-copilot": {"id":"github-copilot","name":"Custom Copilot","cli":"copilot"}
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}

	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if s.AgentSeedVersion != agentSeedVersion {
		t.Fatalf("agent seed version = %d, want %d", s.AgentSeedVersion, agentSeedVersion)
	}
	if agent, ok := s.GetAgent("github-copilot"); !ok || agent.Name != "Custom Copilot" {
		t.Fatalf("existing seeded ID should remain untouched, got %+v", agent)
	}
	if agent, ok := s.GetAgent("antigravity"); !ok {
		t.Fatal("Antigravity should be added to an existing workspace")
	} else if agent.Name != "Astro" {
		t.Fatalf("Antigravity seeded name = %q, want Astro", agent.Name)
	}

	if err := s.DeleteAgent("antigravity"); err != nil {
		t.Fatalf("DeleteAgent: %v", err)
	}
	reloaded, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore after deletion: %v", err)
	}
	if _, ok := reloaded.GetAgent("antigravity"); ok {
		t.Fatal("a deleted seeded agent should not be recreated")
	}
}

func TestDerivedCounters(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	project, err := s.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	statuses := []string{"accepted", "review", "running"}
	var taskIDs []string
	for _, st := range statuses {
		id, ref := s.ReserveTaskID()
		task, err := s.CreateTask(id, ref, card.ID, project.ID, "T "+st, "echo", "sillage/"+id, "main", "/tmp/wt-"+id, "sillage")
		if err != nil {
			t.Fatalf("CreateTask: %v", err)
		}
		if _, err := s.UpdateTask(task.ID, func(tk *Task) {
			tk.Status = st
			tk.Unread = st != "running"
			tk.DocsCount = 1
			line := "en cours"
			if st == "running" {
				tk.LiveActivity = &line
			}
		}); err != nil {
			t.Fatalf("UpdateTask: %v", err)
		}
		taskIDs = append(taskIDs, id)
	}
	_ = taskIDs

	updatedCard, ok := s.GetCard(card.ID)
	if !ok {
		t.Fatalf("carte introuvable")
	}
	if updatedCard.TasksTotal != 3 {
		t.Fatalf("tasksTotal attendu 3, reçu %d", updatedCard.TasksTotal)
	}
	if updatedCard.TasksDone != 1 {
		t.Fatalf("tasksDone attendu 1, reçu %d", updatedCard.TasksDone)
	}
	if updatedCard.ReviewCount != 1 {
		t.Fatalf("reviewCount attendu 1, reçu %d", updatedCard.ReviewCount)
	}
	if updatedCard.Progress != 33 {
		t.Fatalf("progress attendu 33, reçu %d", updatedCard.Progress)
	}
	if updatedCard.DocsCount != 3 {
		t.Fatalf("docsCount attendu 3, reçu %d", updatedCard.DocsCount)
	}
	if updatedCard.LiveActivity == nil || *updatedCard.LiveActivity != "en cours" {
		t.Fatalf("liveActivity attendu 'en cours', reçu %v", updatedCard.LiveActivity)
	}

	updatedProject, ok := s.GetProject(project.ID)
	if !ok {
		t.Fatalf("projet introuvable")
	}
	if updatedProject.Unread != 2 {
		t.Fatalf("unread attendu 2, reçu %d", updatedProject.Unread)
	}

	agent, ok := s.GetAgent("echo")
	if !ok {
		t.Fatalf("agent echo introuvable")
	}
	if !agent.Active {
		t.Fatalf("l'agent echo devrait être actif (une tâche running lui est assignée)")
	}

	otherAgent, ok := s.GetAgent("bolt")
	if !ok {
		t.Fatalf("agent bolt introuvable")
	}
	if otherAgent.Active {
		t.Fatalf("l'agent bolt ne devrait pas être actif")
	}

	// progress à 0 si aucune tâche.
	emptyCard, err := s.AddCard(project.ID, "Vide", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	if emptyCard.Progress != 0 || emptyCard.TasksTotal != 0 {
		t.Fatalf("carte vide inattendue : %+v", emptyCard)
	}
}

// --- Lecture d'une tâche : ne doit jamais faire remonter la liste (v0.3.5) ---

func TestMarkTaskReadDoesNotBumpUpdatedAt(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	id, ref := s.ReserveTaskID()
	task, err := s.CreateTask(id, ref, card.ID, project.ID, "T", "echo", "sillage/"+id, "main", "/tmp/wt", "p")
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	task, err = s.UpdateTask(id, func(tk *Task) { tk.Unread = true })
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	before := task.UpdatedAt

	time.Sleep(2 * time.Millisecond) // rend un éventuel bump détectable
	updated, err := s.MarkTaskRead(id)
	if err != nil {
		t.Fatalf("MarkTaskRead: %v", err)
	}
	if updated.Unread {
		t.Fatalf("Unread devrait être false après MarkTaskRead")
	}
	if !updated.UpdatedAt.Equal(before) {
		t.Fatalf("UpdatedAt ne devrait pas changer : avant=%v après=%v", before, updated.UpdatedAt)
	}

	// Par contraste, UpdateTask (utilisé par les autres mutations) bump bien UpdatedAt.
	time.Sleep(2 * time.Millisecond)
	bumped, err := s.UpdateTask(id, func(tk *Task) { tk.Unread = true })
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if !bumped.UpdatedAt.After(before) {
		t.Fatalf("UpdateTask devrait mettre à jour UpdatedAt : avant=%v après=%v", before, bumped.UpdatedAt)
	}
}

// --- Tout marquer comme lu (menu "..." d'un projet) ---

func TestMarkAllTasksReadForProject(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	other, err := s.AddProject("q", "", "", []Repo{{Path: "/tmp/q"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	otherCard, err := s.AddCard(other.ID, "Autre carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	mkTask := func(cardID, projectID string, unread bool) (Task, time.Time) {
		id, ref := s.ReserveTaskID()
		task, err := s.CreateTask(id, ref, cardID, projectID, "T", "echo", "sillage/"+id, "main", "/tmp/wt", "p")
		if err != nil {
			t.Fatalf("CreateTask: %v", err)
		}
		task, err = s.UpdateTask(id, func(tk *Task) { tk.Unread = unread })
		if err != nil {
			t.Fatalf("UpdateTask: %v", err)
		}
		return task, task.UpdatedAt
	}

	unreadTask, beforeUnread := mkTask(card.ID, project.ID, true)
	readTask, beforeRead := mkTask(card.ID, project.ID, false)
	otherProjectTask, beforeOther := mkTask(otherCard.ID, other.ID, true)

	time.Sleep(2 * time.Millisecond) // rend un éventuel bump de UpdatedAt détectable

	changed, err := s.MarkAllTasksReadForProject(project.ID)
	if err != nil {
		t.Fatalf("MarkAllTasksReadForProject: %v", err)
	}
	if len(changed) != 1 || changed[0].ID != unreadTask.ID {
		t.Fatalf("attendu une seule tâche modifiée (%s), reçu %+v", unreadTask.ID, changed)
	}
	if changed[0].Unread {
		t.Fatalf("la tâche non lue devrait passer à Unread=false")
	}
	if !changed[0].UpdatedAt.Equal(beforeUnread) {
		t.Fatalf("UpdatedAt ne devrait pas changer : avant=%v après=%v", beforeUnread, changed[0].UpdatedAt)
	}

	stillRead, ok := s.GetTask(readTask.ID)
	if !ok || stillRead.Unread || !stillRead.UpdatedAt.Equal(beforeRead) {
		t.Fatalf("la tâche déjà lue ne devrait pas être touchée : %+v", stillRead)
	}
	untouched, ok := s.GetTask(otherProjectTask.ID)
	if !ok || !untouched.Unread || !untouched.UpdatedAt.Equal(beforeOther) {
		t.Fatalf("la tâche d'un autre projet ne devrait pas être touchée : %+v", untouched)
	}

	updatedProject, ok := s.GetProject(project.ID)
	if !ok || updatedProject.Unread != 0 {
		t.Fatalf("project.Unread attendu 0, reçu %+v", updatedProject)
	}
	updatedOther, ok := s.GetProject(other.ID)
	if !ok || updatedOther.Unread != 1 {
		t.Fatalf("l'autre projet devrait garder son unread : %+v", updatedOther)
	}

	// Rien à marquer : pas d'erreur, tranche vide.
	changed, err = s.MarkAllTasksReadForProject(project.ID)
	if err != nil || len(changed) != 0 {
		t.Fatalf("attendu aucune tâche modifiée au second appel, reçu %+v (err=%v)", changed, err)
	}

	if _, err := s.MarkAllTasksReadForProject("inconnu"); err == nil {
		t.Fatalf("un projet inconnu devrait renvoyer une erreur")
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Ajouter le bouton Ship": "ajouter-le-bouton-ship",
		"Créer un fichier été":   "creer-un-fichier-ete",
		"  Espaces  multiples ":  "espaces-multiples",
		"":                       "tache",
		"!!!":                    "tache",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, attendu %q", in, got, want)
		}
	}
}

// --- Gardes agents ---

func TestDeleteAgentGuardsReferencedAgent(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	id, ref := s.ReserveTaskID()
	if _, err := s.CreateTask(id, ref, card.ID, project.ID, "T", "echo", "sillage/"+id, "main", "/tmp/wt", "sillage"); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	if err := s.DeleteAgent("echo"); err == nil {
		t.Fatalf("la suppression d'un agent référencé par une tâche devrait être refusée")
	}

	if err := s.DeleteAgent("bolt"); err != nil {
		t.Fatalf("la suppression d'un agent non référencé devrait réussir : %v", err)
	}
	if _, ok := s.GetAgent("bolt"); ok {
		t.Fatalf("l'agent bolt devrait avoir été supprimé")
	}
}

func TestAddAgentSlugUniqueAndValidation(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	if _, err := s.AddAgent("", "", "", "claude", "", ""); err == nil {
		t.Fatalf("un nom vide devrait être refusé")
	}
	if _, err := s.AddAgent("Test", "", "", "invalid-cli", "", ""); err == nil {
		t.Fatalf("un cli invalide devrait être refusé")
	}
	if _, err := s.AddAgent("Bolt", "🐝", "#000", "claude", "", ""); err == nil {
		t.Fatalf("un agent nommé Bolt existe déjà (slug 'bolt'), la création devrait échouer")
	}

	a, err := s.AddAgent("Nova", "✨", "#fff", "fake", "", "prompt")
	if err != nil {
		t.Fatalf("AddAgent: %v", err)
	}
	if a.ID != "nova" {
		t.Fatalf("id attendu 'nova', reçu %q", a.ID)
	}
	for _, cli := range []string{"copilot", "agy"} {
		if _, err := s.AddAgent("Custom "+cli, "", "", cli, "", ""); err != nil {
			t.Fatalf("AddAgent should accept cli %q: %v", cli, err)
		}
	}
}

func TestUpdateProjectFields(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p, err := s.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}

	name := "Nouveau nom"
	checkCmd := "go test ./..."
	updated, err := s.UpdateProject(p.ID, &name, nil, &checkCmd, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if updated.Name != name || updated.CheckCmd != checkCmd {
		t.Fatalf("mise à jour du projet inattendue : %+v", updated)
	}
	if len(updated.Repos) != 1 || updated.Repos[0].Path != "/tmp/sillage" {
		t.Fatalf("repos ne devraient pas changer quand repos=nil : %+v", updated.Repos)
	}

	newRepos := []Repo{{Name: "a", Path: "/tmp/a"}, {Name: "b", Path: "/tmp/b"}}
	updated, err = s.UpdateProject(p.ID, nil, nil, nil, nil, nil, &newRepos, nil, nil)
	if err != nil {
		t.Fatalf("UpdateProject (repos): %v", err)
	}
	if len(updated.Repos) != 2 {
		t.Fatalf("repos attendus au nombre de 2, reçu %+v", updated.Repos)
	}
}

func TestUpdateProjectDescriptionAndContextPrompt(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p, err := s.AddProject("sillage", "Un projet", "Contexte initial", []Repo{{Path: "/tmp/sillage"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if p.Description != "Un projet" || p.ContextPrompt != "Contexte initial" {
		t.Fatalf("description/contextPrompt inattendus à la création : %+v", p)
	}

	desc := "Nouvelle description"
	ctx := "Nouveau contexte"
	updated, err := s.UpdateProject(p.ID, nil, &desc, nil, &ctx, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if updated.Description != desc || updated.ContextPrompt != ctx {
		t.Fatalf("description/contextPrompt inattendus après mise à jour : %+v", updated)
	}
}

// --- Cartes créées uniquement en "soon" (v0.3.1 section 1) ---

func TestAddCardRejectsNonSoonColumn(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}

	if _, err := s.AddCard(project.ID, "Carte", "doing", ""); err == nil {
		t.Fatalf("une carte créée en doing devrait être refusée")
	} else if err.Error() != "cards are created in the soon column" {
		t.Fatalf("message d'erreur inattendu : %q", err.Error())
	}
	if _, err := s.AddCard(project.ID, "Carte", "done", ""); err == nil {
		t.Fatalf("une carte créée en done devrait être refusée")
	}

	c, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard (vide): %v", err)
	}
	if c.Column != "soon" {
		t.Fatalf("colonne attendue 'soon', reçue %q", c.Column)
	}

	c2, err := s.AddCard(project.ID, "Carte2", "soon", "")
	if err != nil {
		t.Fatalf("AddCard (soon explicite): %v", err)
	}
	if c2.Column != "soon" {
		t.Fatalf("colonne attendue 'soon', reçue %q", c2.Column)
	}
}

// --- Contexte de chantier : Card.contextPrompt (v0.3.3 section 1) ---

func TestAddCardWithContextPrompt(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}

	c, err := s.AddCard(project.ID, "Chantier", "", "Contexte du chantier")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	if c.ContextPrompt != "Contexte du chantier" {
		t.Fatalf("contextPrompt attendu 'Contexte du chantier', reçu %q", c.ContextPrompt)
	}
}

func TestUpdateCardTitleAndContextPrompt(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	c, err := s.AddCard(project.ID, "Chantier", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	// Titre vide refusé.
	empty := ""
	if _, err := s.UpdateCard(c.ID, nil, &empty, nil); err == nil {
		t.Fatalf("un titre vide devrait être refusé")
	}

	title := "Nouveau titre"
	ctx := "Nouveau contexte"
	updated, err := s.UpdateCard(c.ID, nil, &title, &ctx)
	if err != nil {
		t.Fatalf("UpdateCard: %v", err)
	}
	if updated.Title != title || updated.ContextPrompt != ctx {
		t.Fatalf("mise à jour inattendue : %+v", updated)
	}
	if updated.Column != "soon" {
		t.Fatalf("la colonne ne devrait pas changer quand column=nil, reçue %q", updated.Column)
	}

	// column peut toujours être modifiée seule (déplacement manuel).
	col := "doing"
	updated, err = s.UpdateCard(c.ID, &col, nil, nil)
	if err != nil {
		t.Fatalf("UpdateCard (column): %v", err)
	}
	if updated.Column != "doing" || updated.Title != title {
		t.Fatalf("mise à jour inattendue : %+v", updated)
	}

	if _, err := s.UpdateCard("inconnue", &col, nil, nil); err == nil {
		t.Fatalf("une carte inconnue devrait être refusée")
	}
}

// --- Cycle de vie des tâches : accept/cancel/reopen ---

// mkTaskWithStatus crée une tâche sur la carte donnée puis force son statut
// (sauf "running", statut initial de CreateTask).
func mkTaskWithStatus(t *testing.T, s *Store, cardID, projectID, status string) string {
	t.Helper()
	id, ref := s.ReserveTaskID()
	if _, err := s.CreateTask(id, ref, cardID, projectID, "T "+status+" "+id, "echo", "sillage/"+id, "main", "/tmp/wt-"+id, "p"); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if status != "running" {
		if _, err := s.UpdateTask(id, func(tk *Task) { tk.Status = status }); err != nil {
			t.Fatalf("UpdateTask: %v", err)
		}
	}
	return id
}

func TestTaskAcceptTransitions(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	id := mkTaskWithStatus(t, s, card.ID, project.ID, "review")
	task, err := s.AcceptTask(id)
	if err != nil {
		t.Fatalf("AcceptTask depuis review: %v", err)
	}
	if task.Status != "accepted" {
		t.Fatalf("statut attendu 'accepted', reçu %q", task.Status)
	}

	running := mkTaskWithStatus(t, s, card.ID, project.ID, "running")
	if _, err := s.AcceptTask(running); err == nil {
		t.Fatalf("accepter depuis running devrait être refusé")
	} else if err.Error() != "interrupt the agent before accepting" {
		t.Fatalf("message inattendu : %q", err.Error())
	}

	for _, status := range []string{"accepted", "cancelled"} {
		id := mkTaskWithStatus(t, s, card.ID, project.ID, status)
		if _, err := s.AcceptTask(id); err == nil {
			t.Fatalf("accepter depuis %s devrait être refusé", status)
		}
	}

	if _, err := s.AcceptTask("inconnu"); err == nil {
		t.Fatalf("accepter une tâche inconnue devrait échouer")
	}
}

func TestTaskCancelTransitions(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	for _, status := range []string{"running", "review"} {
		id := mkTaskWithStatus(t, s, card.ID, project.ID, status)
		task, err := s.CancelTask(id)
		if err != nil {
			t.Fatalf("CancelTask depuis %s: %v", status, err)
		}
		if task.Status != "cancelled" {
			t.Fatalf("statut attendu 'cancelled', reçu %q", task.Status)
		}
	}

	for _, status := range []string{"accepted", "cancelled"} {
		id := mkTaskWithStatus(t, s, card.ID, project.ID, status)
		if _, err := s.CancelTask(id); err == nil {
			t.Fatalf("cancel depuis %s devrait être refusé", status)
		}
	}
}

func TestTaskReopenTransitions(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	for _, status := range []string{"accepted", "cancelled"} {
		id := mkTaskWithStatus(t, s, card.ID, project.ID, status)
		task, err := s.ReopenTask(id)
		if err != nil {
			t.Fatalf("ReopenTask depuis %s: %v", status, err)
		}
		if task.Status != "review" {
			t.Fatalf("statut attendu 'review', reçu %q", task.Status)
		}
	}

	for _, status := range []string{"running", "review"} {
		id := mkTaskWithStatus(t, s, card.ID, project.ID, status)
		if _, err := s.ReopenTask(id); err == nil {
			t.Fatalf("reopen depuis %s devrait être refusé", status)
		}
	}
}

// --- Auto-déplacement de carte : "done" veut dire livré ---

func TestCardAutoMoveToDoneAndBack(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	id1, ref1 := s.ReserveTaskID()
	if _, err := s.CreateTask(id1, ref1, card.ID, project.ID, "T1", "echo", "sillage/"+id1, "main", "/tmp/wt1", "p"); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	id2, ref2 := s.ReserveTaskID()
	if _, err := s.CreateTask(id2, ref2, card.ID, project.ID, "T2", "echo", "sillage/"+id2, "main", "/tmp/wt2", "p"); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	// Une branche de chantier existe (première tâche créée sur le dépôt), mais
	// rien n'est encore livré.
	if _, err := s.SetCardBranch(card.ID, CardBranch{RepoName: "p", Branch: "sillage/ws-1-carte", Base: "main", WorktreeDir: "/tmp/ws"}); err != nil {
		t.Fatalf("SetCardBranch: %v", err)
	}

	// Dès qu'une tâche active existe, la carte quitte "soon" pour "doing".
	if c, _ := s.GetCard(card.ID); c.Column != "doing" {
		t.Fatalf("colonne attendue 'doing' dès qu'une tâche est active, reçue %q", c.Column)
	}

	// T1 est acceptée (running -> review -> accepted).
	if _, err := s.UpdateTask(id1, func(tk *Task) { tk.Status = "review" }); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if _, err := s.AcceptTask(id1); err != nil {
		t.Fatalf("AcceptTask: %v", err)
	}
	if c, _ := s.GetCard(card.ID); c.Column != "doing" {
		t.Fatalf("colonne attendue 'doing' tant que T2 est active, reçue %q", c.Column)
	}

	// T2 est refusée : les deux tâches sont terminales, mais rien n'est livré,
	// donc la carte reste en "doing" (la colonne "Terminé" veut dire livré).
	if _, err := s.CancelTask(id2); err != nil {
		t.Fatalf("CancelTask: %v", err)
	}
	c, _ := s.GetCard(card.ID)
	if c.Column != "doing" {
		t.Fatalf("colonne attendue 'doing' avant livraison, reçue %q", c.Column)
	}
	if !c.ShipReady {
		t.Fatalf("le chantier devrait être livrable (blocage %q)", c.ShipBlocker)
	}
	if !c.AwaitingShip {
		t.Fatalf("le chantier devrait être signalé comme non livré (tout est terminal)")
	}

	// Livraison : la carte passe en "done" et n'est plus signalée en attente.
	if _, err := s.MarkCardBranchShipped(card.ID, "p", "https://example.com/pr/1", time.Now().UTC()); err != nil {
		t.Fatalf("MarkCardBranchShipped: %v", err)
	}
	c, _ = s.GetCard(card.ID)
	if c.Column != "done" {
		t.Fatalf("colonne attendue 'done' après livraison, reçue %q", c.Column)
	}
	if c.AwaitingShip {
		t.Fatalf("le chantier livré ne devrait plus être signalé comme non livré")
	}

	// Réouvrir T1 : la carte doit repasser en doing.
	if _, err := s.ReopenTask(id1); err != nil {
		t.Fatalf("ReopenTask: %v", err)
	}
	c, _ = s.GetCard(card.ID)
	if c.Column != "doing" {
		t.Fatalf("colonne attendue 'doing' après réouverture, reçue %q", c.Column)
	}

	// Accepter à nouveau T1 : la carte a déjà été livrée, retour en done.
	if _, err := s.AcceptTask(id1); err != nil {
		t.Fatalf("AcceptTask: %v", err)
	}
	c, _ = s.GetCard(card.ID)
	if c.Column != "done" {
		t.Fatalf("colonne attendue 'done', reçue %q", c.Column)
	}

	// Une nouvelle tâche sur une carte "done" la fait revenir en "doing".
	id3, ref3 := s.ReserveTaskID()
	if _, err := s.CreateTask(id3, ref3, card.ID, project.ID, "T3", "echo", "sillage/"+id3, "main", "/tmp/wt3", "p"); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	c, _ = s.GetCard(card.ID)
	if c.Column != "doing" {
		t.Fatalf("colonne attendue 'doing' après nouvelle tâche sur carte done, reçue %q", c.Column)
	}
}

// --- AwaitingShip : un chantier entièrement refusé n'a rien à livrer ---

func TestCardAwaitingShipFalseWhenNothingAccepted(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	if _, err := s.SetCardBranch(card.ID, CardBranch{RepoName: "p", Branch: "sillage/ws-1-carte", Base: "main", WorktreeDir: "/tmp/ws"}); err != nil {
		t.Fatalf("SetCardBranch: %v", err)
	}
	mkTaskWithStatus(t, s, card.ID, project.ID, "cancelled")

	c, ok := s.GetCard(card.ID)
	if !ok {
		t.Fatalf("carte introuvable")
	}
	if c.AwaitingShip {
		t.Fatalf("un chantier entièrement refusé n'a rien à livrer, ne devrait pas être signalé")
	}
}

// --- Compteurs de carte avec des tâches cancelled (v0.3.1 section 4) ---

func TestCardCountersExcludeCancelled(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	for _, status := range []string{"accepted", "accepted", "cancelled", "review"} {
		mkTaskWithStatus(t, s, card.ID, project.ID, status)
	}

	c, ok := s.GetCard(card.ID)
	if !ok {
		t.Fatalf("carte introuvable")
	}
	// 3 tâches comptent (deux acceptées, une en revue) ; cancelled est exclue.
	if c.TasksTotal != 3 {
		t.Fatalf("tasksTotal attendu 3 (cancelled exclue), reçu %d", c.TasksTotal)
	}
	if c.TasksDone != 2 {
		t.Fatalf("tasksDone attendu 2 (acceptées), reçu %d", c.TasksDone)
	}
	if c.ReviewCount != 1 {
		t.Fatalf("reviewCount attendu 1, reçu %d", c.ReviewCount)
	}
	wantProgress := 2 * 100 / 3
	if c.Progress != wantProgress {
		t.Fatalf("progress attendu %d, reçu %d", wantProgress, c.Progress)
	}
	// La tâche "review" n'est pas terminale : la carte ne doit pas être en
	// "done". Les tâches ont été ajoutées une à une (passant chacune par
	// "running"), ce qui a fait transiter la carte par "done" avant de
	// revenir en "doing" (jamais "soon", conformément à la règle de retour).
	if c.Column != "doing" {
		t.Fatalf("colonne attendue 'doing' tant qu'une tâche est en review, reçue %q", c.Column)
	}
}

// --- Réglages globaux (v0.3.1 section 6) ---

func TestSettingsUpdateValidatesLang(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	invalid := "de"
	if _, err := s.UpdateSettings(nil, &invalid, nil); err == nil {
		t.Fatalf("une langue invalide devrait être refusée")
	}

	name := "Ada"
	lang := "fr"
	settings, err := s.UpdateSettings(&name, &lang, nil)
	if err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if settings.DisplayName != "Ada" || settings.Lang != "fr" {
		t.Fatalf("settings inattendus : %+v", settings)
	}

	empty := ""
	settings, err = s.UpdateSettings(nil, &empty, nil)
	if err != nil {
		t.Fatalf("UpdateSettings (lang vide): %v", err)
	}
	if settings.Lang != "" || settings.DisplayName != "Ada" {
		t.Fatalf("settings inattendus après reset lang : %+v", settings)
	}

	// La valeur persiste après rechargement.
	s2, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore (reload): %v", err)
	}
	if got := s2.GetSettings(); got.DisplayName != "Ada" {
		t.Fatalf("displayName attendu 'Ada' après rechargement, reçu %q", got.DisplayName)
	}
}

// --- Réassignation d'agent (v0.3.2 section 1) ---

func TestReassignTask(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	id, ref := s.ReserveTaskID()
	if _, err := s.CreateTask(id, ref, card.ID, project.ID, "T", "echo", "sillage/"+id, "main", "/tmp/wt", "p"); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	// Refusé pendant l'exécution.
	if _, err := s.ReassignTask(id, "bolt"); err == nil {
		t.Fatalf("la réassignation d'une tâche running devrait être refusée")
	} else if err.Error() != "interrupt the agent before reassigning" {
		t.Fatalf("message d'erreur inattendu : %q", err.Error())
	}

	if _, err := s.UpdateTask(id, func(tk *Task) {
		tk.Status = "review"
		tk.SessionID = "sess-abc"
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	// Agent inconnu refusé.
	if _, err := s.ReassignTask(id, "inconnu"); err == nil {
		t.Fatalf("la réassignation vers un agent inconnu devrait être refusée")
	}

	task, err := s.ReassignTask(id, "bolt")
	if err != nil {
		t.Fatalf("ReassignTask: %v", err)
	}
	if task.AgentID != "bolt" {
		t.Fatalf("agentId attendu 'bolt', reçu %q", task.AgentID)
	}
	if task.SessionID != "" {
		t.Fatalf("sessionId devrait être vidé après réassignation, reçu %q", task.SessionID)
	}
}

// --- Avertissements de santé des agents (v0.3.2 section 2) ---

func TestAgentWarningCodexSandboxBlockedByAppArmor(t *testing.T) {
	origPath := apparmorRestrictPath
	defer func() { apparmorRestrictPath = origPath }()
	apparmorRestrictPath = filepath.Join(t.TempDir(), "apparmor_restrict_unprivileged_userns")
	if err := os.WriteFile(apparmorRestrictPath, []byte("1\n"), 0o644); err != nil {
		t.Fatalf("écriture fichier test : %v", err)
	}

	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()

	t.Setenv("SILLAGE_CODEX_SANDBOX", "")
	lookPath = func(string) (string, error) { return "", fmt.Errorf("not found") }
	if got := agentWarning(Agent{Cli: "codex"}); got != "codex CLI not found in PATH" {
		t.Fatalf("a missing CLI should take priority over sandbox diagnostics, got %q", got)
	}

	lookPath = func(string) (string, error) { return "/usr/bin/codex", nil }
	want := "codex sandbox is blocked on this machine (AppArmor); see README (SILLAGE_CODEX_SANDBOX)"
	if got := agentWarning(Agent{Cli: "codex"}); got != want {
		t.Fatalf("warning attendu %q, reçu %q", want, got)
	}

	t.Setenv("SILLAGE_CODEX_SANDBOX", "danger-full-access")
	if got := agentWarning(Agent{Cli: "codex"}); got == want {
		t.Fatalf("SILLAGE_CODEX_SANDBOX définie ne devrait plus déclencher l'avertissement AppArmor")
	}
}

func TestAgentWarningMissingBinary(t *testing.T) {
	origPath := apparmorRestrictPath
	defer func() { apparmorRestrictPath = origPath }()
	apparmorRestrictPath = filepath.Join(t.TempDir(), "absent") // fichier absent -> pas de blocage AppArmor.

	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()
	lookPath = func(string) (string, error) { return "", fmt.Errorf("not found") }

	if got := agentWarning(Agent{Cli: "claude"}); got != "claude CLI not found in PATH" {
		t.Fatalf("warning attendu 'claude CLI not found in PATH', reçu %q", got)
	}
	if got := agentWarning(Agent{Cli: "codex"}); got != "codex CLI not found in PATH" {
		t.Fatalf("warning attendu 'codex CLI not found in PATH', reçu %q", got)
	}
	if got := agentWarning(Agent{Cli: "copilot"}); got != "copilot CLI not found in PATH" {
		t.Fatalf("warning = %q, want missing Copilot CLI", got)
	}
	if got := agentWarning(Agent{Cli: "agy"}); got != "agy CLI not found in PATH" {
		t.Fatalf("warning = %q, want missing Antigravity CLI", got)
	}
	if got := agentWarning(Agent{Cli: "fake"}); got != "" {
		t.Fatalf("l'agent fake ne devrait jamais avoir d'avertissement, reçu %q", got)
	}
}

func TestAgentWarningAntigravityToolPermission(t *testing.T) {
	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()
	lookPath = func(file string) (string, error) { return "/usr/bin/" + file, nil }

	origSettings := antigravitySettingsPath
	defer func() { antigravitySettingsPath = origSettings }()
	dir := t.TempDir()
	antigravitySettingsPath = filepath.Join(dir, "settings.json")

	// Fichier absent : la CLI applique sa politique par défaut, qui auto-refuse
	// les commandes en mode headless.
	if got := agentWarning(Agent{Cli: "agy"}); got != antigravityPolicyWarning {
		t.Fatalf("warning = %q, want the tool permission warning", got)
	}

	cases := map[string]string{
		`{"toolPermission":"request-review"}`:       antigravityPolicyWarning,
		`{"toolPermission":"strict"}`:               antigravityPolicyWarning,
		`{"permissions":{"allow":["file(*)"]}}`:     antigravityPolicyWarning,
		`{"toolPermission":"proceed-in-sandbox"}`:   "",
		`{"toolPermission":"always-proceed"}`:       "",
		`{"permissions":{"allow":["command(go)"]}}`: "",
	}
	for settings, want := range cases {
		if err := os.WriteFile(antigravitySettingsPath, []byte(settings), 0o644); err != nil {
			t.Fatalf("écriture settings : %v", err)
		}
		if got := agentWarning(Agent{Cli: "agy"}); got != want {
			t.Fatalf("settings %s: warning = %q, want %q", settings, got, want)
		}
	}
}

func TestFixAntigravityToolPermission(t *testing.T) {
	origSettings := antigravitySettingsPath
	defer func() { antigravitySettingsPath = origSettings }()
	dir := t.TempDir()
	antigravitySettingsPath = filepath.Join(dir, "nested", "settings.json")

	// Fichier (et dossier) absents : le correctif les crée.
	if err := fixAntigravityToolPermission(); err != nil {
		t.Fatalf("fixAntigravityToolPermission: %v", err)
	}
	if !antigravityAllowsHeadlessCommands() {
		t.Fatal("the policy should allow headless commands after the fix")
	}

	// Les autres clés de l'utilisateur survivent : c'est son fichier.
	kept := `{"enableTelemetry":false,"trustedWorkspaces":["/home/me/repo"],"toolPermission":"request-review"}`
	if err := os.WriteFile(antigravitySettingsPath, []byte(kept), 0o600); err != nil {
		t.Fatalf("écriture settings : %v", err)
	}
	if err := os.Chmod(antigravitySettingsPath, 0o600); err != nil { // le fichier existait déjà : WriteFile ne change pas ses droits
		t.Fatalf("chmod settings : %v", err)
	}
	if err := fixAntigravityToolPermission(); err != nil {
		t.Fatalf("fixAntigravityToolPermission: %v", err)
	}
	data, err := os.ReadFile(antigravitySettingsPath)
	if err != nil {
		t.Fatalf("relecture settings : %v", err)
	}
	var got struct {
		EnableTelemetry   bool     `json:"enableTelemetry"`
		TrustedWorkspaces []string `json:"trustedWorkspaces"`
		ToolPermission    string   `json:"toolPermission"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("settings illisibles après correctif : %v", err)
	}
	if got.ToolPermission != "proceed-in-sandbox" {
		t.Fatalf("toolPermission = %q, want proceed-in-sandbox", got.ToolPermission)
	}
	if got.EnableTelemetry || len(got.TrustedWorkspaces) != 1 || got.TrustedWorkspaces[0] != "/home/me/repo" {
		t.Fatalf("the other keys should survive untouched, got %s", data)
	}
	if info, err := os.Stat(antigravitySettingsPath); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("file mode should be preserved, got %v (%v)", info.Mode().Perm(), err)
	}

	// Fichier illisible : refuser plutôt que perdre la configuration.
	broken := []byte("{ not json")
	if err := os.WriteFile(antigravitySettingsPath, broken, 0o600); err != nil {
		t.Fatalf("écriture settings : %v", err)
	}
	if err := fixAntigravityToolPermission(); err == nil {
		t.Fatal("an unreadable settings file should be an error, not an overwrite")
	}
	if data, _ := os.ReadFile(antigravitySettingsPath); string(data) != string(broken) {
		t.Fatalf("the unreadable file should be left untouched, got %s", data)
	}
}

func TestAgentWarningHealthy(t *testing.T) {
	origPath := apparmorRestrictPath
	defer func() { apparmorRestrictPath = origPath }()
	apparmorRestrictPath = filepath.Join(t.TempDir(), "absent")

	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()
	lookPath = func(file string) (string, error) { return "/usr/bin/" + file, nil }

	origSettings := antigravitySettingsPath
	defer func() { antigravitySettingsPath = origSettings }()
	antigravitySettingsPath = filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(antigravitySettingsPath, []byte(`{"toolPermission":"proceed-in-sandbox"}`), 0o644); err != nil {
		t.Fatalf("écriture settings : %v", err)
	}

	if got := agentWarning(Agent{Cli: "codex"}); got != "" {
		t.Fatalf("aucun avertissement attendu, reçu %q", got)
	}
	if got := agentWarning(Agent{Cli: "claude"}); got != "" {
		t.Fatalf("aucun avertissement attendu, reçu %q", got)
	}
	if got := agentWarning(Agent{Cli: "copilot"}); got != "" {
		t.Fatalf("healthy Copilot should have no warning, got %q", got)
	}
	if got := agentWarning(Agent{Cli: "agy"}); got != "" {
		t.Fatalf("healthy Antigravity should have no warning, got %q", got)
	}
}

func TestListAgentsExposesWarningNotPersisted(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()
	lookPath = func(string) (string, error) { return "", fmt.Errorf("not found") }
	origPath := apparmorRestrictPath
	defer func() { apparmorRestrictPath = origPath }()
	apparmorRestrictPath = filepath.Join(t.TempDir(), "absent")

	agents := s.ListAgents()
	found := false
	for _, a := range agents {
		if a.ID == "bolt" {
			found = true
			if a.Warning == "" {
				t.Fatalf("l'agent bolt (claude) devrait porter un avertissement quand le binaire est introuvable")
			}
		}
	}
	if !found {
		t.Fatalf("agent seedé 'bolt' introuvable")
	}

	// Le champ warning ne doit jamais être écrit dans state.json.
	raw, err := os.ReadFile(filepath.Join(dir, "state.json"))
	if err != nil {
		t.Fatalf("lecture state.json : %v", err)
	}
	if strings.Contains(string(raw), "warning") {
		t.Fatalf("state.json ne devrait jamais contenir le champ warning : %s", raw)
	}
}

// TestSetCodexQuotaExposedOnlyOnCodexAgents vérifie que le quota codex,
// une fois défini, est attaché à tous les agents cli=codex (quota de compte,
// pas par agent) et à eux seuls, et qu'il survit à un rechargement disque.
func TestSetCodexQuotaExposedOnlyOnCodexAgents(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	windows := []AgentQuotaWindow{
		{Label: "5h", UsedPercent: 42, ResetsAt: time.Unix(1771005432, 0)},
		{Label: "week", UsedPercent: 48, ResetsAt: time.Unix(1771409457, 0)},
	}
	if err := s.SetCodexQuota(windows); err != nil {
		t.Fatalf("SetCodexQuota: %v", err)
	}

	agents := s.ListAgents()
	for _, a := range agents {
		if a.Cli == "codex" {
			if a.Quota == nil || len(a.Quota.Windows) != 2 {
				t.Fatalf("agent codex %q devrait porter le quota, reçu %+v", a.ID, a.Quota)
			}
		} else if a.Quota != nil {
			t.Fatalf("agent %q (cli=%s) ne devrait pas porter de quota, reçu %+v", a.ID, a.Cli, a.Quota)
		}
	}

	// Persisté : un rechargement depuis disque garde l'instantané.
	s2, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore (reload): %v", err)
	}
	reloaded := s2.ListAgents()
	found := false
	for _, a := range reloaded {
		if a.Cli == "codex" {
			found = true
			if a.Quota == nil || len(a.Quota.Windows) != 2 {
				t.Fatalf("quota codex non persisté après rechargement, reçu %+v", a.Quota)
			}
		}
	}
	if !found {
		t.Fatalf("aucun agent codex trouvé après rechargement")
	}
}

// --- Rappel de contexte au départ frais (v0.3.2 section 1) ---

func TestContextualizeCliInput(t *testing.T) {
	if got := contextualizeCliInput("Mon titre", "un message"); got != "Task: Mon titre\n\nun message" {
		t.Fatalf("résultat inattendu : %q", got)
	}
	if got := contextualizeCliInput("Mon titre", ""); got != "Task: Mon titre" {
		t.Fatalf("résultat inattendu (sans texte) : %q", got)
	}
}

func TestCardAutoMovesFromSoonToDoingWhenTaskStarts(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p, _ := s.AddProject("p", "", "", []Repo{{Name: "r", Path: "/tmp/x"}}, nil, nil)
	card, _ := s.AddCard(p.ID, "Carte", "", "")
	if card.Column != "soon" {
		t.Fatalf("colonne initiale attendue soon, reçue %q", card.Column)
	}
	id, ref := s.ReserveTaskID()
	if _, err := s.CreateTask(id, ref, card.ID, p.ID, "T", "echo", "sillage/x", "main", "/tmp/wt", "r"); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	got, _ := s.GetCard(card.ID)
	if got.Column != "doing" {
		t.Fatalf("la carte devrait passer en doing quand une tâche démarre, reçue %q", got.Column)
	}
}

// --- Suppressions (Store, hors orchestration Runner) ---

func TestDeleteTaskRemovesTaskAndMessages(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p, _ := s.AddProject("p", "", "", []Repo{{Name: "r", Path: "/tmp/x"}}, nil, nil)
	card, _ := s.AddCard(p.ID, "Carte", "", "")
	id := mkTaskWithStatus(t, s, card.ID, p.ID, "accepted")
	if _, _, err := s.AddMessage(id, "user", "", "bonjour"); err != nil {
		t.Fatalf("AddMessage: %v", err)
	}

	deleted, err := s.DeleteTask(id)
	if err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	if deleted.ID != id {
		t.Fatalf("la tâche retournée devrait être celle supprimée, reçu %q", deleted.ID)
	}
	if _, ok := s.GetTask(id); ok {
		t.Fatalf("la tâche ne devrait plus exister après suppression")
	}
	if msgs := s.GetMessages(id); len(msgs) != 0 {
		t.Fatalf("les messages de la tâche devraient avoir disparu, reçu %d", len(msgs))
	}
	got, _ := s.GetCard(card.ID)
	if got.TasksTotal != 0 {
		t.Fatalf("le compteur de tâches de la carte devrait être à 0, reçu %d", got.TasksTotal)
	}
}

func TestDeleteTaskRefusesUnknownID(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if _, err := s.DeleteTask("t-inconnu"); err == nil {
		t.Fatalf("la suppression d'une tâche inconnue devrait échouer")
	}
}

func TestCardCascadeDeletesAllTasks(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p, _ := s.AddProject("p", "", "", []Repo{{Name: "r", Path: "/tmp/x"}}, nil, nil)
	card, _ := s.AddCard(p.ID, "Carte", "", "")
	var taskIDs []string
	for i := 0; i < 3; i++ {
		taskIDs = append(taskIDs, mkTaskWithStatus(t, s, card.ID, p.ID, "accepted"))
	}

	runner := NewRunner(s, NewHub())
	if err := runner.DeleteCard(card.ID); err != nil {
		t.Fatalf("DeleteCard: %v", err)
	}

	for _, id := range taskIDs {
		if _, ok := s.GetTask(id); ok {
			t.Fatalf("la tâche %s devrait avoir été supprimée par la cascade de la carte", id)
		}
	}
	if _, ok := s.GetCard(card.ID); ok {
		t.Fatalf("la carte devrait avoir été supprimée")
	}
}

func TestProjectCascadeDeletesCardsAndTasks(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p, _ := s.AddProject("p", "", "", []Repo{{Name: "r", Path: "/tmp/x"}}, nil, nil)
	card1, _ := s.AddCard(p.ID, "Carte 1", "", "")
	card2, _ := s.AddCard(p.ID, "Carte 2", "", "")
	t1 := mkTaskWithStatus(t, s, card1.ID, p.ID, "accepted")
	t2 := mkTaskWithStatus(t, s, card2.ID, p.ID, "review")

	runner := NewRunner(s, NewHub())
	if err := runner.DeleteProject(p.ID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	if _, ok := s.GetProject(p.ID); ok {
		t.Fatalf("le projet devrait avoir été supprimé")
	}
	if _, ok := s.GetCard(card1.ID); ok {
		t.Fatalf("la carte 1 devrait avoir été supprimée")
	}
	if _, ok := s.GetCard(card2.ID); ok {
		t.Fatalf("la carte 2 devrait avoir été supprimée")
	}
	if _, ok := s.GetTask(t1); ok {
		t.Fatalf("la tâche %s devrait avoir été supprimée par la cascade du projet", t1)
	}
	if _, ok := s.GetTask(t2); ok {
		t.Fatalf("la tâche %s devrait avoir été supprimée par la cascade du projet", t2)
	}
}

func TestNormalizeAllowedTools(t *testing.T) {
	got := NormalizeAllowedTools([]string{"  Bash(pytest:*)  ", "", "   ", "Bash(ruff:*)"})
	if len(got) != 2 || got[0] != "Bash(pytest:*)" || got[1] != "Bash(ruff:*)" {
		t.Fatalf("liste nettoyée attendue [Bash(pytest:*) Bash(ruff:*)], reçue %v", got)
	}
	// Non nulle même vide : nil est réservé aux projets antérieurs au champ,
	// que migrateProjectAllowedTools amorce au chargement.
	if empty := NormalizeAllowedTools(nil); empty == nil {
		t.Fatalf("une liste vide doit rester non nulle pour ne pas être prise pour un projet à migrer")
	}
}

// Un projet antérieur au champ récupère la chaîne Go qui vivait dans le socle,
// pour garder exactement son comportement d'avant la mise à jour. Un projet qui
// a explicitement vidé sa liste n'est pas réamorcé à chaque démarrage.
func TestMigrateProjectAllowedTools(t *testing.T) {
	s := &Store{
		Projects: map[string]Project{
			"p1": {ID: "p1"},                              // antérieur au champ
			"p2": {ID: "p2", AllowedTools: []string{}},    // vidé volontairement
			"p3": {ID: "p3", AllowedTools: []string{"X"}}, // déjà réglé
		},
	}
	migrateProjectAllowedTools(s)

	if got := s.Projects["p1"].AllowedTools; len(got) != len(legacyGoAllowedTools) {
		t.Fatalf("projet antérieur : chaîne Go attendue %v, reçue %v", legacyGoAllowedTools, got)
	}
	if got := s.Projects["p2"].AllowedTools; len(got) != 0 {
		t.Fatalf("une liste vidée volontairement ne doit pas être réamorcée, reçue %v", got)
	}
	if got := s.Projects["p3"].AllowedTools; len(got) != 1 || got[0] != "X" {
		t.Fatalf("une liste déjà réglée ne doit pas bouger, reçue %v", got)
	}

	// Idempotence : un second chargement ne doit rien réécrire.
	migrateProjectAllowedTools(s)
	if got := s.Projects["p1"].AllowedTools; len(got) != len(legacyGoAllowedTools) {
		t.Fatalf("migration non idempotente, reçue %v", got)
	}
}

func TestUpdateProjectAllowedTools(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p, err := s.AddProject("demo", "", "", []Repo{{Name: "demo", Path: "/tmp/demo"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if p.AllowedTools == nil {
		t.Fatalf("un projet neuf doit avoir une liste vide non nulle, pas nil")
	}

	tools := []string{"Bash(pytest:*)", "  "}
	updated, err := s.UpdateProject(p.ID, nil, nil, nil, nil, &tools, nil, nil, nil)
	if err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if len(updated.AllowedTools) != 1 || updated.AllowedTools[0] != "Bash(pytest:*)" {
		t.Fatalf("outils attendus [Bash(pytest:*)], reçus %v", updated.AllowedTools)
	}

	// nil (champ non fourni) ne touche pas à la liste existante.
	updated, err = s.UpdateProject(p.ID, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("UpdateProject (sans allowedTools): %v", err)
	}
	if len(updated.AllowedTools) != 1 {
		t.Fatalf("la liste ne devait pas changer, reçue %v", updated.AllowedTools)
	}

	// Liste vide fournie : l'utilisateur retire tout, et ça se persiste.
	none := []string{}
	updated, err = s.UpdateProject(p.ID, nil, nil, nil, nil, &none, nil, nil, nil)
	if err != nil {
		t.Fatalf("UpdateProject (liste vide): %v", err)
	}
	if len(updated.AllowedTools) != 0 {
		t.Fatalf("liste vide attendue, reçue %v", updated.AllowedTools)
	}
}

// TestStateFormatVersionStamped : toute sauvegarde signe le fichier (format +
// version qui l'a écrit), y compris sur un state.json antérieur au champ.
func TestStateFormatVersionStamped(t *testing.T) {
	dir := t.TempDir()
	legacy := `{"Projects":{},"Cards":{},"Tasks":{},"Messages":{},"Agents":{}}`
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}

	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore sur un fichier legacy: %v", err)
	}
	if s.FormatVersion != stateFormatVersion {
		t.Fatalf("format attendu %d, reçu %d", stateFormatVersion, s.FormatVersion)
	}
	if s.WrittenBy != buildVersion {
		t.Fatalf("writtenBy attendu %q, reçu %q", buildVersion, s.WrittenBy)
	}

	got, err := StateFileFormatVersion(filepath.Join(dir, "state.json"))
	if err != nil {
		t.Fatalf("StateFileFormatVersion: %v", err)
	}
	if got != stateFormatVersion {
		t.Fatalf("format sur disque attendu %d, reçu %d", stateFormatVersion, got)
	}
}

// TestStateTooNewRefusesToLoad est le garde-fou contre un binaire plus ancien :
// un fichier d'un format plus récent ne doit ni se charger ni, surtout, être
// réécrit (c'est la réécriture qui détruit les champs inconnus).
func TestStateTooNewRefusesToLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	future := fmt.Sprintf(`{"FormatVersion":%d,"WrittenBy":"99.0.0","Projects":{},"Cards":{},"Tasks":{},"Messages":{},"Agents":{},"UnknownFutureField":{"keep":"me"}}`, stateFormatVersion+1)
	if err := os.WriteFile(path, []byte(future), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := NewStore(dir); err == nil {
		t.Fatalf("NewStore devait refuser un state.json plus récent")
	} else if !errors.Is(err, ErrStateTooNew) {
		t.Fatalf("erreur attendue ErrStateTooNew, reçue %v", err)
	} else if !strings.Contains(err.Error(), "upgrade Sillage") {
		t.Fatalf("le message doit dire quoi faire, reçu %q", err.Error())
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("le fichier a été réécrit alors qu'il devait être laissé intact")
	}
}

// TestDowngradeWarning : ce que le format ne voit pas. Deux versions publiées
// du même format, dont celle qui tourne est la plus ancienne : le chargement
// réussit, mais on le dit. Une compilation locale ne se compare à rien.
func TestDowngradeWarning(t *testing.T) {
	cases := []struct {
		writtenBy, running string
		want               bool
	}{
		{"0.10.0", "0.4.2", true},
		{"0.4.2", "0.10.0", false}, // mise à jour normale : rien à signaler
		{"0.10.0", "0.10.0", false},
		{"dev", "0.10.0", false}, // écrit par un build local : incomparable
		{"0.10.0", "dev", false}, // c'est le build local qui tourne : incomparable
		{"", "0.10.0", false},    // fichier antérieur au champ
	}
	for _, c := range cases {
		dir := t.TempDir()
		s, err := NewStore(dir)
		if err != nil {
			t.Fatalf("NewStore: %v", err)
		}
		s.previousWriter = c.writtenBy
		prev := buildVersion
		buildVersion = c.running
		got := s.DowngradeWarning()
		buildVersion = prev
		if (got != "") != c.want {
			t.Errorf("écrit par %q, tourne en %q : avertissement=%q, attendu %v", c.writtenBy, c.running, got, c.want)
		}
	}
}
