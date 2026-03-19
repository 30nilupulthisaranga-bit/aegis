package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/yagna-1/aegis/internal/audit"
	"github.com/yagna-1/aegis/internal/config"
	"github.com/yagna-1/aegis/internal/env"
	infisicalPkg "github.com/yagna-1/aegis/internal/infisical"
	"github.com/yagna-1/aegis/internal/keychain"
	"github.com/yagna-1/aegis/internal/mcp"
	"github.com/yagna-1/aegis/internal/proxy"
	"github.com/yagna-1/aegis/internal/resolver"
	"github.com/yagna-1/aegis/internal/scaffold"
)

const banner = `
  █████╗ ███████╗ ██████╗ ██╗███████╗
 ██╔══██╗██╔════╝██╔════╝ ██║██╔════╝
 ███████║█████╗  ██║  ███╗██║███████╗
 ██╔══██║██╔══╝  ██║   ██║██║╚════██║
 ██║  ██║███████╗╚██████╔╝██║███████║
 ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═╝╚══════╝
 Credential proxy for AI agents · v2.0.0
`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "infisical":
			runInfisical(os.Args[2:])
			return
		case "init":
			runInit(os.Args[2:])
			return
		case "keychain":
			runKeychain(os.Args[2:])
			return
		case "status":
			runStatus()
			return
		case "help", "-h", "--help":
			usage()
			return
		}
	}

	var (
		mode         = flag.String("mode", "proxy", "Run mode: proxy | mcp")
		explicitConf = flag.String("config", "", "Path to aegis.yaml (default: auto-discover)")
		explicitEnv  = flag.String("env", "", "Path to .env (default: auto-discover)")
		port         = flag.Int("port", 0, "Override port from config")
	)
	flag.Usage = usage
	flag.Parse()

	cwd, _ := os.Getwd()
	res, err := resolver.Find(cwd, *explicitConf, *explicitEnv)
	if err != nil {
		log.Fatalf("no aegis.yaml found — run `aegis init` first: %v", err)
	}

	cfg, err := config.Load(res.ConfigPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	if *port != 0 {
		cfg.Port = *port
	}

	if err := loadSecrets(cfg, res); err != nil {
		log.Fatalf("secrets: %v", err)
	}

	auditLog, err := audit.New(cfg.AuditLog)
	if err != nil {
		log.Fatalf("audit log: %v", err)
	}

	switch *mode {
	case "proxy":
		fmt.Print(banner)
		fmt.Printf("[aegis] config:    %s\n", res.ConfigPath)
		if cfg.Infisical != nil {
			fmt.Printf("[aegis] secrets:   Infisical (%s / %s)\n",
				cfg.Infisical.ProjectID, cfg.Infisical.Environment)
		}
		srv := proxy.New(cfg, auditLog)
		if err := srv.Run(cfg.Port); err != nil {
			log.Fatalf("proxy: %v", err)
		}
	case "mcp":
		srv := mcp.New(cfg, auditLog)
		srv.Run()
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", *mode)
		os.Exit(1)
	}
}

func loadSecrets(cfg *config.Config, res *resolver.Result) error {

	if cfg.Infisical != nil {
		infCfg := infisicalPkg.Config{
			SiteURL:            cfg.Infisical.SiteURL,
			ProjectID:          cfg.Infisical.ProjectID,
			Environment:        cfg.Infisical.Environment,
			SecretPath:         cfg.Infisical.SecretPath,
			ClientIDSource:     cfg.Infisical.ClientIDSource,
			ClientSecretSource: cfg.Infisical.ClientSecretSource,
			ClientID:           cfg.Infisical.ClientID,
			ClientSecret:       cfg.Infisical.ClientSecret,
		}

		client, err := infisicalPkg.New(infCfg, keychain.Get)
		if err != nil {
			return fmt.Errorf("infisical connection failed: %w", err)
		}
		n, err := client.LoadIntoEnv()
		if err != nil {
			return fmt.Errorf("infisical load failed: %w", err)
		}
		fmt.Printf("[aegis] loaded %d secrets from Infisical ✓\n", n)
		return nil
	}

	if res.EnvPath != "" {
		if err := env.Load(res.EnvPath); err != nil {
			fmt.Printf("[aegis] WARNING: .env load failed: %v\n", err)
		} else {
			fmt.Printf("[aegis] loaded secrets from %s\n", res.EnvPath)
		}
	}

	return nil
}

