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

	project, err := s1.AddProject("sillage", "/tmp/sillage")
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
	task, err := s1.CreateTask(taskID, ref, card.ID, project.ID, "Titre de tâche", "echo", "sillage/100-titre", "main", "/tmp/worktree")
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

	msg, task, err := s1.AddMessage(taskID, "user", "Alice", "bonjour")
	if err != nil {
		t.Fatalf("AddMessage: %v", err)
	}
	if task.MessagesCount != 1 {
		t.Fatalf("messagesCount attendu 1, reçu %d", task.MessagesCount)
	}
	if msg.Author != "user" || msg.Text != "bonjour" || msg.AuthorName != "Alice" {
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
	if loadedProject.Path != "/tmp/sillage" {
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

	project, err := s.AddProject("sillage", "/tmp/sillage")
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
		task, err := s.CreateTask(id, ref, card.ID, project.ID, "T "+st, "echo", "sillage/"+id, "main", "/tmp/wt-"+id)
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

// --- Migration des utilisateurs (v0.2) ---

func TestMigrateUsersCreatesAdminFromLegacyHash(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	if err := s.MigrateUsers("hash-from-config"); err != nil {
		t.Fatalf("MigrateUsers: %v", err)
	}
	admin, ok := s.FindUserByName("admin")
	if !ok {
		t.Fatalf("l'utilisateur admin devrait avoir été créé")
	}
	if admin.Role != "admin" {
		t.Fatalf("rôle attendu 'admin', reçu %q", admin.Role)
	}
	if admin.PasswordHash != "hash-from-config" {
		t.Fatalf("hash attendu 'hash-from-config', reçu %q", admin.PasswordHash)
	}

	// Un second appel sans SILLAGE_PASSWORD ne doit rien changer : la
	// migration a déjà eu lieu (des utilisateurs existent).
	if err := s.MigrateUsers("autre-hash"); err != nil {
		t.Fatalf("MigrateUsers (second appel): %v", err)
	}
	admin, _ = s.FindUserByName("admin")
	if admin.PasswordHash != "hash-from-config" {
		t.Fatalf("le hash admin ne devrait pas changer sans SILLAGE_PASSWORD, reçu %q", admin.PasswordHash)
	}

	// Avec SILLAGE_PASSWORD positionnée, le mot de passe admin est toujours remplacé.
	t.Setenv("SILLAGE_PASSWORD", "peu-importe")
	if err := s.MigrateUsers("hash-env"); err != nil {
		t.Fatalf("MigrateUsers (env): %v", err)
	}
	admin, _ = s.FindUserByName("admin")
	if admin.PasswordHash != "hash-env" {
		t.Fatalf("le hash admin devrait être remplacé par SILLAGE_PASSWORD, reçu %q", admin.PasswordHash)
	}
}

func TestMigrateUsersRoundtrip(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := s.MigrateUsers("hash-1"); err != nil {
		t.Fatalf("MigrateUsers: %v", err)
	}

	// Recharge depuis le disque : l'utilisateur admin doit avoir survécu,
	// c'est le scénario de compatibilité d'un state.json existant sans users.
	s2, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore (reload): %v", err)
	}
	admin, ok := s2.FindUserByName("admin")
	if !ok || admin.PasswordHash != "hash-1" {
		t.Fatalf("admin non persisté correctement après rechargement : %+v ok=%v", admin, ok)
	}
}

// --- Gardes utilisateurs et agents ---

func TestDeleteUserGuardsLastAdmin(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := s.MigrateUsers("hash-admin"); err != nil {
		t.Fatalf("MigrateUsers: %v", err)
	}
	admin, _ := s.FindUserByName("admin")

	if err := s.DeleteUser(admin.ID); err == nil {
		t.Fatalf("la suppression du dernier admin devrait être refusée")
	}

	member, err := s.AddUser("bob", "secret1234", "member")
	if err != nil {
		t.Fatalf("AddUser: %v", err)
	}
	if err := s.DeleteUser(member.ID); err != nil {
		t.Fatalf("la suppression d'un membre non-admin devrait réussir : %v", err)
	}

	second, err := s.AddUser("alice", "secret1234", "admin")
	if err != nil {
		t.Fatalf("AddUser (second admin): %v", err)
	}
	if err := s.DeleteUser(admin.ID); err != nil {
		t.Fatalf("la suppression d'un admin devrait réussir s'il en reste un autre : %v", err)
	}
	if err := s.DeleteUser(second.ID); err == nil {
		t.Fatalf("la suppression du dernier admin restant devrait être refusée")
	}
}

func TestUpdateUserGuardsLastAdminDemotion(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := s.MigrateUsers("hash-admin"); err != nil {
		t.Fatalf("MigrateUsers: %v", err)
	}
	admin, _ := s.FindUserByName("admin")

	member := "member"
	if _, err := s.UpdateUser(admin.ID, nil, &member); err == nil {
		t.Fatalf("rétrograder le dernier admin devrait être refusé")
	}

	if _, err := s.AddUser("alice", "secret1234", "admin"); err != nil {
		t.Fatalf("AddUser (second admin): %v", err)
	}
	if _, err := s.UpdateUser(admin.ID, nil, &member); err != nil {
		t.Fatalf("rétrograder un admin devrait réussir s'il en reste un autre : %v", err)
	}
}

func TestDeleteAgentGuardsReferencedAgent(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("sillage", "/tmp/sillage")
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	id, ref := s.ReserveTaskID()
	if _, err := s.CreateTask(id, ref, card.ID, project.ID, "T", "echo", "sillage/"+id, "main", "/tmp/wt"); err != nil {
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
	p, err := s.AddProject("sillage", "/tmp/sillage")
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}

	name := "Nouveau nom"
	checkCmd := "go test ./..."
	updated, err := s.UpdateProject(p.ID, &name, &checkCmd)
	if err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if updated.Name != name || updated.CheckCmd != checkCmd {
		t.Fatalf("mise à jour du projet inattendue : %+v", updated)
	}
}
