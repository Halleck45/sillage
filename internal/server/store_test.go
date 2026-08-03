package server

import (
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

	project, err := s1.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}}, nil)
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
		tk.Status = "shipped"
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
	if loadedTask.Status != "shipped" {
		t.Fatalf("statut attendu 'shipped' après rechargement, reçu %q", loadedTask.Status)
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
}

func TestDerivedCounters(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	project, err := s.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}}, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	statuses := []string{"shipped", "review", "running"}
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
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil)
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
	project, err := s.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}}, nil)
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
}

func TestUpdateProjectFields(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p, err := s.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}}, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}

	name := "Nouveau nom"
	checkCmd := "go test ./..."
	updated, err := s.UpdateProject(p.ID, &name, nil, &checkCmd, nil, nil, nil)
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
	updated, err = s.UpdateProject(p.ID, nil, nil, nil, nil, &newRepos, nil)
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
	p, err := s.AddProject("sillage", "Un projet", "Contexte initial", []Repo{{Path: "/tmp/sillage"}}, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if p.Description != "Un projet" || p.ContextPrompt != "Contexte initial" {
		t.Fatalf("description/contextPrompt inattendus à la création : %+v", p)
	}

	desc := "Nouvelle description"
	ctx := "Nouveau contexte"
	updated, err := s.UpdateProject(p.ID, nil, &desc, nil, &ctx, nil, nil)
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
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil)
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
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil)
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
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil)
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

// --- Cycle de vie des tâches : finish/cancel/reopen (v0.3.1 section 4) ---

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

func TestTaskFinishTransitions(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	for _, status := range []string{"review", "shipped"} {
		id := mkTaskWithStatus(t, s, card.ID, project.ID, status)
		task, err := s.FinishTask(id)
		if err != nil {
			t.Fatalf("FinishTask depuis %s: %v", status, err)
		}
		if task.Status != "done" {
			t.Fatalf("statut attendu 'done', reçu %q", task.Status)
		}
	}

	running := mkTaskWithStatus(t, s, card.ID, project.ID, "running")
	if _, err := s.FinishTask(running); err == nil {
		t.Fatalf("finish depuis running devrait être refusé")
	} else if err.Error() != "task must be reviewed before finishing" {
		t.Fatalf("message inattendu : %q", err.Error())
	}

	for _, status := range []string{"done", "cancelled"} {
		id := mkTaskWithStatus(t, s, card.ID, project.ID, status)
		if _, err := s.FinishTask(id); err == nil {
			t.Fatalf("finish depuis %s devrait être refusé", status)
		}
	}

	if _, err := s.FinishTask("inconnu"); err == nil {
		t.Fatalf("finish d'une tâche inconnue devrait échouer")
	}
}

