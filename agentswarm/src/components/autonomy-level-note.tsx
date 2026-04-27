type Props = { id: string };

/** Explains levels per gateway / miroclaw (ZeroClaw) session security: ReadOnly, Supervised, Full. */
export function AutonomyLevelNote({ id }: Props) {
  return (
    <p
      id={id}
      className="text-micro text-muted-foreground mt-1.5 max-w-prose leading-relaxed"
      role="note"
    >
      <span className="font-medium text-foreground/80">ReadOnly</span> — the agent can observe but not
      act. <span className="font-medium text-foreground/80">Supervised</span> (default in many setups)
      — the agent acts with approval for medium- or high-risk operations.{" "}
      <span className="font-medium text-foreground/80">Full</span> — the agent acts autonomously
      within policy bounds.
    </p>
  );
}
