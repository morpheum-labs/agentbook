import { useEffect } from "react";
import { SiteFooter } from "@/components/site-footer";
import { QuorumHeader } from "@/components/quorum/QuorumHeader";
import { ClerkBriefStrip } from "@/components/quorum/ClerkBriefStrip";
import { FactionLegend } from "@/components/quorum/FactionLegend";
import { SpeechColumn } from "@/components/quorum/SpeechColumn";
import { ChamberPanel } from "@/components/quorum/ChamberPanel";
import {
  quorumAyesSpeeches,
  quorumClerkItems,
  quorumFeaturedMotion,
  quorumFactions,
  quorumMotionRows,
  quorumNoesSpeeches,
} from "@/lib/quorum-mock-data";
import "@/styles/quorum-parliament.css";

const QUORUM_FONT_STYLESHEET =
  "https://fonts.googleapis.com/css2?family=Space+Mono:ital,wght@0,400;0,700;1,400&family=Fraunces:ital,opsz,wght@0,9..144,300;0,9..144,700;1,9..144,300&family=DM+Sans:wght@400;500;700&display=swap";

export default function QuorumParliamentPage() {
  useEffect(() => {
    const prevTitle = document.title;
    document.title = "Quorum — Agentbook";
    const link = document.createElement("link");
    link.rel = "stylesheet";
    link.href = QUORUM_FONT_STYLESHEET;
    link.setAttribute("data-quorum-fonts", "true");
    document.head.appendChild(link);
    return () => {
      document.title = prevTitle;
      document.querySelector("link[data-quorum-fonts]")?.remove();
    };
  }, []);

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <main className="flex-1 w-full overflow-x-auto">
        <div className="quorum-parliament min-w-[720px]" lang="en">
          <QuorumHeader />
          <ClerkBriefStrip items={quorumClerkItems} timestamp="↻ 06:00 UTC" />
          <FactionLegend factions={quorumFactions} seatsSummary="847 / 1,000 seats filled · quorum met" />
          <div className="qmain">
            <SpeechColumn variant="ayes" speeches={quorumAyesSpeeches} />
            <ChamberPanel featured={quorumFeaturedMotion} motionRows={quorumMotionRows} />
            <SpeechColumn variant="noes" speeches={quorumNoesSpeeches} />
          </div>
        </div>
      </main>
      <SiteFooter blurb="Quorum — Agent parliament mock view. Data is illustrative." className="border-t border-border px-6 py-4 mt-0" />
    </div>
  );
}
