# Aegis

<div align="center">
<img src="assets/aegis-banner.svg" alt="Aegis banner" width="860">

<p><strong>Secure credential proxy for AI agents</strong></p>

<p>
  <img src="https://img.shields.io/badge/version-2.0.0-F2994A" alt="Version">
  <a href="https://github.com/yagna-1/aegis/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-orange" alt="License"></a>
  <a href="https://github.com/yagna-1/aegis/blob/main/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/code%20of%20conduct-Contributor%20Covenant-E78661" alt="Code of Conduct"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/go-1.21%2B-00ADD8" alt="Go version"></a>
  <a href="https://github.com/yagna-1/aegis/releases"><img src="https://img.shields.io/badge/releases-none-9CA3AF" alt="Releases"></a>
  <a href="https://github.com/yagna-1/aegis/packages"><img src="https://img.shields.io/badge/packages-none-9CA3AF" alt="Packages"></a>
  <a href="https://github.com/yagna-1/aegis/tags"><img src="https://img.shields.io/badge/tags-view-D97757" alt="Tags"></a>
  <a href="https://github.com/yagna-1/aegis/stargazers"><img src="https://img.shields.io/github/stars/yagna-1/aegis?color=F2994A" alt="GitHub stars"></a>
  <a href="https://github.com/yagna-1/aegis/commits/main"><img src="https://img.shields.io/github/last-commit/yagna-1/aegis?color=D97757" alt="Last commit"></a>
</p>

<p>
  <a href="#quickstart-5-minutes"><strong>Quickstart</strong></a>
  ·
  <a href="#ide-integration-mcp-mode"><strong>MCP Integration</strong></a>
  ·
  <a href="#security-model"><strong>Security Model</strong></a>
</p>
</div>

A lightweight credential proxy for AI agent workflows. Sits between your agent and any API — injecting real secrets at the network boundary. The agent **never sees credentials**.

Single-binary Go runtime with a compact footprint: **~17 MB executable**, **~660 KB repo size**, and **~76 KB internal source**.

Works natively as an MCP server with **Cursor**, **Claude Desktop**, **VS Code / Cline**, and **Windsurf**.

---

## Project

