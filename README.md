# Site vitrine de Sillage

Cette branche `documentation` contient **les sources** du site vitrine.
La branche `gh-pages` contient **uniquement le build** : elle est écrite par la CI,
il ne faut jamais l'éditer à la main.

```
documentation  --(GitHub Actions : jekyll build)-->  gh-pages  -->  GitHub Pages
```

## Contenu

| Chemin                              | Rôle                                             |
| ----------------------------------- | ------------------------------------------------ |
| `index.html`                        | Le site, en une seule page (HTML, CSS et JS inline pour l'animation) |
| `assets/style.css`                  | Toute la mise en forme                            |
| `assets/*.png`                      | Captures d'écran de l'application                 |
| `_config.yml`                       | Configuration Jekyll                              |
| `.github/workflows/documentation.yml` | Build et publication vers `gh-pages`            |

## Travailler en local

```bash
jekyll serve      # http://127.0.0.1:4000
```

Ou simplement `jekyll build`, puis ouvrir `_site/index.html`.
Le dossier `_site/` est ignoré par git : c'est la CI qui produit la version publiée.

## Publication

Tout push sur `documentation` déclenche le workflow, qui construit le site et
remplace le contenu de `gh-pages`. Le workflow est aussi lançable à la main
depuis l'onglet Actions (`workflow_dispatch`).