func runInfisical(args []string) {
	if len(args) == 0 {
		fmt.Print(infisicalUsage)
		os.Exit(1)
	}

	switch args[0] {
	case "setup":

		fmt.Println("Aegis × Infisical setup")
		fmt.Println("=======================")
		fmt.Println()
		fmt.Println("You need a Machine Identity with Universal Auth.")
		fmt.Println("Guide: https://infisical.com/docs/documentation/platform/identities/universal-auth")
		fmt.Println()

		clientID := prompt("Infisical Client ID: ")
		clientSecret := prompt("Infisical Client Secret: ")
		projectID := prompt("Project ID (from Project Settings): ")
		environment := promptWithDefault("Environment slug", "dev")

		must(keychain.Set("infisical_client_id", clientID), "storing client ID")
		must(keychain.Set("infisical_client_secret", clientSecret), "storing client secret")

		fmt.Println()
		fmt.Printf("✓ Credentials stored in OS keychain\n")
		fmt.Println()
		fmt.Println("Add this to your aegis.yaml:")
		fmt.Println()
		fmt.Println("  infisical:")
		fmt.Printf("    project_id: %s\n", projectID)
		fmt.Printf("    environment: %s\n", environment)
		fmt.Println("    # client_id_source: keychain  ← default, most secure")
		fmt.Println()
		fmt.Println("Your credential templates can reference Infisical secrets directly:")
		fmt.Println("  credentials:")
		fmt.Println("    api.github.com:")
		fmt.Println("      header: Authorization")
		fmt.Println("      template: \"Bearer ${GITHUB_TOKEN}\"  ← fetched from Infisical")
		fmt.Println()
		fmt.Println("Run `aegis status` to verify the connection.")

	case "test":

		cwd, _ := os.Getwd()
		res, err := resolver.Find(cwd, "", "")
		if err != nil {
			log.Fatal("no aegis.yaml found")
		}
		cfg, err := config.Load(res.ConfigPath)
		if err != nil {
			log.Fatalf("config: %v", err)
		}
		if cfg.Infisical == nil {
			log.Fatal("no `infisical:` block in aegis.yaml — run `aegis infisical setup` first")
		}

		fmt.Printf("Testing Infisical connection...\n")
		fmt.Printf("  Site:        %s\n", orDefault(cfg.Infisical.SiteURL, "https://app.infisical.com"))
		fmt.Printf("  Project:     %s\n", cfg.Infisical.ProjectID)
		fmt.Printf("  Environment: %s\n", orDefault(cfg.Infisical.Environment, "dev"))
		fmt.Printf("  Path:        %s\n\n", orDefault(cfg.Infisical.SecretPath, "/"))

		infCfg := infisicalPkg.Config{
			SiteURL:     cfg.Infisical.SiteURL,
			ProjectID:   cfg.Infisical.ProjectID,
			Environment: cfg.Infisical.Environment,
			SecretPath:  cfg.Infisical.SecretPath,
		}
		client, err := infisicalPkg.New(infCfg, keychain.Get)
		if err != nil {
			log.Fatalf("Connection FAILED: %v", err)
		}

		secrets, err := client.Secrets()
		if err != nil {
			log.Fatalf("Fetch FAILED: %v", err)
		}

		fmt.Printf("✓ Connected! Found %d secrets:\n", len(secrets))
		for k := range secrets {
			fmt.Printf("  • %s\n", k)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown infisical command: %s\n\n%s", args[0], infisicalUsage)
		os.Exit(1)
	}
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	servicesFlag := fs.String("services", "", "Comma-separated: github,slack,stripe,openai,anthropic,linear,notion,discord")
	portFlag := fs.Int("port", 8080, "Proxy port")
	forceFlag := fs.Bool("force", false, "Overwrite existing aegis.yaml")
	fs.Parse(args)

	var services []string
	if *servicesFlag != "" {
		for _, s := range strings.Split(*servicesFlag, ",") {
			services = append(services, strings.TrimSpace(s))
		}
	}

	if err := scaffold.Run(scaffold.Options{
		Services: services,
		Port:     *portFlag,
		Force:    *forceFlag,
	}); err != nil {
		log.Fatal(err)
	}
}

func runKeychain(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: aegis keychain set <name> <value>")
		fmt.Fprintln(os.Stderr, "       aegis keychain get <name>")
		os.Exit(1)
	}
	switch args[0] {
	case "set":
		if len(args) < 3 {
			log.Fatal("Usage: aegis keychain set <name> <value>")
		}
		must(keychain.Set(args[1], args[2]), "keychain set")
		fmt.Printf("✓ Stored %q in OS keychain\n", args[1])
	case "get":
		val, err := keychain.Get(args[1])
		must(err, "keychain get")
		fmt.Println(val)
	default:
		fmt.Fprintf(os.Stderr, "unknown keychain command: %s\n", args[0])
		os.Exit(1)
	}
}

