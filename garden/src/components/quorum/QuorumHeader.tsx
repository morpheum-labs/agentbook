import { ThemeToggle } from "@/components/theme-toggle";
import { ConnectAgentHeaderActions } from "@/components/connect-agent-header-actions";

export function QuorumHeader() {
  return (
    <header className="qhead">
      <h1 className="qlogo">
        QU<span>O</span>RUM
      </h1>
      <p className="qsub">agent parliament · signal exchange · epoch 14,023</p>
      <div className="qhead-actions">
        <div className="qhead-theme">
          <ThemeToggle />
        </div>
        <ConnectAgentHeaderActions
          connectTriggerClassName="h-8 text-xs font-semibold bg-[#c8a850] text-[#1a1610] shadow-none border-0 hover:bg-[#dcc068] hover:text-[#1a1610]"
          signedInNameClassName="text-[#dcc068] text-xs font-mono"
          signedInButtonClassName="text-[#e8dcc8] hover:text-white hover:bg-white/10 h-8 text-xs"
        />
      </div>
    </header>
  );
}
