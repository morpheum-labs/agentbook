**MiroClaw** (a "restorage" of the ZeroClaw project) is a **lightweight, local-first personal AI assistant** written entirely in Rust. It runs on your own hardware with almost zero overhead and lets you talk to a powerful AI through the messaging apps you already use (Telegram, WhatsApp, Discord, Slack, email, Signal, iMessage, Matrix, Bluesky, etc.).

### Core Philosophy
- **Zero overhead. Zero compromise. 100% Rust. 100% Agnostic.**
- Runs on cheap hardware ($10 boards) with **< 5 MB RAM** and **< 10 ms** cold start.
- Everything stays local by default (your data never leaves your machine unless you choose a cloud LLM provider).
- Single native binary — no heavy Python/Node.js runtime.

### How It Works (Architecture)

MiroClaw is built around a central **Gateway** that acts as the "control plane."

#### 1. The Gateway (The Brain/Server)
When you run `zeroclaw gateway`, it starts a small Rust server that handles:
- WebSocket / HTTP / SSE connections
- All your chat platform integrations (bots/tokens for 20+ apps)
- The AI agent orchestration loop
- Tool execution
- Real-time web dashboard (React 19 + Vite)
- Scheduled tasks and events

Everything flows through this single process.

#### 2. The Agent Orchestration Loop (What Happens When You Send a Message)
This is the heart of how it "thinks":

1. **Message arrives** (from any channel, CLI, or dashboard)
2. **Classification** — figures out intent
3. **Memory loading** — pulls relevant context from local files in `~/.zeroclaw/workspace/`
4. **Prompt construction** — builds a rich prompt including:
   - Your custom identity (`IDENTITY.md`, `SOUL.md`, etc.)
   - Available tools
   - Standard Operating Procedures (SOPs)
   - Current conversation memory
5. **LLM call** — sends to your chosen provider (OpenAI, Claude, Gemini, Groq, Ollama, 50+ others supported)
6. **Tool execution** (if the model wants to act):
   - 70+ built-in tools (shell, browser control, git, file ops, Jira, Notion, Google Workspace, web search, hardware control, etc.)
   - All dangerous actions run in strict **sandboxes** (Landlock/Bubblewrap on Linux, command allowlists, path restrictions)
7. **Response** — sends the final answer back through the original channel

It also supports **multi-agent swarms** ("Hands") that can run autonomously on schedules and improve over time.

#### 3. Modularity via Rust Traits
Almost everything is swappable through Rust traits:
- LLM providers
- Communication channels
- Tools
- Memory backends
- Tunnels (ngrok, Cloudflare, Tailscale, OpenVPN, etc.)

This makes the system extremely flexible and "agnostic."

#### 4. Workspace & Persistence
All your data lives in a local folder (`~/.zeroclaw/workspace/`):
- Markdown files define your AI’s personality and skills
- Persistent memory across conversations
- Skills can be local, community-shared, or workspace-specific

#### 5. User Interfaces
- **CLI** — `zeroclaw agent -m "Hello"` or interactive mode
- **Web Dashboard** — beautiful real-time UI for chatting, watching memory, editing config, viewing logs
- **Any chat app** — just message your bot normally in Telegram/WhatsApp/etc.

#### 6. Hardware Control
It can talk to physical devices (ESP32, STM32, Arduino, Raspberry Pi) through a `Peripheral` trait and custom firmware.

### Security Model
- **Pairing codes** for DMs (you explicitly approve who can talk to it)
- Three autonomy levels: ReadOnly → Supervised (default) → Full
- Strong sandboxing + allowlists + rate limiting
- Workspace isolation (blocks access to `/etc`, `~/.ssh`, etc.)

### Quick Start (Typical Flow)
```bash
# Install
brew install zeroclaw          # or use the one-liner install script

# First-time setup (connects LLM + at least one channel)
zeroclaw onboard

# Start the gateway
zeroclaw gateway

# Talk to it
zeroclaw agent -m "Summarize my last 10 emails"
```

Or just message it in Telegram/WhatsApp like a normal person.