func runStatus() {
	cwd, _ := os.Getwd()
	res, err := resolver.Find(cwd, "", "")
	if err != nil {
		fmt.Println("No aegis.yaml found. Run `aegis init`")
		return
	}

	cfg, err := config.Load(res.ConfigPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	fmt.Printf("Config:  %s\n", res.ConfigPath)
	fmt.Printf("Port:    %d\n\n", cfg.Port)

	if cfg.Infisical != nil {
		fmt.Printf("Secrets: Infisical\n")
		fmt.Printf("  Site:    %s\n", orDefault(cfg.Infisical.SiteURL, "https://app.infisical.com"))
		fmt.Printf("  Project: %s\n", cfg.Infisical.ProjectID)
		fmt.Printf("  Env:     %s\n", orDefault(cfg.Infisical.Environment, "dev"))

		infCfg := infisicalPkg.Config{
			SiteURL:     cfg.Infisical.SiteURL,
			ProjectID:   cfg.Infisical.ProjectID,
			Environment: cfg.Infisical.Environment,
			SecretPath:  cfg.Infisical.SecretPath,
		}
		client, err := infisicalPkg.New(infCfg, keychain.Get)
		if err != nil {
			fmt.Printf("  Status:  ✗ NOT CONNECTED — %v\n", err)
		} else {
			secrets, _ := client.Secrets()
			fmt.Printf("  Status:  ✓ connected, %d secrets available\n", len(secrets))
			client.LoadIntoEnv()
		}
	} else if res.EnvPath != "" {
		fmt.Printf("Secrets: .env (%s)\n", res.EnvPath)
		env.Load(res.EnvPath)
	} else {
		fmt.Printf("Secrets: none configured (run `aegis infisical setup` or add .env)\n")
	}

	fmt.Printf("\nAllowlist (%d domains):\n", len(cfg.Allowlist))
	for _, d := range cfg.Allowlist {
		fmt.Printf("  • %s\n", d)
	}

	fmt.Printf("\nCredentials (%d configured):\n", len(cfg.Credentials))
	for domain, cred := range cfg.Credentials {
		expanded := config.ExpandTemplate(cred.Template, keychain.Get)
		status := "✓ loaded"
		if expanded == "" || expanded == cred.Template {
			status = "✗ NOT FOUND"
		}
		fmt.Printf("  • %-32s %s: %s\n", domain, cred.Header, status)
	}

	if len(cfg.Budgets) > 0 {
		fmt.Printf("\nBudget guards (%d domains):\n", len(cfg.Budgets))
		for domain, b := range cfg.Budgets {
			fmt.Printf("  • %-32s max %d req/hr, loop threshold: %d\n",
				domain, b.MaxRequestsPerHour, b.LoopThreshold)
		}
	}
}

func prompt(label string) string {
	fmt.Print(label)
	var val string
	fmt.Scanln(&val)
	val = strings.TrimSpace(val)
	if val == "" {
		log.Fatalf("value required for: %s", label)
	}
	return val
}

func promptWithDefault(label, def string) string {
	fmt.Printf("%s [%s]: ", label, def)
	var val string
	fmt.Scanln(&val)
	val = strings.TrimSpace(val)
	if val == "" {
		return def
	}
	return val
}

func must(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

const infisicalUsage = `aegis infisical — Infisical secret backend

Commands:
  aegis infisical setup   Interactive setup: store machine identity in OS keychain
  aegis infisical test    Test connection and list available secret keys

After setup, add to aegis.yaml:

  infisical:
    project_id: your-project-id
    environment: dev           # or staging, prod, or any custom env slug
    # site_url: https://your-self-hosted-instance.com   (optional, for self-hosting)
    # secret_path: /           # folder path within the environment

Docs: https://infisical.com/docs/documentation/platform/identities/universal-auth
`

func usage() {
	fmt.Fprint(os.Stderr, `
Aegis v2 — credential proxy for AI agents

Usage:
  aegis [flags]                    Start proxy or MCP server
  aegis infisical setup            Connect to Infisical (interactive)
  aegis infisical test             Test Infisical connection
  aegis init [-services ...]       Scaffold aegis.yaml
  aegis status                     Show config + connection status
  aegis keychain set <name> <val>  Store in OS keychain
  aegis keychain get <name>        Retrieve from OS keychain

Flags:
  -mode   proxy|mcp    Run mode (default: proxy)
  -config path         aegis.yaml path (default: auto-discover)
  -env    path         .env path (default: fallback only)
  -port   int          Override config port

Secret priority (highest to lowest):
  1. Infisical (if aegis.yaml has 'infisical:' block)
  2. OS Keychain (${keychain:name} in credential templates)
  3. .env file (plaintext fallback)

`)
}
