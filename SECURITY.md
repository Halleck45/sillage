# Security Policy

## Reporting a vulnerability

Please do not open a public issue for security problems. Use GitHub's private vulnerability reporting ("Report a vulnerability" under the Security tab of this repository). You will get an answer within a week.

## Scope and model

Sillage is designed to run on a personal machine, bound to `127.0.0.1`, optionally reached through Tailscale or a TLS reverse proxy. Reports are especially welcome on:

- anything that lets an agent process push, publish, or exfiltrate without the explicit human ship confirmation;
- authentication or session weaknesses (login, cookies, rate limiting, CSRF);
- path traversal or command injection through project paths, task titles, branch names, or agent output.

Known and documented tradeoffs (not vulnerabilities): the HTTP port must not be exposed publicly without TLS termination in front, and `SILLAGE_CODEX_SANDBOX=danger-full-access` deliberately trades OS-level sandboxing of codex for Sillage's own containment (see README).