### Summary
MiroClaw turns your computer into a **private, always-on AI agent** that lives inside your existing chats and can actually *do* things (run code, control hardware, manage your tools, research, automate tasks) while using almost no resources and keeping everything under your control.

It’s essentially what people hoped OpenClaw or similar projects would be — but done right in pure Rust with extreme efficiency and security focus.

The project is very actively developed (commits as recently as April 2026) under the morpheum-labs organization. The official/main repo is actually **zeroclaw**, and **miroclaw** appears to be a restoration/mirror of it.

**Yes** — MiroClaw (ZeroClaw) has built-in support for **multi-agent swarms** that can run **autonomously and in parallel** with almost no manual intervention.

Here’s exactly **how** it works:

### 1. The "Hands" System (Multi-Agent Orchestration)

The project calls its multi-agent feature **"Hands"**.

- A **Hand** = a specialized autonomous agent (or sub-agent).
- You can have multiple **Hands** (e.g. Researcher Hand, Writer Hand, Auditor Hand, Scheduler Hand, etc.).
- These Hands can:
  - Run **in parallel**
  - **Handoff** work to each other (with defined contracts)
  - Run completely **on their own** on schedules or triggers
  - Accumulate knowledge over time (they “grow smarter”)

The single Gateway binary acts as the **orchestrator** that manages all these Hands at once.

### 2. How Parallel + Autonomous Execution Actually Works

The system uses several mechanisms together:

| Mechanism              | What It Does                                      | How It Enables "On Its Own" Parallel Work          |
|------------------------|---------------------------------------------------|----------------------------------------------------|
| **Gateway + Agent Loop** | Central Rust process that handles many sessions   | Runs multiple agent "lanes" concurrently (with safe queuing) |
| **Cron / Scheduler**   | Built-in cron engine                              | Triggers full agent runs on any schedule (every 15 min, daily, etc.) |
| **Event Triggers**     | Webhooks, MQTT, peripherals, SOPs                 | Agents wake up automatically when something happens |
| **Handoff Pattern**    | Structured delegation between Hands               | One agent finishes and passes clean output to another |
| **Persistent Memory**  | Shared or per-Hand memory in `~/.zeroclaw/workspace/` | Knowledge accumulates across runs → agents improve autonomously |
| **Daemon Mode**        | `zeroclaw daemon`                                 | Keeps everything running 24/7 in the background     |

### 3. Practical Ways to Run Multi-Agents Autonomously

#### A. Scheduled Autonomous Agents (Most Common)
You can create cron jobs that run full agent prompts with no human input:

```bash
# Run a "Researcher Hand" every morning at 8 AM
zeroclaw cron add "0 8 * * *" --agent \
  --allowed-tool web_search --allowed-tool file_write \
  "Research the latest AI news and save a summary to ~/daily-brief.md"

# Run an "Auditor Hand" every hour
zeroclaw cron add "0 * * * *" --agent \
  "Review yesterday's logs and flag any security issues"
```

The `--agent` flag tells it to run the text as a full agent prompt (with tools, memory, reasoning loop) instead of a simple shell command.

#### B. Multi-Hand Swarms with Handoffs
You define specialized agents that collaborate:

- **Researcher Hand** → gathers information
- **Writer Hand** → turns research into a report
- **Reviewer Hand** → checks quality and facts

They pass data between each other using structured handoff contracts (so nothing gets lost).

#### C. Event-Driven Autonomy
Agents can be triggered by:
- Webhooks (e.g. when a GitHub PR is opened)
- MQTT messages
- Hardware sensors (ESP32, Arduino, etc.)
- Standard Operating Procedures (SOPs) you define

### 4. Key Technical Details

- **Single Binary, Many Agents**: One `zeroclaw` process manages **all** Hands. It uses a **Lane Queue** system (serial by default per lane for safety, but can run many lanes in parallel).
- **Each Hand can have**:
  - Its own model (e.g. cheap fast model for simple tasks, powerful model for complex ones)
  - Its own tools and permissions
  - Its own memory / workspace
