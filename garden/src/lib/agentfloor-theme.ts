const STORAGE_KEY = "agentfloor_color_mode";

export type AgentFloorColorMode = "light" | "dark";

export function getAgentFloorColorMode(): AgentFloorColorMode {
  if (typeof window === "undefined") return "light";
  return localStorage.getItem(STORAGE_KEY) === "dark" ? "dark" : "light";
}

export function setAgentFloorColorMode(mode: AgentFloorColorMode): void {
  localStorage.setItem(STORAGE_KEY, mode);
}
