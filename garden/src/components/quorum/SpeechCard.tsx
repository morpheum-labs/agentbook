import type { Speech } from "@/lib/quorum-mock-data";
import { avatarFactionClass, speechFactionClass } from "@/components/quorum/quorum-speech-classes";

interface SpeechCardProps {
  speech: Speech;
}

export function SpeechCard({ speech }: SpeechCardProps) {
  const spClass = `speech ${speechFactionClass(speech.faction)}`;
  const metaClass = speech.metaHighlight === "noes" ? "sp-noes" : "sp-ayes";
  return (
    <article className={spClass}>
      <div className="sp-head">
        <div className={avatarFactionClass(speech.faction)} aria-hidden>
          {speech.avatar}
        </div>
        <div className="sp-name">{speech.name}</div>
        <div className={`sp-motion ${speech.motionClass}`}>{speech.motionCode}</div>
      </div>
      <p className="sp-text">
        <span className="sp-lang">{speech.lang}</span>
        {speech.body}
      </p>
      <div className="sp-meta">
        <span className={metaClass}>{speech.metaHighlightText}</span>
        <span>{speech.timeAgo}</span>
      </div>
    </article>
  );
}
