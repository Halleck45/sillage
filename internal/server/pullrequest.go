package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PullRequest est une pull request GitHub telle que gh la décrit. Seuls les
// champs utiles à la livraison et à l'adoption d'une branche sont lus.
type PullRequest struct {
	Number            int    `json:"number"`
	URL               string `json:"url"`
	State             string `json:"state"` // OPEN | MERGED | CLOSED
	HeadRefName       string `json:"headRefName"`
	BaseRefName       string `json:"baseRefName"`
	IsCrossRepository bool   `json:"isCrossRepository"`
}

// prStateOf ramène l'état gh (OPEN, MERGED, CLOSED) au vocabulaire de
// CardBranch.PrState (open, merged, closed).
func prStateOf(pr PullRequest) string {
	return strings.ToLower(pr.State)
}

const prJSONFields = "number,url,state,headRefName,baseRefName,isCrossRepository"

// prForge regroupe les appels à la forge qui concernent les pull requests :
// tous en lecture, sauf open (la création, appelée par la seule livraison).
// Variable de paquet pour que les tests remplacent gh sans réseau.
type prForge struct {
	provider func(dir string) string
	find     func(ctx context.Context, dir, branch string) (PullRequest, bool, error)
	view     func(ctx context.Context, dir string, number int) (PullRequest, error)
	open     func(dir, branch, base, title, body string) (prURL, fallbackURL string, err error)
}

var forgePR = prForge{
	provider: func(dir string) string { return DetectForge(dir).Provider },
	find:     ghFindPR,
	view:     ghViewPR,
	open:     OpenPR,
}

// ghFindPR cherche la pull request dont la tête est branch, dans le dépôt
// même (jamais un fork homonyme). Une PR ouverte l'emporte sur tout ; à
// défaut, la plus récente fusionnée. Une PR fermée sans fusion est ignorée :
// la relivraison en ouvrira une nouvelle. found est faux quand il n'y en a
// aucune, err ne dit que l'échec de gh.
func ghFindPR(ctx context.Context, dir, branch string) (PullRequest, bool, error) {
	out, err := runGh(ctx, dir, "pr", "list", "--head", branch, "--state", "all",
		"--json", prJSONFields, "--limit", "20")
	if err != nil {
		return PullRequest{}, false, err
	}
	var prs []PullRequest
	if err := json.Unmarshal(out, &prs); err != nil {
		return PullRequest{}, false, errors.New("could not read the list returned by gh")
	}
	return pickPR(prs, branch)
}

// pickPR applique la règle de ghFindPR à une liste déjà lue (gh la trie de la
// plus récente à la plus ancienne).
func pickPR(prs []PullRequest, branch string) (PullRequest, bool, error) {
	var merged *PullRequest
	for i := range prs {
		pr := prs[i]
		if pr.IsCrossRepository || pr.HeadRefName != branch {
			continue
		}
		switch pr.State {
		case "OPEN":
			return pr, true, nil
		case "MERGED":
			if merged == nil {
				merged = &prs[i]
			}
		}
	}
	if merged != nil {
		return *merged, true, nil
	}
	return PullRequest{}, false, nil
}

// ghViewPR lit une pull request par son numéro.
func ghViewPR(ctx context.Context, dir string, number int) (PullRequest, error) {
	out, err := runGh(ctx, dir, "pr", "view", strconv.Itoa(number), "--json", prJSONFields)
	if err != nil {
		return PullRequest{}, err
	}
	var pr PullRequest
	if err := json.Unmarshal(out, &pr); err != nil {
		return PullRequest{}, errors.New("could not read the pull request returned by gh")
	}
	return pr, nil
}

// pullURLRe reconnaît l'URL d'une pull request (GitHub) ou d'une merge request
// (GitLab) existante, par opposition à une page de création pré-remplie.
var pullURLRe = regexp.MustCompile(`^https?://[^?#]+/(pull/\d+|-/merge_requests/\d+)/?$`)

// isPullURL dit si u désigne une pull request réelle. Les versions antérieures
// enregistraient l'URL de compare dans PrURL quand gh échouait : elle n'est pas
// une PR et ne doit ni s'afficher comme telle ni bloquer une nouvelle ouverture.
func isPullURL(u string) bool {
	return pullURLRe.MatchString(u)
}

// prLookupTTL borne la fréquence des `gh pr list` lancés par l'aperçu de
// livraison : il est relu à chaque ouverture d'un chantier.
const prLookupTTL = time.Minute

// prLookups mémorise, par dossier et branche, la dernière recherche de PR.
type prLookups struct {
	mu   sync.Mutex
	last map[string]time.Time
}

