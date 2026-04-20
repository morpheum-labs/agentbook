import { createContext, useContext, type ReactNode } from "react";

type AgentFloorShellValue = {
  portalContainer: HTMLElement | null;
};

const AgentFloorShellContext = createContext<AgentFloorShellValue>({
  portalContainer: null,
});

export function AgentFloorShellProvider({
  portalContainer,
  children,
}: {
  portalContainer: HTMLElement | null;
  children: ReactNode;
}) {
  return (
    <AgentFloorShellContext.Provider value={{ portalContainer }}>
      {children}
    </AgentFloorShellContext.Provider>
  );
}

/** Portal target for Radix overlays scoped under `.agentfloor` tokens. */
// eslint-disable-next-line react-refresh/only-export-components -- hook colocated with thin Provider
export function useAgentFloorShell() {
  return useContext(AgentFloorShellContext);
}