- Version: `2.0.0`
- License: [Apache 2.0](https://github.com/yagna-1/aegis/blob/main/LICENSE)
- Code of Conduct: [Contributor Covenant](https://github.com/yagna-1/aegis/blob/main/CODE_OF_CONDUCT.md)
- Releases: [No releases published](https://github.com/yagna-1/aegis/releases)
- Packages: [No packages published](https://github.com/yagna-1/aegis/packages)
- Tags: [View tags](https://github.com/yagna-1/aegis/tags)

---

## The problem

```
Agent context / system prompt:
  GITHUB_TOKEN=ghp_abc123   ← agent can read this
  STRIPE_SECRET_KEY=sk_...  ← one prompt injection and it's gone
```

## The fix

```
Agent  ──►  http://localhost:8080  ──►  Aegis  ──►  api.github.com
              (no key visible)              │
                                           └── pulls GITHUB_TOKEN from Infisical
                                               at request time, never stored locally
```

---

## Quickstart (5 minutes)

### 1. Install Aegis
```bash
go install github.com/yagna-1/aegis/cmd/aegis@latest
```

Or install a prebuilt binary from the `v2.0.0` release:

```bash
# macOS (Apple Silicon)
curl -L https://github.com/yagna-1/aegis/releases/download/v2.0.0/aegis-darwin-arm64 -o aegis
chmod +x aegis && sudo mv aegis /usr/local/bin/aegis

# macOS (Intel)
curl -L https://github.com/yagna-1/aegis/releases/download/v2.0.0/aegis-darwin-amd64 -o aegis
chmod +x aegis && sudo mv aegis /usr/local/bin/aegis

# Linux (amd64)
curl -L https://github.com/yagna-1/aegis/releases/download/v2.0.0/aegis-linux-amd64 -o aegis
chmod +x aegis && sudo mv aegis /usr/local/bin/aegis
```

```powershell
# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/yagna-1/aegis/releases/download/v2.0.0/aegis-windows-amd64.exe" -OutFile "aegis.exe"
```

```bash
# Optional: verify checksums
curl -L https://github.com/yagna-1/aegis/releases/download/v2.0.0/checksums.txt -o checksums.txt
shasum -a 256 aegis
```

### 2. Scaffold your project
```bash
cd ~/projects/my-agent
aegis init -services github,slack,stripe
```

### 3. Connect to Infisical (free, open source)
```bash
# Sign up free at https://app.infisical.com
# Create a Machine Identity with Universal Auth
# (Guide: https://infisical.com/docs/documentation/platform/identities/universal-auth)

aegis infisical setup
# → stores Client ID + Secret in your OS keychain
# → your actual API secrets stay in Infisical, never on disk
```

### 4. Add the Infisical block to aegis.yaml
```yaml
infisical:
  project_id: your-project-id
  environment: dev
```

### 5. Add your real secrets to Infisical
In the Infisical dashboard, add `GITHUB_TOKEN`, `STRIPE_SECRET_KEY`, etc.

### 6. Start and verify
```bash
aegis status   # checks connection + shows which secrets are found
aegis          # start proxy on :8080
```

---

## How secrets flow

```
Infisical dashboard
  → stores GITHUB_TOKEN, STRIPE_SECRET_KEY, etc.
  → your actual secrets never touch your machine's filesystem

aegis infisical setup
  → stores only the Machine Identity credentials in OS keychain
  → two values: Client ID (not secret) + Client Secret

aegis starts
  → authenticates with Infisical API using Machine Identity
  → fetches all secrets into process memory
  → auto-refreshes every 5 minutes (rotated secrets propagate automatically)

Agent calls http://localhost:8080
  → Aegis injects the real credential header
  → forwards to allowlisted domain
  → agent sees the API response, never the key
```

---

## Per-project isolation

Aegis auto-discovers `aegis.yaml` by walking up from the working directory — like `.git`:

```
~/projects/
  stripe-app/
    aegis.yaml   ← infisical.environment: prod, allowlist: api.stripe.com only
  github-bot/
    aegis.yaml   ← infisical.environment: dev, allowlist: api.github.com only
```

MCP mode inherits the IDE's working directory. Open a different project → different Infisical environment → different secrets. **No flags needed.**

---

## Secret priority order

| Source | How | When to use |
|---|---|---|
| **Infisical** | `infisical:` block in aegis.yaml | Recommended — team-friendly, rotatable |
| **OS Keychain** | `${keychain:name}` in templates | Good for solo devs, no external service |
| **.env file** | Auto-discovered in project root | Simple fallback, never commit to git |

---

## IDE integration (MCP mode)

One config, works for all projects (auto-discovers per-project aegis.yaml).

**Claude Desktop** — `~/Library/Application Support/Claude/claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "aegis": { "command": "aegis", "args": ["-mode", "mcp"] }
  }
}
```

**Cursor** — `.cursor/mcp.json`:
```json
{
  "mcpServers": {
    "aegis": { "command": "aegis", "args": ["-mode", "mcp"] }
  }
}
```

**VS Code / Cline / Windsurf** — same pattern, check your tool's MCP config path.

The IDE spawns `aegis -mode mcp` as a child process automatically. The agent calls the `http_request` tool — no auth headers, no secrets in context.

---

## Proxy mode — any language

```bash
# .env or system prompt instruction for the agent:
# "Call http://localhost:8080 with header X-Aegis-Target set to the full URL"

curl http://localhost:8080 \
  -H "X-Aegis-Target: https://api.github.com/user/repos"
```

Works with Python, TypeScript, Go, Ruby — any HTTP client.

---

## `aegis status`

```
Config:  /projects/my-agent/aegis.yaml
Port:    8080

Secrets: Infisical
  Site:    https://app.infisical.com
  Project: abc123
  Env:     dev
  Status:  ✓ connected, 8 secrets available

Allowlist (5 domains):
  • api.github.com
  • api.stripe.com
  • slack.com

Credentials (3 configured):
  • api.github.com     Authorization: ✓ loaded
  • api.stripe.com     Authorization: ✓ loaded
  • slack.com          Authorization: ✓ loaded
```

---

## Security model

| Threat | Mitigation |
|---|---|
| Agent reads secrets from context | Secrets in Infisical — never in agent context or disk |
| Agent exfiltrates credential via API | Domain allowlist — 403 for anything not listed |
| Agent sets its own auth header (MCP) | Explicitly rejected with error |
| Runaway agent burns API credits | Per-domain rate limit — configurable 429 |
| Prompt injection causes looping | Loop detection + X-Aegis-Warning header |
| Audit trail | NDJSON log — every request recorded, values never logged |
| Secret rotation | Infisical secrets refresh every 5 min — rotations propagate automatically |

---

## Self-hosting Infisical

Infisical is MIT-licensed and fully self-hostable. Point Aegis at your own instance:

```yaml
infisical:
  site_url: https://secrets.your-company.com
  project_id: your-project-id
  environment: dev
```

No data ever leaves your infrastructure.

---

## License

Apache 2.0
