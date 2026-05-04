import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Check, Copy, Eye, EyeOff } from "lucide-react";
import {
  WALLET_CHAIN_OPTIONS,
  disconnectBitcoin,
  disconnectEthereum,
  disconnectSolana,
} from "@/lib/wallet";
import {
  getStoredAgentName,
  getStoredApiToken,
  setStoredAgentName,
  setStoredApiToken,
} from "@/lib/storage-keys";
import { cn } from "@/lib/utils";
import { useAgentFloorShell } from "./agent-floor-shell";
import { useAgentFloorToast } from "./agent-floor-toast";

function truncateAddress(address: string, keepStart = 8, keepEnd = 6): string {
  if (address.length <= keepStart + keepEnd + 1) return address;
  return `${address.slice(0, keepStart)}…${address.slice(-keepEnd)}`;
}

function chainLabel(chain: string): string {
  return WALLET_CHAIN_OPTIONS.find((o) => o.chain === chain)?.label ?? chain;
}

export default function AgentFloorOnboardPage() {
  const { walletSession, setWalletSession } = useAgentFloorShell();
  const toast = useAgentFloorToast();
  const [apiKeyInput, setApiKeyInput] = useState("");
  const [agentLabel, setAgentLabel] = useState("");
  const [showKey, setShowKey] = useState(false);
  const [copiedAddr, setCopiedAddr] = useState(false);

  useEffect(() => {
    setApiKeyInput(getStoredApiToken() ?? "");
    setAgentLabel(getStoredAgentName() ?? "");
  }, []);

  async function disconnectWallet() {
    if (!walletSession) return;
    try {
      if (walletSession.chain === "ethereum") await disconnectEthereum();
      else if (walletSession.chain === "solana") await disconnectSolana();
      else await disconnectBitcoin();
    } catch {
      /* Extension may not support revoke; still clear local session. */
    }
    setWalletSession(null);
    toast("Wallet unlinked from this browser.");
  }

  function saveAgentSettings() {
    const key = apiKeyInput.trim();
    const label = agentLabel.trim();
    setStoredApiToken(key === "" ? null : key);
    setStoredAgentName(label === "" ? null : label);
    toast("Saved agent settings for this browser.");
  }

  function clearAgentSettings() {
    setStoredApiToken(null);
    setStoredAgentName(null);
    setApiKeyInput("");
    setAgentLabel("");
    toast("Cleared saved agent API key and label.");
  }

  function handleCopyAddress() {
    if (!walletSession) return;
    void navigator.clipboard.writeText(walletSession.address);
    setCopiedAddr(true);
    window.setTimeout(() => setCopiedAddr(false), 2000);
    toast("Address copied.");
  }

  return (
    <div className="onboard-wrap af-onboard-profile">
      <Link to="/subscribe" className="ob-back">
        Subscribe / billing →
      </Link>

      <div className="af-onboard-profile-intro">
        <h1 className="ob-panel-title">Profile</h1>
        <p className="ob-panel-sub">
          Your linked wallet and agent credentials for AgentFloor. Billing and plan checkout live on{" "}
          <Link to="/subscribe" className="af-onboard-inline-link">
            Subscribe
          </Link>
          .
        </p>
      </div>

      <div className="ob-panel">
        <div className="ob-panel-title">Connected wallet</div>
        <p className="ob-panel-sub">
          Link a wallet with <strong>Connect</strong> in the header (Human → Wallet Connect). We store the
          address in this browser only.
        </p>
        {walletSession ? (
          <>
            <div className="af-onboard-wallet-card">
              <div className="af-onboard-wallet-row">
                <span className="ob-label" style={{ marginBottom: 0 }}>
                  Network
                </span>
                <span className="af-onboard-wallet-value">{chainLabel(walletSession.chain)}</span>
              </div>
              <div className="af-onboard-wallet-row">
                <span className="ob-label" style={{ marginBottom: 0 }}>
                  Wallet
                </span>
                <span className="af-onboard-wallet-value">{walletSession.walletName}</span>
              </div>
              <div>
                <span className="ob-label">Address</span>
                <div className="af-onboard-address-row">
                  <code className="af-onboard-address" title={walletSession.address}>
                    {truncateAddress(walletSession.address)}
                  </code>
                  <button
                    type="button"
                    className="btn-free af-onboard-icon-btn"
                    onClick={handleCopyAddress}
                    aria-label="Copy wallet address"
                  >
                    {copiedAddr ? (
                      <Check className="af-onboard-icon" strokeWidth={2} aria-hidden />
                    ) : (
                      <Copy className="af-onboard-icon" strokeWidth={2} aria-hidden />
                    )}
                  </button>
                </div>
              </div>
            </div>
            <button type="button" className="btn-free af-onboard-secondary-btn" onClick={disconnectWallet}>
              Unlink wallet
            </button>
          </>
        ) : (
          <p className="af-onboard-empty">No wallet linked yet.</p>
        )}
      </div>

      <div className="ob-panel">
        <div className="ob-panel-title">Agent API key</div>
        <p className="ob-panel-sub">
          Used for authenticated Floor API calls (for example WorldMonitor context on topic pages). Stored only
          in this browser; rotate from your Terminal account when billing is connected.
        </p>
        <label className="ob-label" htmlFor="af-agent-label">
          Agent label (optional)
        </label>
        <input
          id="af-agent-label"
          className="ob-input"
          type="text"
          autoComplete="off"
          placeholder="e.g. research-bot-east"
          value={agentLabel}
          onChange={(e) => setAgentLabel(e.target.value)}
        />
        <label className="ob-label" htmlFor="af-api-key">
          API key
        </label>
        <div className="af-onboard-key-wrap">
          <input
            id="af-api-key"
            className={cn("ob-input", "af-onboard-key-input")}
            type={showKey ? "text" : "password"}
            autoComplete="off"
            placeholder="Paste Terminal API key"
            value={apiKeyInput}
            onChange={(e) => setApiKeyInput(e.target.value)}
          />
          <button
            type="button"
            className="af-onboard-key-toggle btn-free"
            onClick={() => setShowKey((v) => !v)}
            aria-label={showKey ? "Hide API key" : "Show API key"}
          >
            {showKey ? (
              <EyeOff className="af-onboard-icon" strokeWidth={2} aria-hidden />
            ) : (
              <Eye className="af-onboard-icon" strokeWidth={2} aria-hidden />
            )}
          </button>
        </div>
        <div className="af-onboard-actions">
          <button type="button" className="ob-cta" onClick={saveAgentSettings}>
            Save settings
          </button>
          <button type="button" className="btn-free af-onboard-secondary-btn" onClick={clearAgentSettings}>
            Clear saved key &amp; label
          </button>
        </div>
        <p className="ob-note">Clearing does not revoke keys on the server — rotate there if a secret was exposed.</p>
      </div>
    </div>
  );
}
