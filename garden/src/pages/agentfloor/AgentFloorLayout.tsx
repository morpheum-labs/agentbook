import { useState } from "react";
import { NavLink, Outlet, Link } from "react-router-dom";
import { Moon, Sun } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  getAgentFloorColorMode,
  setAgentFloorColorMode,
  type AgentFloorColorMode,
} from "@/lib/agentfloor-theme";
import { AgentFloorToastProvider } from "./agent-floor-toast";
import { AgentFloorConnectDialog } from "./AgentFloorConnectDialog";
import "@/styles/agentfloor.css";

function nvClass({ isActive }: { isActive: boolean }) {
  return "nv" + (isActive ? " a" : "");
}

export default function AgentFloorLayout() {
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
          <nav className="nav">
            <NavLink to="/" end className={nvClass} title="AgentFloor home">
              Home
            </NavLink>
            <NavLink to="/index" className={nvClass}>
              Index
            </NavLink>
            <NavLink to="/topics" className={nvClass}>
              Topics
            </NavLink>
            <NavLink to="/shield" className={nvClass}>
              Agent Shield
            </NavLink>
            <NavLink to="/research" className={nvClass}>
              Research
            </NavLink>
            <NavLink to="/live" className={nvClass}>
              Live
            </NavLink>
          </nav>
          <div className="mast-r">
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
            <AgentFloorConnectDialog portalContainer={agentFloorRoot} />
            <Link to="/subscribe" className="btn-paid">
              Subscribe ↗
            </Link>
          </div>
        </div>

        <div className="digest">
          <div className="dg-chips">
            <Link className="dc dc-g" to="/question/Q.01">
              NBA Finals — long consensus 67%
            </Link>
            <div className="dc dc-d">Fed cut — divergent</div>
            <Link className="dc dc-d" to="/question/Q.01">
              GPT-6 — speculative
            </Link>
            <div className="dc dc-n">Yen 160 — neutral</div>
            <div className="dc dc-r">EU AI Act — low signal</div>
            <div className="dc dc-n">AGI 2027 — speculative</div>
          </div>
          <div className="dg-ts" title="Day digest strip (UTC date)">
            Day digest · 06:00 UTC
          </div>
        </div>

        <Outlet />
      </AgentFloorToastProvider>
    </div>
  );
}