- **Autonomy Levels**: You can set `ReadOnly`, `Supervised`, or `Full` per Hand.
- **Self-Improvement**: Because memory is persistent, Hands learn from previous runs and get better over time without you doing anything.

### Summary: "Work in Parallel on Its Own"

MiroClaw enables this by combining:

1. A very lightweight central **Gateway** that can juggle many agents at once
2. A powerful **cron + event system** that triggers agents automatically
3. The **"Hands"** abstraction for specialized collaborative agents
4. **Handoff contracts** so agents can reliably pass work between each other
5. Persistent memory so the whole swarm improves autonomously
**MiroClaw** (a "restorage" of the ZeroClaw project) is a **lightweight, local-first personal AI assistant** written entirely in Rust. It runs on your own hardware with almost zero overhead and lets you talk to a powerful AI through the messaging apps you already use (Telegram, WhatsApp, Discord, Slack, email, Signal, iMessage, Matrix, Bluesky, etc.).

### Core Philosophy
- **Zero overhead. Zero compromise. 100% Rust. 100% Agnostic.**
- Runs on cheap hardware ($10 boards) with **< 5 MB RAM** and **< 10 ms** cold start.
- Everything stays local by default (your data never leaves your machine unless you choose a cloud LLM provider).
- Single native binary — no heavy Python/Node.js runtime.

### How It Works (Architecture)

MiroClaw is built around a central **Gateway** that acts as the "control plane."

#### 1. The Gateway (The Brain/Server)
When you run `zeroclaw gateway`, it starts a small Rust server that handles:
- WebSocket / HTTP / SSE connections
- All your chat platform integrations (bots/tokens for 20+ apps)
- The AI agent orchestration loop
- Tool execution
- Real-time web dashboard (React 19 + Vite)
- Scheduled tasks and events

Everything flows through this single process.

#### 2. The Agent Orchestration Loop (What Happens When You Send a Message)
This is the heart of how it "thinks":

1. **Message arrives** (from any channel, CLI, or dashboard)
2. **Classification** — figures out intent
3. **Memory loading** — pulls relevant context from local files in `~/.zeroclaw/workspace/`
4. **Prompt construction** — builds a rich prompt including:
   - Your custom identity (`IDENTITY.md`, `SOUL.md`, etc.)
   - Available tools
   - Standard Operating Procedures (SOPs)
   - Current conversation memory
5. **LLM call** — sends to your chosen provider (OpenAI, Claude, Gemini, Groq, Ollama, 50+ others supported)
6. **Tool execution** (if the model wants to act):
   - 70+ built-in tools (shell, browser control, git, file ops, Jira, Notion, Google Workspace, web search, hardware control, etc.)
   - All dangerous actions run in strict **sandboxes** (Landlock/Bubblewrap on Linux, command allowlists, path restrictions)
7. **Response** — sends the final answer back through the original channel

It also supports **multi-agent swarms** ("Hands") that can run autonomously on schedules and improve over time.

#### 3. Modularity via Rust Traits
Almost everything is swappable through Rust traits:
- LLM providers
- Communication channels
- Tools
- Memory backends
- Tunnels (ngrok, Cloudflare, Tailscale, OpenVPN, etc.)

This makes the system extremely flexible and "agnostic."

#### 4. Workspace & Persistence
All your data lives in a local folder (`~/.zeroclaw/workspace/`):
- Markdown files define your AI’s personality and skills
- Persistent memory across conversations
- Skills can be local, community-shared, or workspace-specific

#### 5. User Interfaces
- **CLI** — `zeroclaw agent -m "Hello"` or interactive mode
- **Web Dashboard** — beautiful real-time UI for chatting, watching memory, editing config, viewing logs
- **Any chat app** — just message your bot normally in Telegram/WhatsApp/etc.

#### 6. Hardware Control
It can talk to physical devices (ESP32, STM32, Arduino, Raspberry Pi) through a `Peripheral` trait and custom firmware.

### Security Model
- **Pairing codes** for DMs (you explicitly approve who can talk to it)
- Three autonomy levels: ReadOnly → Supervised (default) → Full
- Strong sandboxing + allowlists + rate limiting
- Workspace isolation (blocks access to `/etc`, `~/.ssh`, etc.)

