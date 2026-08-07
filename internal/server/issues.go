package server

// Import d'une issue GitHub : lister les issues ouvertes d'un dépôt de projet
// avec le CLI gh, en LECTURE SEULE. Rien n'est écrit sur la forge (ni
// commentaire, ni assignation, ni fermeture), et rien n'est créé dans Sillage :
// l'issue choisie ne fait que pré-remplir le formulaire de création de tâche,
// que l'humain relit avant qu'un agent démarre.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	// issueListLimit : au-delà, ce n'est plus une liste qu'on parcourt des yeux
	// mais un backlog. La recherche est là pour aller chercher au-delà.
	issueListLimit = 50

	// issueListTimeout : gh appelle l'API GitHub. Court, sinon l'humain reste
	// devant une liste vide sans savoir si elle arrive.
	issueListTimeout = 20 * time.Second

	// issueBodyMax : le corps de l'issue part dans le prompt de la tâche, où
	// il est relu et corrigé. Tronquer garde la liste transportable et un
	// prompt lisible ; une issue-fleuve se complète à la main.
	issueBodyMax = 4000
)

// issueListFields : les champs demandés à gh. Le corps est du lot pour que
// choisir une issue n'ait pas à repartir chercher son contenu.
const issueListFields = "number,title,url,body,labels,updatedAt,author"

// ListGithubIssues retourne les issues ouvertes du dépôt de dir. search, s'il
// est renseigné, est passé à la recherche GitHub (mêmes qualificatifs que dans
// l'interface web : "label:bug", "author:...", ou du texte libre).
func ListGithubIssues(ctx context.Context, dir, search string) ([]GithubIssue, error) {
	if DetectForge(dir).Provider != "github" {
		return nil, errors.New("origin remote is not a github repository")
	}
	if _, err := lookPath("gh"); err != nil {
		return nil, errors.New("gh not found in PATH")
	}

	ctx, cancel := context.WithTimeout(ctx, issueListTimeout)
	defer cancel()

	args := []string{"issue", "list", "--state", "open",
		"--limit", strconv.Itoa(issueListLimit), "--json", issueListFields}
	if search = strings.TrimSpace(search); search != "" {
		args = append(args, "--search", search)
	}

	// Setpgid + Cancel comme pour la planification : sans quoi l'expiration du
	// délai tuerait gh mais laisserait tourner ses enfants.
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, errors.New("gh took too long to answer")
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s", truncate(msg, 300))
	}
	return parseGithubIssues(stdout.Bytes())
}

// parseGithubIssues lit la sortie --json de gh. Une entrée sans numéro ni
// titre n'est pas exploitable et est écartée plutôt que de faire échouer la
// liste entière.
func parseGithubIssues(raw []byte) ([]GithubIssue, error) {
	var items []struct {
		Number    int    `json:"number"`
		Title     string `json:"title"`
		URL       string `json:"url"`
		Body      string `json:"body"`
		UpdatedAt string `json:"updatedAt"`
		Labels    []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, errors.New("could not read the issue list returned by gh")
	}
	out := make([]GithubIssue, 0, len(items))
	for _, it := range items {
		title := strings.TrimSpace(it.Title)
		if it.Number == 0 || title == "" {
			continue
		}
		labels := make([]string, 0, len(it.Labels))
		for _, l := range it.Labels {
			if name := strings.TrimSpace(l.Name); name != "" {
				labels = append(labels, name)
			}
		}
		out = append(out, GithubIssue{
			Number:    it.Number,
			Title:     title,
			URL:       it.URL,
			Body:      truncate(normalizeIssueBody(it.Body), issueBodyMax),
			Labels:    labels,
			Author:    it.Author.Login,
			UpdatedAt: it.UpdatedAt,
		})
	}
	return out, nil
}

// normalizeIssueBody aplatit les fins de ligne Windows : le corps est recopié
// dans un textarea puis dans un prompt, où un \r traîne sans servir.
func normalizeIssueBody(body string) string {
	return strings.TrimSpace(strings.ReplaceAll(body, "\r\n", "\n"))
}
