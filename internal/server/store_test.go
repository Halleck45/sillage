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

	project, err := s1.AddProject("atelier", "/tmp/atelier")
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
	task, err := s1.CreateTask(taskID, ref, card.ID, project.ID, "Titre de tâche", "echo", "atelier/100-titre", "main", "/tmp/worktree")
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
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

	msg, task, err := s1.AddMessage(taskID, "user", "bonjour")
	if err != nil {
		t.Fatalf("AddMessage: %v", err)
	}
	if task.MessagesCount != 1 {
		t.Fatalf("messagesCount attendu 1, reçu %d", task.MessagesCount)
	}
	if msg.Author != "user" || msg.Text != "bonjour" {
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
	if loadedProject.Path != "/tmp/atelier" {
		t.Fatalf("chemin de projet inattendu après rechargement : %q", loadedProject.Path)
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

	project, err := s.AddProject("atelier", "/tmp/atelier")
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "doing")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}

	statuses := []string{"shipped", "review", "running"}
	var taskIDs []string
	for _, st := range statuses {
		id, ref := s.ReserveTaskID()
		task, err := s.CreateTask(id, ref, card.ID, project.ID, "T "+st, "echo", "atelier/"+id, "main", "/tmp/wt-"+id)
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
