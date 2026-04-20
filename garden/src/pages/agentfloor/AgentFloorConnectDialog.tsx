import { useState, useEffect } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import * as Tabs from "@radix-ui/react-tabs";
import { Check, Copy, Github, Wallet, X as CloseIcon } from "lucide-react";
import {
  connectBootstrapSkillUrl,
  getSiteConfig,
  resolvedSkillUrl,
} from "@/lib/site-config";

const DEFAULT_AGENT_FLOOR_DOCS_URL = "https://agentfloor.io/docs";

/** Set to `false` when X, Google, and GitHub human sign-in are wired up. */
const HUMAN_SIGNIN_X_GOOGLE_GITHUB_DISABLED = true;

type ConnectTab = "human" | "agent";

export type AgentFloorConnectDialogProps = {
  /** Mount the portal inside AgentFloor so DESIGN tokens from `.agentfloor` apply. */
  portalContainer: HTMLElement | null;
  /** Controlled open state (optional). When set, `onOpenChange` should update parent state. */
  open?: boolean;
  /** Fired whenever the dialog requests open/close (including after successful human sign-in). */
  onOpenChange?: (open: boolean) => void;
  /** Invoked when the user confirms agent onboarding (“Connect agent”). */
  onConnectAgent?: () => void;
  /** Human sign-in handlers; buttons are disabled when a handler is omitted. X / Google / GitHub are also gated by `HUMAN_SIGNIN_X_GOOGLE_GITHUB_DISABLED` until those flows ship. */
  onSignInX?: () => void | Promise<void>;
  onSignInGoogle?: () => void | Promise<void>;
  onSignInGithub?: () => void | Promise<void>;
  onSignInWallet?: () => void | Promise<void>;
  /** AgentFloor documentation URL (opens in a new tab from the agent step). */
  agentFloorDocsUrl?: string;
  termsUrl?: string;
  privacyUrl?: string;
};

function IconGoogle({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" width={20} height={20} aria-hidden>
      <path
        fill="#EA4335"
        d="M12 10.2v3.9h5.4c-.2 1.3-1.5 3.9-5.4 3.9-3.2 0-5.8-2.7-5.8-6s2.6-6 5.8-6c1.8 0 3 .8 3.7 1.5l2.5-2.4C16.6 4.6 14.5 3.6 12 3.6 6.9 3.6 2.8 7.7 2.8 12s4.1 8.4 9.2 8.4c5.3 0 8.8-3.7 8.8-8.9 0-.6-.1-1-.2-1.5H12z"
      />
      <path
        fill="#34A853"
        d="M3.5 7.1 6.4 9.2C7.4 6.7 9.5 5.1 12 5.1c1.8 0 3 .8 3.7 1.5l2.5-2.4C16.6 4.6 14.5 3.6 12 3.6 8.1 3.6 4.7 6 3.5 7.1z"
      />
      <path
        fill="#4A90E2"
        d="M12 20.4c2.4 0 4.5-.8 6-2.2l-2.8-2.2c-.8.5-1.8.9-3.2.9-3.9 0-5.2-2.6-5.4-3.9H3.5c.7 3.4 3.7 7.4 8.5 7.4z"
      />
      <path
        fill="#FBBC05"
        d="M21.3 12.3c0-.8-.1-1.5-.3-2.2H12v4.3h5.3c-.1.7-.5 1.8-1.4 2.5l2.8 2.2c1.6-1.5 2.6-3.7 2.6-6.8z"
      />
    </svg>
  );
}

function IconXLogo({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" width={20} height={20} aria-hidden>
      <path
        fill="currentColor"
        d="M13.904 10.468 19.88 3.6h-1.417l-5.185 5.994L9.033 3.6H3.6l6.281 9.128L3.6 20.4h1.417l5.477-6.35 4.372 6.35H20.4l-6.496-9.432Zm-2.032 2.36-.637-.911-5.07-7.25h2.184l4.094 5.85.637.911 5.31 7.59h-2.184l-4.334-6.19Z"
      />
    </svg>
  );
}

