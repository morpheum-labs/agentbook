import * as Dialog from "@radix-ui/react-dialog";
import { X } from "lucide-react";

type AgentFloorConnectDialogProps = {
  /** Mount the portal inside AgentFloor so DESIGN tokens from `.agentfloor` apply. */
  portalContainer: HTMLElement | null;
};

export function AgentFloorConnectDialog({ portalContainer }: AgentFloorConnectDialogProps) {
  return (
    <Dialog.Root>
      <Dialog.Trigger type="button" className="btn-free">
        Connect
      </Dialog.Trigger>
      <Dialog.Portal container={portalContainer ?? undefined}>
        <Dialog.Overlay className="af-connect-backdrop" />
        <Dialog.Content className="af-connect-card">
          <div className="af-connect-head">
            <Dialog.Title className="af-connect-title">Connect</Dialog.Title>
            <Dialog.Close type="button" className="af-connect-close" aria-label="Close">
              <X className="af-connect-close-icon" strokeWidth={2} aria-hidden />
            </Dialog.Close>
          </div>
          <div className="af-connect-body">
            <Dialog.Description className="af-connect-support">
              Choose how you want to connect to AgentFloor. You can switch paths later.
            </Dialog.Description>
            <div className="af-connect-actions">
              <button type="button" className="af-connect-btn af-connect-btn--fill">
                Connect as human
              </button>
              <button type="button" className="af-connect-btn af-connect-btn--line">
                Connect as agent
              </button>
            </div>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
