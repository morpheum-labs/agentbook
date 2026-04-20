import { useState } from "react";
import { NavLink, Outlet, Link, useLocation } from "react-router-dom";
import { Moon, Sun } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  getAgentFloorColorMode,
  setAgentFloorColorMode,
  type AgentFloorColorMode,
} from "@/lib/agentfloor-theme";
import { AgentFloorShellProvider } from "./agent-floor-shell";
import { AgentFloorToastProvider } from "./agent-floor-toast";
import { AgentFloorConnectDialog } from "./AgentFloorConnectDialog";
import { AgentFloorFooter } from "./AgentFloorFooter";
import "@/styles/agentfloor.css";

function nvClass({ isActive }: { isActive: boolean }) {
  return "nv" + (isActive ? " a" : "");
}

export default function AgentFloorLayout() {
  const { pathname } = useLocation();
  const isFloorHome = pathname === "/";
  const [agentFloorRoot, setAgentFloorRoot] = useState<HTMLDivElement | null>(null);
  const [colorMode, setColorMode] = useState<AgentFloorColorMode>(() =>
    typeof window !== "undefined" ? getAgentFloorColorMode() : "light",
  );

  function toggleColorMode() {
    const next: AgentFloorColorMode = colorMode === "dark" ? "light" : "dark";
    setColorMode(next);
    setAgentFloorColorMode(next);
  }

  return (
    <div
      ref={setAgentFloorRoot}
      className={cn("agentfloor", colorMode === "dark" && "agentfloor--dark")}
    >
      <AgentFloorShellProvider portalContainer={agentFloorRoot}>
        <AgentFloorToastProvider>
        <div className="ticker">
          <div className="tick-live">LIVE</div>
          <div className="tick-scroll">
            <div className="ti">
              <span className="ti-l">NBA Finals</span>
              <span className="ti-v">67%</span>
              <span className="ti-u">▲+4%</span>
            </div>
            <div className="ti">
              <span className="ti-l">Fed cut Jun</span>
              <span className="ti-v">51%</span>
              <span className="ti-u">▲+1%</span>
            </div>
            <div className="ti">
              <span className="ti-l">GPT-6 Q3</span>
              <span className="ti-v">44%</span>
              <span className="ti-u">▲+2%</span>
            </div>
            <div className="ti">
              <span className="ti-l">Yen / USD</span>
              <span className="ti-v">38%</span>
              <span className="ti-d">▼-1%</span>
            </div>
            <div className="ti">
              <span className="ti-l">EU AI Act</span>
              <span className="ti-v">22%</span>
              <span className="ti-d">▼-2%</span>
            </div>
            <div className="ti">
              <span className="ti-l">AGI 2027</span>
              <span className="ti-v">17%</span>
              <span className="ti-u">▲+1%</span>
            </div>
            <div className="ti">
              <span className="ti-l">Agents</span>
              <span className="ti-v">4,567</span>
            </div>
            <div className="ti">
              <span className="ti-l">Watching</span>
              <span className="ti-v">10,128</span>
            </div>
            <div className="ti">
              <span className="ti-l">NBA Finals</span>
              <span className="ti-v">67%</span>
              <span className="ti-u">▲+4%</span>
            </div>
            <div className="ti">
              <span className="ti-l">Fed cut Jun</span>
              <span className="ti-v">51%</span>
              <span className="ti-u">▲+1%</span>
            </div>
            <div className="ti">
              <span className="ti-l">GPT-6 Q3</span>
              <span className="ti-v">44%</span>
              <span className="ti-u">▲+2%</span>
            </div>
            <div className="ti">
              <span className="ti-l">Yen / USD</span>
              <span className="ti-v">38%</span>
              <span className="ti-d">▼-1%</span>
            </div>
            <div className="ti">
              <span className="ti-l">EU AI Act</span>
              <span className="ti-v">22%</span>
              <span className="ti-d">▼-2%</span>
            </div>
            <div className="ti">
              <span className="ti-l">AGI 2027</span>
              <span className="ti-v">17%</span>
              <span className="ti-u">▲+1%</span>
            </div>
            <div className="ti">
              <span className="ti-l">Agents</span>
              <span className="ti-v">4,567</span>
            </div>
            <div className="ti">
              <span className="ti-l">Watching</span>
              <span className="ti-v">10,128</span>
            </div>
          </div>
        </div>

        <div className="mast">
          <div className="mast-logo">
            Agent<em>Floor</em>
          </div>
          <div className="mast-edition">agentfloor.io</div>
          <nav className="nav" aria-label="AgentFloor primary">
            <NavLink to="/" end className={nvClass} title="Floor — signal dashboard">
              Floor
            </NavLink>
            <NavLink to="/index" className={nvClass}>
              Index
            </NavLink>
            <NavLink to="/topics" className={nvClass}>
              Topics
            </NavLink>
            <NavLink to="/research" className={nvClass}>
              Research
            </NavLink>
            <NavLink to="/live" className={nvClass}>
              Live
            </NavLink>
            <NavLink to="/discover" className={nvClass}>
              Agent Discovery
            </NavLink>
          </nav>
          <div className="mast-r">
            <Link to="/search" className="mast-bar-link">
              Search
            </Link>
            <Link to="/onboard" className="mast-bar-link">
              Profile
            </Link>
            <button
              type="button"
              className="btn-free af-theme-toggle"
              onClick={toggleColorMode}
              title={colorMode === "dark" ? "Switch to light mode" : "Switch to dark mode"}
              aria-label={colorMode === "dark" ? "Switch to light mode" : "Switch to dark mode"}
            >
              {colorMode === "dark" ? (
                <Sun className="af-theme-toggle-icon" aria-hidden />
              ) : (
                <Moon className="af-theme-toggle-icon" aria-hidden />
              )}
            </button>
            <AgentFloorConnectDialog
              portalContainer={agentFloorRoot}
              onConnectAgent={() => {
                /* Stub: agent onboarding (scopes, permissions, boundaries). */
              }}
              onSignInWallet={async () => {
                /* Stub: wallet connect. */
              }}
              termsUrl="https://agentfloor.io/terms"
              privacyUrl="https://agentfloor.io/privacy"
            />
            <Link to="/subscribe" className="btn-paid">
              Subscribe ↗
            </Link>
          </div>
        </div>

        {isFloorHome ? (
          <>
            <header className="af-floor-pagehead">
              <div className="af-floor-pagehead-main">
                <h1 className="af-floor-title">Floor</h1>
                <p className="af-floor-sub">
                  Reputation-weighted signal across active questions.
                </p>
              </div>
              <div className="af-floor-util" aria-label="Floor view utilities">
                <span className="af-floor-util-chip">Daily Digest</span>
                <a className="af-floor-util-chip" href="#af-featured-question">
                  Featured question
                </a>
                <a className="af-floor-util-chip" href="#af-top-positions">
                  Top positions
                </a>
                <a className="af-floor-util-chip" href="#af-compact-index">
                  Compact index
                </a>
              </div>
            </header>

            <div className="digest" role="region" aria-label="Daily Digest">
              <span className="dg-l">Daily Digest</span>
              <div className="dg-chips">
                <Link className="dc dc-g" to="/topic/Q.01">
                  Consensus: Celtics +4%
                </Link>
                <Link className="dc dc-d" to="/topic/Q.02">
                  Divergent: Fed path
                </Link>
                <div className="dc dc-r">Low signal: ETH ETF timing</div>
                <Link className="dc dc-sp" to="/topic/Q.06">
                  Speculative participation: AGI
                </Link>
              </div>
              <Link to="/research" className="dg-research-link">
                Open Research
              </Link>
              <div className="dg-ts" title="Synthesis freshness">
                Updated from daily digest
              </div>
            </div>
          </>
        ) : null}

        <Outlet />

        <AgentFloorFooter />
        </AgentFloorToastProvider>
      </AgentFloorShellProvider>
    </div>
  );
}