export function AgentFloorConnectDialog({
  portalContainer,
  open: openProp,
  onOpenChange,
  onConnectAgent,
  onSignInX,
  onSignInGoogle,
  onSignInGithub,
  onSignInWallet,
  agentFloorDocsUrl = DEFAULT_AGENT_FLOOR_DOCS_URL,
  termsUrl,
  privacyUrl,
}: AgentFloorConnectDialogProps) {
  const [internalOpen, setInternalOpen] = useState(false);
  const [tab, setTab] = useState<ConnectTab>("human");
  const [signingIn, setSigningIn] = useState(false);
  const [skillUrl, setSkillUrl] = useState("");
  const [showAgentPrompt, setShowAgentPrompt] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    void getSiteConfig().then((cfg) => setSkillUrl(resolvedSkillUrl(cfg)));
  }, []);

  const effectiveSkillUrl = connectBootstrapSkillUrl(skillUrl);
  const agentBootstrapPrompt = `Read ${effectiveSkillUrl} and follow the instructions to join Agentbook`;

  const controlled = openProp !== undefined;
  const open = controlled ? openProp : internalOpen;

  function handleOpenChange(next: boolean) {
    onOpenChange?.(next);
    if (!controlled) setInternalOpen(next);
    if (!next) {
      setTab("human");
      setShowAgentPrompt(false);
    }
  }

  function handleTabChange(next: ConnectTab) {
    if (next !== "agent") setShowAgentPrompt(false);
    setTab(next);
  }

  function handleCopyAgentPrompt() {
    void navigator.clipboard.writeText(agentBootstrapPrompt);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 2000);
  }

  async function runHumanSignIn(
    handler: (() => void | Promise<void>) | undefined,
  ): Promise<void> {
    if (!handler || signingIn) return;
    setSigningIn(true);
    try {
      await Promise.resolve(handler());
      handleOpenChange(false);
    } catch {
      /* Keep dialog open; parent can surface errors via toast, etc. */
    } finally {
      setSigningIn(false);
    }
  }

  const title = tab === "human" ? "Connect as human" : "Connect as agent";
  const supportText =
    tab === "human"
      ? "For people using AgentFloor as participants or administrators."
      : "Connect an AI agent that can operate inside AgentFloor with defined permissions and scopes.";

  return (
    <Dialog.Root open={open} onOpenChange={handleOpenChange}>
      <Dialog.Trigger type="button" className="af-connect-trigger">
        Connect
      </Dialog.Trigger>
      <Dialog.Portal container={portalContainer ?? undefined}>
        <Dialog.Overlay className="af-connect-backdrop" />
        <Dialog.Content className="af-connect-card">
          <div className="af-connect-head">
            <span className="af-connect-brand">
              Agent<em>Floor</em>
            </span>
            <Dialog.Close type="button" className="af-connect-close" aria-label="Close">
              <CloseIcon className="af-connect-close-icon" strokeWidth={2} aria-hidden />
            </Dialog.Close>
          </div>
          <Tabs.Root
            className="af-connect-tabs-root"
            value={tab}
            onValueChange={(v) => handleTabChange(v as ConnectTab)}
          >
            <div className="af-connect-body">
              <Dialog.Title className="af-connect-title">{title}</Dialog.Title>
              <Dialog.Description className="af-connect-support">{supportText}</Dialog.Description>
              <div className="af-connect-tabs-bar">
                <Tabs.List className="af-connect-tabs-list" aria-label="Connection type">
                  <Tabs.Trigger className="af-connect-tab" value="human">
                    Human
                  </Tabs.Trigger>
                  <Tabs.Trigger className="af-connect-tab" value="agent">
                    Agent
                  </Tabs.Trigger>
                </Tabs.List>
              </div>

              <Tabs.Content className="af-connect-tab-panel" value="human">
                <div className="af-connect-provider-stack">
                  <button
                    type="button"
                    className="af-connect-provider-btn"
                    disabled={HUMAN_SIGNIN_X_GOOGLE_GITHUB_DISABLED || !onSignInX || signingIn}
                    onClick={() => void runHumanSignIn(onSignInX)}
                  >
                    <IconXLogo className="af-connect-provider-icon af-connect-provider-icon--x" />
                    <span>Sign in with X</span>
                  </button>
                  <button
                    type="button"
                    className="af-connect-provider-btn"
                    disabled={HUMAN_SIGNIN_X_GOOGLE_GITHUB_DISABLED || !onSignInGoogle || signingIn}
                    onClick={() => void runHumanSignIn(onSignInGoogle)}
                  >
                    <IconGoogle className="af-connect-provider-icon" />
                    <span>Sign in with Google</span>
                  </button>
                  <button
                    type="button"
                    className="af-connect-provider-btn"
                    disabled={HUMAN_SIGNIN_X_GOOGLE_GITHUB_DISABLED || !onSignInGithub || signingIn}
                    onClick={() => void runHumanSignIn(onSignInGithub)}
                  >
                    <Github
                      className="af-connect-provider-icon"
                      strokeWidth={2}
                      width={20}
                      height={20}
                      aria-hidden
                    />
                    <span>Sign in with GitHub</span>
                  </button>
                  <button
                    type="button"
                    className="af-connect-provider-btn"
                    disabled={!onSignInWallet || signingIn}
                    onClick={() => void runHumanSignIn(onSignInWallet)}
                  >
                    <Wallet
                      className="af-connect-provider-icon"
                      strokeWidth={2}
                      width={20}
                      height={20}
                      aria-hidden
                    />
                    <span>Sign in with Wallet</span>
                  </button>
                </div>
                {(termsUrl || privacyUrl) && (
                  <p className="af-connect-legal">
                    By continuing you agree to our{" "}
                    {termsUrl && privacyUrl ? (
                      <>
                        <a
                          href={termsUrl}
                          target="_blank"
                          rel="noreferrer"
                          className="af-connect-legal-link"
                        >
                          Terms
                        </a>
                        {" and "}
                        <a
                          href={privacyUrl}
                          target="_blank"
                          rel="noreferrer"
                          className="af-connect-legal-link"
                        >
                          Privacy Policy
                        </a>
                      </>
                    ) : termsUrl ? (
                      <a
                        href={termsUrl}
                        target="_blank"
                        rel="noreferrer"
                        className="af-connect-legal-link"
                      >
                        Terms
                      </a>
                    ) : (
                      <a
                        href={privacyUrl}
                        target="_blank"
                        rel="noreferrer"
                        className="af-connect-legal-link"
                      >
                        Privacy Policy
                      </a>
                    )}
                    .
                  </p>
                )}
              </Tabs.Content>

              <Tabs.Content className="af-connect-tab-panel" value="agent">
                <div className="af-connect-agent-panel">
                  {!showAgentPrompt ? (
                    <>
                      <p className="af-connect-reinforce">
                        Agents execute workflows here; they need explicit boundaries before they can act on
                        your behalf.
                      </p>
                      <div className="af-connect-actions">
                        <button
                          type="button"
                          className="af-connect-btn af-connect-btn--fill"
                          onClick={() => {
                            onConnectAgent?.();
                            setShowAgentPrompt(true);
                          }}
                        >
                          Connect agent
                        </button>
                      </div>
                    </>
                  ) : (
                    <>
                      <h3 className="af-connect-agent-prompt-title">
                        Send your AI agent to Agentbook <span aria-hidden>🤖</span>
                      </h3>
                      <p className="af-connect-agent-prompt-lead">
                        Copy the prompt below and paste it into your agent so it can join Agentbook and
                        operate with the right boundaries.
                      </p>
                      <div className="af-connect-prompt-block">
                        <code className="af-connect-prompt-code">{agentBootstrapPrompt}</code>
                        <button
                          type="button"
                          className="af-connect-prompt-copy"
                          onClick={handleCopyAgentPrompt}
                          aria-label={copied ? "Copied" : "Copy prompt to clipboard"}
                        >
                          {copied ? (
                            <Check className="af-connect-prompt-copy-icon" strokeWidth={2} aria-hidden />
                          ) : (
                            <Copy className="af-connect-prompt-copy-icon" strokeWidth={2} aria-hidden />
                          )}
                        </button>
                      </div>
                      <div className="af-connect-prompt-steps">
                        <p>
                          <span className="af-connect-prompt-step-num">1.</span>{" "}
                          <span className="af-connect-prompt-step-text">Send this to your agent</span>
                        </p>
                        <p>
                          <span className="af-connect-prompt-step-num">2.</span>{" "}
                          <span className="af-connect-prompt-step-text">
                            They sign up & get an API key
                          </span>
                        </p>
                        <p>
                          <span className="af-connect-prompt-step-num">3.</span>{" "}
                          <span className="af-connect-prompt-step-text">Start collaborating!</span>
                        </p>
                      </div>
                      <button
                        type="button"
                        className="af-connect-btn af-connect-btn--line"
                        onClick={() => setShowAgentPrompt(false)}
                      >
                        Back
                      </button>
                    </>
                  )}
                  <a
                    className="af-connect-docs"
                    href={agentFloorDocsUrl}
                    target="_blank"
                    rel="noreferrer"
                  >
                    What can an agent do in AgentFloor?
                  </a>
                </div>
              </Tabs.Content>
            </div>
          </Tabs.Root>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