func (p *prLookups) due(key string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.last == nil {
		p.last = map[string]time.Time{}
	}
	if t, ok := p.last[key]; ok && time.Since(t) < prLookupTTL {
		return false
	}
	p.last[key] = time.Now()
	return true
}

// syncBranchPR rapproche la PR connue d'une branche de chantier de ce que dit
// la forge : une PR ouverte à la main est rattachée, une PR fusionnée ou
// fermée change d'état, une URL de compare héritée est oubliée. Lecture seule
// côté forge. force ignore le cache (livraison) ; sinon une recherche par
// minute et par branche au plus (aperçu). Seul GitHub est interrogé : sur une
// autre forge, b est rendu tel quel (URL de compare mise à part).
func (s *Server) syncBranchPR(cardID string, b CardBranch, force bool) CardBranch {
	if b.PrURL != "" && !isPullURL(b.PrURL) {
		b.PrURL, b.PrState = "", ""
		if _, err := s.store.SetCardBranchPR(cardID, b.RepoName, "", ""); err != nil {
			return b
		}
	}
	if b.WorktreeDir == "" || forgePR.provider(b.WorktreeDir) != "github" {
		return b
	}
	if !force && !s.prLookups.due(b.WorktreeDir+"\x00"+b.Branch) {
		return b
	}
	ctx, cancel := context.WithTimeout(context.Background(), watchGhTimeout)
	defer cancel()
	pr, found, err := forgePR.find(ctx, b.WorktreeDir, b.Branch)
	if err != nil {
		return b
	}
	url, state := "", ""
	if found {
		url, state = pr.URL, prStateOf(pr)
	} else if b.PrState == "open" || b.PrState == "merged" {
		// Plus rien côté forge pour une PR qu'on croyait vivante : supprimée,
		// ou fermée sans fusion. L'URL reste, l'état dit qu'elle est fermée.
		url, state = b.PrURL, "closed"
	} else {
		url, state = b.PrURL, b.PrState
	}
	if url == b.PrURL && state == b.PrState {
		return b
	}
	if _, err := s.store.SetCardBranchPR(cardID, b.RepoName, url, state); err == nil {
		b.PrURL, b.PrState = url, state
	}
	return b
}

// validBranchName refuse ce qui ne serait pas un nom de branche git simple :
// le nom de tête d'une PR vient de la forge, il finit dans `git push origin
// <branche>` et ne doit jamais pouvoir se lire comme une refspec ou une option.
func validBranchName(dir, name string) error {
	if name == "" || strings.HasPrefix(name, "-") || strings.ContainsAny(name, ": \t\n") {
		return fmt.Errorf("invalid branch name %q", name)
	}
	if _, err := runGit(dir, gitDefaultTimeout, "check-ref-format", "--branch", name); err != nil {
		return fmt.Errorf("invalid branch name %q", name)
	}
	return nil
}

// adoptPRBranch prépare la branche de chantier d'un dépôt sur la tête d'une
// pull request existante, au lieu d'une branche sillage/ws-… neuve : le
// chantier retouche la PR, et sa livraison pousse sur elle. Réservé aux PR
// ouvertes du dépôt même (un fork demanderait un autre remote).
func (s *Server) adoptPRBranch(card Card, repo Repo, number int) (CardBranch, error) {
	if forgePR.provider(repo.Path) != "github" {
		return CardBranch{}, errors.New("attaching a pull request needs a GitHub origin")
	}
	ctx, cancel := context.WithTimeout(context.Background(), watchGhTimeout)
	defer cancel()
	pr, err := forgePR.view(ctx, repo.Path, number)
	if err != nil {
		return CardBranch{}, fmt.Errorf("could not read pull request #%d: %w", number, err)
	}
	if pr.IsCrossRepository {
		return CardBranch{}, fmt.Errorf("pull request #%d comes from a fork: Sillage can only push to branches of origin", number)
	}
	if pr.State != "OPEN" {
		return CardBranch{}, fmt.Errorf("pull request #%d is %s", number, strings.ToLower(pr.State))
	}
	if err := validBranchName(repo.Path, pr.HeadRefName); err != nil {
		return CardBranch{}, err
	}
	if err := validBranchName(repo.Path, pr.BaseRefName); err != nil {
		return CardBranch{}, err
	}
	dir, err := CreatePRWorktree(repo.Path, s.dataDir, card.ID, repo.Name, pr.HeadRefName, pr.BaseRefName)
	if err != nil {
		return CardBranch{}, fmt.Errorf("failed to check out pull request #%d: %w", number, err)
	}
	return CardBranch{
		RepoName: repo.Name, Branch: pr.HeadRefName, Base: pr.BaseRefName, WorktreeDir: dir,
		PrURL: pr.URL, PrState: "open", Adopted: true,
	}, nil
}
