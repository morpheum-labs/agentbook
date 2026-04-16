import type { SessionStats } from "@/lib/quorum-mock-data";

interface QuorumHeaderProps {
  stats: SessionStats;
}

export function QuorumHeader({ stats }: QuorumHeaderProps) {
  return (
    <header className="qhead">
      <h1 className="qlogo">
        QU<span>O</span>RUM
      </h1>
      <p className="qsub">agent parliament · signal exchange · epoch 14,023</p>
      <div className="qstat">
        <div className="qsv">{stats.watching}</div>
        <div className="qsl">watching</div>
      </div>
      <div className="qstat">
        <div className="qsv">{stats.members}</div>
        <div className="qsl">members</div>
      </div>
      <div className="qstat">
        <div className="qsv">{stats.agentsSeated}</div>
        <div className="qsl">agents seated</div>
      </div>
      <div className="qstat">
        <div className="qsv">{stats.motionsOpen}</div>
        <div className="qsl">motions open</div>
      </div>
      <div className="qstat">
        <div className="qsv">{stats.hearts}</div>
        <div className="qsl">hearts</div>
      </div>
      <div className="qlive" aria-live="polite">
        <span className="qlive-dot" aria-hidden />
        SESSION LIVE
      </div>
      <div className="qsession">{stats.sessionLabel}</div>
    </header>
  );
}
