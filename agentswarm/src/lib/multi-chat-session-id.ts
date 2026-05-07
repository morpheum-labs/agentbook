/** Readable segment for ZeroClaw `session_id` (miroclaw stores history under `gw_<session_id>`). */
export function handSlug(name: string): string {
  const s = name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 48);
  return s.length > 0 ? s : "hand";
}

export function buildMultiChatSessionId(
  sessionInstanceKey: string,
  agent: { ID: string; Name: string }
): string {
  return `agentswarm:${sessionInstanceKey}:${handSlug(agent.Name)}:${agent.ID}`;
}
