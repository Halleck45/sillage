package server

import (
	"testing"
)

func TestStoreRoundtripSaveLoad(t *testing.T) {
	dir := t.TempDir()

	s1, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	project, err := s1.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}})
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s1.AddCard(project.ID, "Ma carte", "")
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

	project, err := s.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}})
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "")
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
	emptyCard, err := s.AddCard(project.ID, "Vide", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	if emptyCard.Progress != 0 || emptyCard.TasksTotal != 0 {
		t.Fatalf("carte vide inattendue : %+v", emptyCard)
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
	project, err := s.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}})
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "")
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
	p, err := s.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}})
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}

	name := "Nouveau nom"
	checkCmd := "go test ./..."
	updated, err := s.UpdateProject(p.ID, &name, nil, &checkCmd, nil, nil)
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
	updated, err = s.UpdateProject(p.ID, nil, nil, nil, nil, &newRepos)
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
	p, err := s.AddProject("sillage", "Un projet", "Contexte initial", []Repo{{Path: "/tmp/sillage"}})
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if p.Description != "Un projet" || p.ContextPrompt != "Contexte initial" {
		t.Fatalf("description/contextPrompt inattendus à la création : %+v", p)
	}

	desc := "Nouvelle description"
	ctx := "Nouveau contexte"
	updated, err := s.UpdateProject(p.ID, nil, &desc, nil, &ctx, nil)
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
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}})
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}

	if _, err := s.AddCard(project.ID, "Carte", "doing"); err == nil {
		t.Fatalf("une carte créée en doing devrait être refusée")
	} else if err.Error() != "cards are created in the soon column" {
		t.Fatalf("message d'erreur inattendu : %q", err.Error())
	}
	if _, err := s.AddCard(project.ID, "Carte", "done"); err == nil {
		t.Fatalf("une carte créée en done devrait être refusée")
	}

	c, err := s.AddCard(project.ID, "Carte", "")
	if err != nil {
		t.Fatalf("AddCard (vide): %v", err)
	}
	if c.Column != "soon" {
		t.Fatalf("colonne attendue 'soon', reçue %q", c.Column)
	}

	c2, err := s.AddCard(project.ID, "Carte2", "soon")
	if err != nil {
		t.Fatalf("AddCard (soon explicite): %v", err)
	}
	if c2.Column != "soon" {
		t.Fatalf("colonne attendue 'soon', reçue %q", c2.Column)
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
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}})
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	for _, status := range []string{"review", "ready", "shipped"} {
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
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}})
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	for _, status := range []string{"running", "review", "ready"} {
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
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}})
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "")
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

	for _, status := range []string{"running", "review", "ready"} {
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
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}})
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "")
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

	if c, _ := s.GetCard(card.ID); c.Column != "soon" {
		t.Fatalf("colonne attendue 'soon' tant que des tâches sont actives, reçue %q", c.Column)
	}

	// T1 termine (running -> review -> done).
	if _, err := s.UpdateTask(id1, func(tk *Task) { tk.Status = "review" }); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if _, err := s.FinishTask(id1); err != nil {
		t.Fatalf("FinishTask: %v", err)
	}
	if c, _ := s.GetCard(card.ID); c.Column != "soon" {
		t.Fatalf("colonne attendue 'soon' tant que T2 est active, reçue %q", c.Column)
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
	project, err := s.AddProject("p", "", "", []Repo{{Path: "/tmp/p"}})
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "")
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
