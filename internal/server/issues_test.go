package server

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

// TestParseGithubIssuesFixture éprouve le parseur sur une sortie `gh issue list
// --json` réaliste : labels aplatis, auteur, fins de ligne Windows, et entrées
// inexploitables écartées sans faire échouer la liste.
func TestParseGithubIssuesFixture(t *testing.T) {
	raw := []byte(`[
	  {"number":221,"title":"Importer une issue github","url":"https://github.com/acme/demo/issues/221",
	   "body":"Ligne 1\r\nLigne 2\r\n","updatedAt":"2026-08-06T10:00:00Z",
	   "labels":[{"name":"bug"},{"name":""},{"name":"ux"}],"author":{"login":"jf"}},
	  {"number":0,"title":"sans numéro","url":"","body":"","labels":[],"author":{"login":""}},
	  {"number":7,"title":"   ","url":"","body":"","labels":[],"author":{"login":""}},
	  {"number":12,"title":"  Titre à trimmer  ","url":"https://github.com/acme/demo/issues/12",
	   "body":"","labels":null,"author":{"login":"ada"}}
	]`)
	issues, err := parseGithubIssues(raw)
	if err != nil {
		t.Fatalf("parseGithubIssues: %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("attendu 2 issues exploitables, obtenu %d : %+v", len(issues), issues)
	}
	first := issues[0]
	if first.Number != 221 || first.Title != "Importer une issue github" || first.Author != "jf" {
		t.Errorf("première issue mal lue : %+v", first)
	}
	if first.Body != "Ligne 1\nLigne 2" {
		t.Errorf("corps = %q, attendu les \\r retirés et le corps trimé", first.Body)
	}
	if got := strings.Join(first.Labels, ","); got != "bug,ux" {
		t.Errorf("labels = %q, attendu \"bug,ux\" (les vides écartés)", got)
	}
	if issues[1].Title != "Titre à trimmer" {
		t.Errorf("titre non trimé : %q", issues[1].Title)
	}
	if issues[1].Labels == nil {
		t.Error("labels doit être une liste vide, jamais null : le JSON de l'API ne doit pas porter de null")
	}
}

// TestParseGithubIssuesRejectsGarbage : une sortie qui n'est pas du JSON de
// liste remonte une erreur, pas une liste vide qui se lirait comme « aucune
// issue ouverte ».
func TestParseGithubIssuesRejectsGarbage(t *testing.T) {
	if _, err := parseGithubIssues([]byte("gh: command failed")); err == nil {
		t.Fatal("attendu une erreur sur une sortie non JSON")
	}
}

// TestListGithubIssuesRefusesNonGithubRepo : la liste n'est proposée que sur un
// dépôt github.com. Aucun appel réseau ici, le refus tombe avant gh.
func TestListGithubIssuesRefusesNonGithubRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}
	dir := t.TempDir()
	runTestGit(t, dir, "init")
	runTestGit(t, dir, "remote", "add", "origin", "git@gitlab.com:acme/demo.git")
	if _, err := ListGithubIssues(context.Background(), dir, ""); err == nil {
		t.Fatal("attendu un refus sur un dépôt qui n'est pas sur github.com")
	}
}

// TestProjectOutListsGithubRepos : c'est ce champ qui décide de l'affichage de
// l'entrée « Importer une issue » dans l'interface, il ne doit nommer que les
// dépôts réellement sur github.com.
func TestProjectOutListsGithubRepos(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}
	hub, lab := t.TempDir(), t.TempDir()
	runTestGit(t, hub, "init")
	runTestGit(t, hub, "remote", "add", "origin", "git@github.com:acme/demo.git")
	runTestGit(t, lab, "init")
	runTestGit(t, lab, "remote", "add", "origin", "git@gitlab.com:acme/demo.git")

	out := projectOut(Project{
		ID:       "p1",
		Name:     "Demo",
		Repos:    []Repo{{Name: "hub", Path: hub}, {Name: "lab", Path: lab}},
		Delivery: Delivery{Mode: "merge"},
	})
	if got := strings.Join(out.GithubRepos, ","); got != "hub" {
		t.Errorf("githubRepos = %q, attendu \"hub\"", got)
	}
}
