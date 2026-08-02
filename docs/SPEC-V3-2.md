# Sillage : fournée corrective (réassignation, santé codex, tokens codex)

Issue d'un cas d'usage réel : un agent codex bloqué par le sandbox de la machine a laissé une tâche orpheline, sans moyen de la confier à un autre agent, et l'utilisateur n'a découvert le problème codex qu'en cours de tâche. Erreurs API en anglais, i18n fr+en.

## 1. Réassigner une tâche à un autre agent

- PATCH `/api/tasks/{id}` `{agentId}` : refusé si status=running (400 `"interrupt the agent before reassigning"`), refusé si agent inconnu. Effets : `task.agentId = nouveau`, `task.sessionId = ""` (le nouvel agent ne peut pas reprendre la session CLI de l'ancien), message système ajouté au fil (author=agent, authorName vide, texte : clé côté front, contenu backend : `"→ Task reassigned to <name>"` en anglais neutre... NON : pour rester i18n-propre, le backend ajoute un Message avec author="agent", authorName="" et text figé `"[reassigned:<agentId>]"`) ; le frontend détecte ce marqueur et affiche une ligne système localisée « Tâche réassignée à X ». SSE task + message + agents.
- Au prochain démarrage d'agent sur cette tâche (nouveau message), session vide = départ frais : préfixer le texte envoyé au CLI par un rappel de contexte minimal : `Task: <title>\n\n<dernier message utilisateur>` (déjà le cas au lancement initial ; vérifier que le chemin resume-vide le fait).
- UI : dans l'en-tête du détail de tâche, le chip agent (emoji + nom) devient cliquable quand status != running : petit menu déroulant sobre listant les autres agents (emoji, nom, modèle, avertissement éventuel : voir section 2) ; choisir → PATCH → la conversation affiche la ligne système. Tooltip « Réassigner la tâche ».

## 2. Santé des agents visible en amont

- Backend : `Agent` gagne un champ calculé (non persisté) `warning: string` (vide si OK), rempli à chaque ListAgents :
  - cli=codex : si `/proc/sys/kernel/apparmor_restrict_unprivileged_userns` existe et vaut `1` ET que `SILLAGE_CODEX_SANDBOX` n'est pas défini -> `"codex sandbox is blocked on this machine (AppArmor); see README (SILLAGE_CODEX_SANDBOX)"`.
  - cli=codex ou claude : si le binaire (`codex`/`claude`) est introuvable dans le PATH -> `"<cli> CLI not found in PATH"`. (LookPath à chaque ListAgents est acceptable : appels rares.)
- UI : petit ⚠ ambre à côté du nom de l'agent partout où on le choisit (cartes de la modale nouvelle tâche, menu de réassignation, liste sidebar) avec le message en tooltip ; dans la modale d'édition d'agent, bandeau discret avec le message complet localisé (le frontend mappe les warnings connus vers des clés i18n, sinon affiche le texte brut).

## 3. Tokens codex : total final, pas la somme des cumulatifs

Bug réel constaté : 1,97M de tokens affichés pour une petite tâche. Les événements `token_count`/`turn.completed` de codex portent des TOTAUX CUMULÉS par exécution ; on les additionne aujourd'hui à chaque événement.

- runCodex : ne plus cumuler au fil des événements. Mémoriser le dernier `{input, output}` vu pendant l'exécution et l'ajouter UNE SEULE FOIS aux compteurs de la tâche à la fin du process (avant finalize). Publier `tokens` seulement à ce moment-là pour codex.
- Test unitaire : flux JSONL simulé avec 3 événements token_count croissants (100/10, 250/25, 400/40) -> la tâche doit gagner exactement 400 input / 40 output.

## Hors périmètre

Retry automatique d'un agent en échec, transfert de session entre CLIs.
