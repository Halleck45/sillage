package server

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	maxProjectLinks    = 12
	linkFetchTimeout   = 5 * time.Second
	linkFetchMaxBytes  = 64 * 1024
	linkFetchMaxRedir  = 5
	linkFetchUserAgent = "Sillage/1.0 (+link preview)"
)

// NormalizeLinks valide une liste de liens épinglés : http(s) uniquement (pas
// de file://, ni aucun autre schéma), au plus maxProjectLinks. Ne fait aucun
// accès réseau (voir fetchLinkTitle, appelée par les handlers pour les liens
// sans titre avant persistance) : validation rapide, testable sans réseau.
func NormalizeLinks(links []Link) ([]Link, error) {
	if len(links) > maxProjectLinks {
		return nil, fmt.Errorf("at most %d links are allowed", maxProjectLinks)
	}
	out := make([]Link, len(links))
	for i, l := range links {
		raw := strings.TrimSpace(l.URL)
		if raw == "" {
			return nil, fmt.Errorf("link url is required")
		}
		u, err := url.Parse(raw)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return nil, fmt.Errorf("invalid link url %q: only http(s) URLs are allowed", raw)
		}
		out[i] = Link{URL: raw, Title: strings.TrimSpace(l.Title)}
	}
	return out, nil
}

// fillMissingLinkTitles complète, pour chaque lien sans titre fourni, le
// titre par récupération best-effort de la page (voir fetchLinkTitle).
// N'effectue jamais de validation : appeler NormalizeLinks avant.
func fillMissingLinkTitles(links []Link) []Link {
	out := make([]Link, len(links))
	for i, l := range links {
		if l.Title == "" {
			l.Title = fetchLinkTitle(l.URL)
		}
		out[i] = l
	}
	return out
}

// titleRe extrait le contenu de la première balise <title> d'un document HTML.
var titleRe = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

// fetchLinkTitle tente de récupérer le <title> de la page à rawURL : requête
// GET avec timeout global de 5 s, lecture plafonnée à 64 Ko, http(s)
// uniquement (y compris pour les redirections suivies, au plus 5). Best
// effort : ne bloque jamais et ne retourne jamais d'erreur, se rabat
// silencieusement sur le nom d'hôte en cas d'échec.
func fetchLinkTitle(rawURL string) string {
	fallback := hostnameOrURL(rawURL)

	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fallback
	}

	client := &http.Client{
		Timeout: linkFetchTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return fmt.Errorf("refusing redirect to non-http(s) scheme %q", req.URL.Scheme)
			}
			if len(via) >= linkFetchMaxRedir {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return fallback
	}
	req.Header.Set("User-Agent", linkFetchUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return fallback
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fallback
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, linkFetchMaxBytes))
	if title := extractTitle(string(body)); title != "" {
		return title
	}
	return fallback
}

// extractTitle nettoie et retourne le contenu de la première balise <title>
// trouvée dans body, chaîne vide si absente ou vide après nettoyage.
func extractTitle(body string) string {
	m := titleRe.FindStringSubmatch(body)
	if m == nil {
		return ""
	}
	title := html.UnescapeString(m[1])
	title = strings.Join(strings.Fields(title), " ") // aplatit espaces/retours à la ligne
	return strings.TrimSpace(title)
}

// hostnameOrURL retourne le nom d'hôte de rawURL, ou rawURL lui-même si le
// parse échoue ou ne produit pas de nom d'hôte exploitable.
func hostnameOrURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Hostname() == "" {
		return rawURL
	}
	return u.Hostname()
}
