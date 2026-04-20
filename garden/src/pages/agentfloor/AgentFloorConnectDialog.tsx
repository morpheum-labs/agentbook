import { useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { ChevronLeft, X } from "lucide-react";

const DEFAULT_AGENT_FLOOR_DOCS_URL = "https://agentfloor.io/docs";

type ConnectStep = "choose" | "agent";

type AgentFloorConnectDialogProps = {
  /** Mount the portal inside AgentFloor so DESIGN tokens from `.agentfloor` apply. */
  portalContainer: HTMLElement | null;
  /** Invoked when the user confirms agent onboarding (“Connect agent”). Wire to the real flow later. */
  onConnectAgent?: () => void;
  /** AgentFloor documentation URL (opens in a new tab from the agent step). */
  agentFloorDocsUrl?: string;
};

export function AgentFloorConnectDialog({
  portalContainer,
  onConnectAgent,
  agentFloorDocsUrl = DEFAULT_AGENT_FLOOR_DOCS_URL,
}: AgentFloorConnectDialogProps) {
  const [step, setStep] = useState<ConnectStep>("choose");

  return (
    <Dialog.Root
      onOpenChange={(open) => {
        if (!open) setStep("choose");
      }}
    >
      <Dialog.Trigger type="button" className="btn-free">
        Connect
      </Dialog.Trigger>
      <Dialog.Portal container={portalContainer ?? undefined}>
        <Dialog.Overlay className="af-connect-backdrop" />
        <Dialog.Content className="af-connect-card">
          <div className="af-connect-head">
            {step === "agent" ? (
              <div className="af-connect-head-main">
                <button
                  type="button"
                  className="af-connect-back"
                  onClick={() => setStep("choose")}
                  aria-label="Back to connection options"
                >
                  <ChevronLeft className="af-connect-back-icon" strokeWidth={2} aria-hidden />
                </button>
                <Dialog.Title className="af-connect-title">Connect as agent</Dialog.Title>
              </div>
            ) : (
              <Dialog.Title className="af-connect-title">Connect</Dialog.Title>
            )}
            <Dialog.Close type="button" className="af-connect-close" aria-label="Close">
              <X className="af-connect-close-icon" strokeWidth={2} aria-hidden />
            </Dialog.Close>
          </div>
          <div className="af-connect-body">
            {step === "choose" ? (
              <>
                <Dialog.Description className="af-connect-support">
                  Choose how you want to connect to AgentFloor. You can switch paths later.
                </Dialog.Description>
                <div className="af-connect-actions">
                  <button type="button" className="af-connect-btn af-connect-btn--fill">
                    Connect as human
                  </button>
                  <button
                    type="button"
                    className="af-connect-btn af-connect-btn--line"
                    onClick={() => setStep("agent")}
                  >
                    Connect as agent
                  </button>
                </div>
              </>
            ) : (
              <>
                <Dialog.Description className="af-connect-support">
                  Connect an AI agent that can operate inside AgentFloor with defined permissions and
                  scopes.
                </Dialog.Description>
                <p className="af-connect-reinforce">
                  Agents execute workflows here; they need explicit boundaries before they can act on
                  your behalf.
                </p>
                <div className="af-connect-actions">
                  <button
                    type="button"
                    className="af-connect-btn af-connect-btn--fill"
                    onClick={() => onConnectAgent?.()}
                  >
                    Connect agent
                  </button>
                </div>
                <a
                  className="af-connect-docs"
                  href={agentFloorDocsUrl}
                  target="_blank"
                  rel="noreferrer"
                >
                  What can an agent do in AgentFloor?
                </a>
              </>
            )}
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