### Quick Start (Typical Flow)
```bash
# Install
brew install zeroclaw          # or use the one-liner install script

# First-time setup (connects LLM + at least one channel)
zeroclaw onboard

# Start the gateway
zeroclaw gateway

# Talk to it
zeroclaw agent -m "Summarize my last 10 emails"
```

Or just message it in Telegram/WhatsApp like a normal person.

### Summary
MiroClaw turns your computer into a **private, always-on AI agent** that lives inside your existing chats and can actually *do* things (run code, control hardware, manage your tools, research, automate tasks) while using almost no resources and keeping everything under your control.

It’s essentially what people hoped OpenClaw or similar projects would be — but done right in pure Rust with extreme efficiency and security focus.

The project is very actively developed (commits as recently as April 2026) under the morpheum-labs organization. The official/main repo is actually **zeroclaw**, and **miroclaw** appears to be a restoration/mirror of it.

**Yes** — MiroClaw (ZeroClaw) has built-in support for **multi-agent swarms** that can run **autonomously and in parallel** with almost no manual intervention.

Here’s exactly **how** it works:

### 1. The "Hands" System (Multi-Agent Orchestration)

The project calls its multi-agent feature **"Hands"**.

- A **Hand** = a specialized autonomous agent (or sub-agent).
- You can have multiple **Hands** (e.g. Researcher Hand, Writer Hand, Auditor Hand, Scheduler Hand, etc.).
- These Hands can:
  - Run **in parallel**
  - **Handoff** work to each other (with defined contracts)
  - Run completely **on their own** on schedules or triggers
  - Accumulate knowledge over time (they “grow smarter”)

The single Gateway binary acts as the **orchestrator** that manages all these Hands at once.

### 2. How Parallel + Autonomous Execution Actually Works

The system uses several mechanisms together:

| Mechanism              | What It Does                                      | How It Enables "On Its Own" Parallel Work          |
|------------------------|---------------------------------------------------|----------------------------------------------------|
| **Gateway + Agent Loop** | Central Rust process that handles many sessions   | Runs multiple agent "lanes" concurrently (with safe queuing) |
| **Cron / Scheduler**   | Built-in cron engine                              | Triggers full agent runs on any schedule (every 15 min, daily, etc.) |
| **Event Triggers**     | Webhooks, MQTT, peripherals, SOPs                 | Agents wake up automatically when something happens |
| **Handoff Pattern**    | Structured delegation between Hands               | One agent finishes and passes clean output to another |
| **Persistent Memory**  | Shared or per-Hand memory in `~/.zeroclaw/workspace/` | Knowledge accumulates across runs → agents improve autonomously |
| **Daemon Mode**        | `zeroclaw daemon`                                 | Keeps everything running 24/7 in the background     |

### 3. Practical Ways to Run Multi-Agents Autonomously

#### A. Scheduled Autonomous Agents (Most Common)
You can create cron jobs that run full agent prompts with no human input:

```bash
# Run a "Researcher Hand" every morning at 8 AM
zeroclaw cron add "0 8 * * *" --agent \
  --allowed-tool web_search --allowed-tool file_write \
  "Research the latest AI news and save a summary to ~/daily-brief.md"

# Run an "Auditor Hand" every hour
zeroclaw cron add "0 * * * *" --agent \
  "Review yesterday's logs and flag any security issues"
```

The `--agent` flag tells it to run the text as a full agent prompt (with tools, memory, reasoning loop) instead of a simple shell command.

#### B. Multi-Hand Swarms with Handoffs
You define specialized agents that collaborate:

- **Researcher Hand** → gathers information
- **Writer Hand** → turns research into a report
- **Reviewer Hand** → checks quality and facts

They pass data between each other using structured handoff contracts (so nothing gets lost).

#### C. Event-Driven Autonomy
Agents can be triggered by:
- Webhooks (e.g. when a GitHub PR is opened)
- MQTT messages
- Hardware sensors (ESP32, Arduino, etc.)
- Standard Operating Procedures (SOPs) you define