func TestTaskCancelTransitions(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil)
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

	for _, status := range []string{"shipped", "done", "cancelled"} {
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
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	for _, status := range []string{"shipped", "done", "cancelled"} {
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

// --- Auto-déplacement de carte (v0.3.1 section 4) ---

func TestCardAutoMoveToDoneAndBack(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil)
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

	// Dès qu'une tâche active existe, la carte quitte "soon" pour "doing".
	if c, _ := s.GetCard(card.ID); c.Column != "doing" {
		t.Fatalf("colonne attendue 'doing' dès qu'une tâche est active, reçue %q", c.Column)
	}

	// T1 termine (running -> review -> done).
	if _, err := s.UpdateTask(id1, func(tk *Task) { tk.Status = "review" }); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if _, err := s.FinishTask(id1); err != nil {
		t.Fatalf("FinishTask: %v", err)
	}
	if c, _ := s.GetCard(card.ID); c.Column != "doing" {
		t.Fatalf("colonne attendue 'doing' tant que T2 est active, reçue %q", c.Column)
	}

	// T2 est annulée : les deux tâches sont maintenant terminales.
	if _, err := s.CancelTask(id2); err != nil {
		t.Fatalf("CancelTask: %v", err)
	}
	c, _ := s.GetCard(card.ID)
	if c.Column != "done" {
		t.Fatalf("colonne attendue 'done' quand toutes les tâches sont terminales, reçue %q", c.Column)
	}

	// Réouvrir T1 : la carte doit repasser en doing.
	if _, err := s.ReopenTask(id1); err != nil {
		t.Fatalf("ReopenTask: %v", err)
	}
	c, _ = s.GetCard(card.ID)
	if c.Column != "doing" {
		t.Fatalf("colonne attendue 'doing' après réouverture, reçue %q", c.Column)
	}

	// Terminer à nouveau T1 : retour en done.
	if _, err := s.FinishTask(id1); err != nil {
		t.Fatalf("FinishTask: %v", err)
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

// --- Compteurs de carte avec des tâches cancelled (v0.3.1 section 4) ---

func TestCardCountersExcludeCancelled(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	for _, status := range []string{"shipped", "done", "cancelled", "review"} {
		mkTaskWithStatus(t, s, card.ID, project.ID, status)
	}

	c, ok := s.GetCard(card.ID)
	if !ok {
		t.Fatalf("carte introuvable")
	}
	// 3 tâches comptent (shipped, done, review) ; cancelled est exclue.
	if c.TasksTotal != 3 {
		t.Fatalf("tasksTotal attendu 3 (cancelled exclue), reçu %d", c.TasksTotal)
	}
	if c.TasksDone != 2 {
		t.Fatalf("tasksDone attendu 2 (shipped+done), reçu %d", c.TasksDone)
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
	if _, err := s.UpdateSettings(nil, &invalid); err == nil {
		t.Fatalf("une langue invalide devrait être refusée")
	}

	name := "Ada"
	lang := "fr"
	settings, err := s.UpdateSettings(&name, &lang)
	if err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if settings.DisplayName != "Ada" || settings.Lang != "fr" {
		t.Fatalf("settings inattendus : %+v", settings)
	}

	empty := ""
	settings, err = s.UpdateSettings(nil, &empty)
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
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}}, nil)
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
	lookPath = func(string) (string, error) { return "/usr/bin/codex", nil } // binaire présent : seul AppArmor doit déclencher.

	t.Setenv("SILLAGE_CODEX_SANDBOX", "")
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
	if got := agentWarning(Agent{Cli: "fake"}); got != "" {
		t.Fatalf("l'agent fake ne devrait jamais avoir d'avertissement, reçu %q", got)
	}
}

func TestAgentWarningHealthy(t *testing.T) {
	origPath := apparmorRestrictPath
	defer func() { apparmorRestrictPath = origPath }()
	apparmorRestrictPath = filepath.Join(t.TempDir(), "absent")

	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()
	lookPath = func(file string) (string, error) { return "/usr/bin/" + file, nil }

	if got := agentWarning(Agent{Cli: "codex"}); got != "" {
		t.Fatalf("aucun avertissement attendu, reçu %q", got)
	}
	if got := agentWarning(Agent{Cli: "claude"}); got != "" {
		t.Fatalf("aucun avertissement attendu, reçu %q", got)
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
	p, _ := s.AddProject("p", "", "", []Repo{{Name: "r", Path: "/tmp/x"}}, nil)
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
	p, _ := s.AddProject("p", "", "", []Repo{{Name: "r", Path: "/tmp/x"}}, nil)
	card, _ := s.AddCard(p.ID, "Carte", "", "")
	id := mkTaskWithStatus(t, s, card.ID, p.ID, "done")
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
	p, _ := s.AddProject("p", "", "", []Repo{{Name: "r", Path: "/tmp/x"}}, nil)
	card, _ := s.AddCard(p.ID, "Carte", "", "")
	var taskIDs []string
	for i := 0; i < 3; i++ {
		taskIDs = append(taskIDs, mkTaskWithStatus(t, s, card.ID, p.ID, "done"))
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
	p, _ := s.AddProject("p", "", "", []Repo{{Name: "r", Path: "/tmp/x"}}, nil)
	card1, _ := s.AddCard(p.ID, "Carte 1", "", "")
	card2, _ := s.AddCard(p.ID, "Carte 2", "", "")
	t1 := mkTaskWithStatus(t, s, card1.ID, p.ID, "done")
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
