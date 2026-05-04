import type { WalletConnectedSession } from "@/lib/wallet";

const KEY = "agentfloor_wallet_session";

function isSession(o: unknown): o is WalletConnectedSession {
  if (!o || typeof o !== "object") return false;
  const r = o as Record<string, unknown>;
  const chain = r.chain;
  if (chain !== "ethereum" && chain !== "solana" && chain !== "bitcoin") return false;
  return typeof r.address === "string" && r.address.length > 0 && typeof r.walletName === "string";
}

export function getAgentFloorWalletSession(): WalletConnectedSession | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return null;
    const parsed: unknown = JSON.parse(raw);
    return isSession(parsed) ? parsed : null;
  } catch {
    return null;
  }
}

export function setAgentFloorWalletSession(session: WalletConnectedSession): void {
  localStorage.setItem(KEY, JSON.stringify(session));
}

export function clearAgentFloorWalletSession(): void {
  localStorage.removeItem(KEY);
}