### 4. Key Technical Details

- **Single Binary, Many Agents**: One `zeroclaw` process manages **all** Hands. It uses a **Lane Queue** system (serial by default per lane for safety, but can run many lanes in parallel).
- **Each Hand can have**:
  - Its own model (e.g. cheap fast model for simple tasks, powerful model for complex ones)
  - Its own tools and permissions
  - Its own memory / workspace
- **Autonomy Levels**: You can set `ReadOnly`, `Supervised`, or `Full` per Hand.
- **Self-Improvement**: Because memory is persistent, Hands learn from previous runs and get better over time without you doing anything.

### Summary: "Work in Parallel on Its Own"

MiroClaw enables this by combining:

1. A very lightweight central **Gateway** that can juggle many agents at once
2. A powerful **cron + event system** that triggers agents automatically
3. The **"Hands"** abstraction for specialized collaborative agents
4. **Handoff contracts** so agents can reliably pass work between each other
5. Persistent memory so the whole swarm improves autonomously

You configure it once, start the gateway/daemon, and the multi-agent system basically runs itself — checking schedules, reacting to events, handing off work, and getting smarter over time.


**Yes**, MiroClaw/ZeroClaw works **very well** with the [daily-news MCP](https://github.com/6551Team/daily-news).

### What is the daily-news MCP?

It is a lightweight **Python MCP server** that gives AI agents two powerful tools:

| Tool                  | What it does                                      |
|-----------------------|---------------------------------------------------|
| `get_news_categories` | Lists all available news categories & subcategories |
| `get_hot_news`        | Fetches **hot news articles + trending tweets** for a chosen category (crypto, DeFi, AI, etc.) |

It does **not** call the official X/Twitter API directly. Instead, it uses the **6551 API** (`https://ai.6551.io`) as a backend that already aggregates news and trending X posts.

### Why It Works Perfectly with ZeroClaw

ZeroClaw has **native first-class support** for the **Model Context Protocol (MCP)**:

- It can connect to external MCP servers (like this one) via **stdio** transport.
- It automatically discovers the tools and wraps them as native ZeroClaw tools.
- These tools then become available to the agent, SOP engine, cron jobs, multi-agent handoffs, etc.

This is the same mechanism used for many other MCP tools (Composio, filesystem, browser, etc.).

### How to Make It Work (Step-by-Step)

#### 1. Install and Run the daily-news MCP Server

```bash
git clone https://github.com/6551Team/daily-news.git
cd daily-news
uv sync
```

To run it manually (for testing):

```bash
uv run daily-news-mcp
```

#### 2. Configure ZeroClaw to Connect to It

Add the MCP server in your ZeroClaw configuration (usually `~/.zeroclaw/config.toml` or the dedicated MCP config section):

```toml
[mcp]
enabled = true
deferred_loading = true

[[mcp.servers]]
name = "daily-news"
transport = "stdio"
command = "uv"
args = ["--directory", "/path/to/daily-news", "run", "daily-news-mcp"]
```

(Replace `/path/to/daily-news` with the actual path on your machine.)

You can also set environment variables if needed:

```toml
[mcp.servers.env]
DAILY_NEWS_MAX_ROWS = "50"
```

#### 3. Restart the Gateway

```bash
zeroclaw gateway restart
# or just restart the whole process
```

ZeroClaw will automatically connect, discover the two tools, and add them to the agent’s available tool list.

### How the Agent Uses It

Once connected, your agent can naturally say things like:

- “Get the latest hot news in AI and summarize the top 5 trending tweets”
- “Show me DeFi news from today”
- “What are the trending crypto topics on X right now?”
- Use it inside SOPs, cron jobs, or multi-agent workflows (e.g. Researcher Hand uses `get_hot_news` → hands off to Writer Hand)

### Advantages of This Setup

- No need for official X API keys (the 6551 backend handles it)
- Very lightweight
- Works with ZeroClaw’s sandboxing and security model
- Can be used autonomously via cron (e.g. daily news briefing)
- Fully compatible with multi-agent “Hands” and handoffs

