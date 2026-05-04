import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import type { WalletConnectedSession } from "@/lib/wallet";
import {
  clearAgentFloorWalletSession,
  getAgentFloorWalletSession,
  setAgentFloorWalletSession,
} from "@/lib/agentfloor-wallet-session";

type AgentFloorShellValue = {
  portalContainer: HTMLElement | null;
  walletSession: WalletConnectedSession | null;
  setWalletSession: (session: WalletConnectedSession | null) => void;
};

const AgentFloorShellContext = createContext<AgentFloorShellValue>({
  portalContainer: null,
  walletSession: null,
  setWalletSession: () => {},
});

export function AgentFloorShellProvider({
  portalContainer,
  children,
}: {
  portalContainer: HTMLElement | null;
  children: ReactNode;
}) {
  const [walletSession, setWalletSessionState] = useState<WalletConnectedSession | null>(() =>
    typeof window !== "undefined" ? getAgentFloorWalletSession() : null,
  );

  const setWalletSession = useCallback((session: WalletConnectedSession | null) => {
    setWalletSessionState(session);
    if (typeof window === "undefined") return;
    if (session) setAgentFloorWalletSession(session);
    else clearAgentFloorWalletSession();
  }, []);

  const value = useMemo(
    () => ({ portalContainer, walletSession, setWalletSession }),
    [portalContainer, walletSession, setWalletSession],
  );

  return (
    <AgentFloorShellContext.Provider value={value}>{children}</AgentFloorShellContext.Provider>
  );
}

/** Portal target for Radix overlays scoped under `.agentfloor` tokens. */
// eslint-disable-next-line react-refresh/only-export-components -- hook colocated with thin Provider
export function useAgentFloorShell() {
  return useContext(AgentFloorShellContext);
}
