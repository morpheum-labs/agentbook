# SOP: MiroClaw (ZeroClaw) — per-agent / per-Hand cron jobs

**Yes**, MiroClaw (ZeroClaw) **fully supports** assigning cron jobs to **specific agents/Hands**.

You can define different agents in your configuration (mainly in `AGENTS.md` and workspace files), and then schedule them independently with different tasks and different time intervals.

## How it works

ZeroClaw’s cron system supports **two types** of jobs:

| Job Type      | Purpose                          | How to target a specific agent          |
|---------------|----------------------------------|----------------------------------------|
| **Shell Job** | Run normal shell commands        | Not agent-specific                      |
| **Agent Job** | Run a full agent reasoning cycle | Use `--agent` + prompt or `session_target` |

For **specific agents**, use **Agent Jobs**. You can target them in two main ways:

1. **By including the agent name in the prompt** (most common)
2. **Using `session_target`** (more advanced, for named sessions/Hands)

## Practical examples

### Example 1: different schedules for different agents

```bash
# Agent A (Researcher) - runs every 6 hours
zeroclaw cron add "0 */6 * * *" --agent \
  "You are the Researcher Hand. Research the latest developments in AI and save a summary to memory."

# Agent B (News Curator) - runs every day at 8:00 AM
zeroclaw cron add "0 8 * * *" --agent \
  "You are the News Curator Hand. Use get_hot_news tool for AI and Crypto categories, then send a summary to Telegram."

# Agent C (Weekly Reporter) - runs every Monday at 9:00 AM
zeroclaw cron add "0 9 * * MON" --agent \
  "You are the Weekly Reporter Hand. Generate a full weekly summary and email it."
```

### Example 2: using config file (recommended for persistence)

You can also define cron jobs directly in your configuration (more reliable for complex setups):

```toml
[[cron.jobs]]
id = "researcher-daily"
name = "Researcher Hand Task"
schedule.cron = "0 */6 * * *"
prompt = "You are the Researcher Hand. Deep research on current AI trends and update memory."
model = "claude-sonnet-4"           # optional
# session_target = "researcher"     # if you have named sessions

[[cron.jobs]]
id = "news-curator-morning"
name = "News Curator Hand"
schedule.cron = "0 8 * * *"
prompt = "You are the News Curator Hand. Fetch today's hot news using get_hot_news and post to Telegram."
```

## Best practices for multi-agent cron jobs

| Tip | Recommendation |
|-----|----------------|
| **Name your Hands clearly** | Define them in `AGENTS.md` (e.g., `researcher`, `writer`, `news_curator`, `executor`) |
| **Use specific prompts** | Start the prompt with `"You are the [Hand Name] Hand..."` so the agent knows its role |
| **Limit tools per job** | Use `--allowed-tool` to restrict what each scheduled agent can do |
| **Avoid overlapping heavy jobs** | Stagger schedules (e.g., Researcher at :00, Writer at :15) to control costs |
| **Use `session_target`** | For advanced setups where you want completely isolated agent instances |

## How to check and manage scheduled jobs

```bash
zeroclaw cron list
zeroclaw cron status
```

Pause, resume, or remove specific jobs:

```bash
zeroclaw cron pause researcher-daily
zeroclaw cron resume news-curator-morning
```

## Summary

**Yes** — MiroClaw supports **per-agent cron scheduling** for separate, specialized Hands.

You can run, for example:

- Agent A (Researcher) every 6 hours  
- Agent B (News Curator) daily at 8:00  
- Agent C (Weekly Reporter) every Monday at 9:00  

All while keeping them as **separate, specialized Hands** defined in your configuration.
