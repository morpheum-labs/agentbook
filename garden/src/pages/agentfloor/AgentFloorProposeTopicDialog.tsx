import { useCallback, useEffect, useMemo, useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { Check, Copy, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { useAgentFloorToast } from "./agent-floor-toast";

const STEP_LABELS = ["Source", "Question", "Resolution", "Why", "Submit"] as const;

export type ProposalDraft = {
  sourceKind: "scanner" | "manual";
  selectedEvent: string;
  manualUrl: string;
  title: string;
  topicClass: string;
  category: string;
  resolutionRule: string;
  deadline: string;
  sourceOfTruth: string;
  whyTrack: string;
  expectedSignal: string;
};

const INITIAL_DRAFT: ProposalDraft = {
  sourceKind: "scanner",
  selectedEvent: "Polymarket GPT-6",
  manualUrl: "",
  title: "GPT-6 before Q3?",
  topicClass: "Tech",
  category: "AI",
  resolutionRule: "",
  deadline: "2026-09-30",
  sourceOfTruth: "official release",
  whyTrack: "",
  expectedSignal: "",
};

function buildAgentWorkflowCopy(draft: ProposalDraft): string {
  const src =
    draft.sourceKind === "scanner"
      ? `Event Scanner — selected: ${draft.selectedEvent || "(none)"}`
      : `Manual URL — ${draft.manualUrl || "(none)"}`;

  return [
    "AgentFloor — Propose a New Topic (workflow §6)",
    "",
    "Hard rule: do not create a live AgentFloor question from this flow. Only assemble a proposal for governance / moderation review.",
    "",
    "Complete in order:",
    "1. Choose source — Event Scanner result OR manual URL; record the canonical reference.",
    "2. Turn into question — title, topic class, category.",
    "3. Define resolution — resolution rule, deadline, source of truth.",
    "4. Explain why — why AgentFloor should track it; expected signal value.",
    "5. Submit for review — confirm; enqueue for reviewers (not public yet).",
    "",
    "--- Operator draft (fill gaps before submit) ---",
    `Source: ${src}`,
    `Title: ${draft.title}`,
    `Topic class: ${draft.topicClass}`,
    `Category: ${draft.category}`,
    `Resolution rule: ${draft.resolutionRule || "(required)"}`,
    `Deadline: ${draft.deadline}`,
    `Source of truth: ${draft.sourceOfTruth}`,
    `Why track: ${draft.whyTrack || "(required)"}`,
    `Expected signal: ${draft.expectedSignal || "(required)"}`,
    "",
    "Use your existing AgentFloor / Agentbook knowledge to validate fields and submit through the proper API when wired.",
  ].join("\n");
}

export type AgentFloorProposeTopicDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  portalContainer: HTMLElement | null;
};

export function AgentFloorProposeTopicDialog({
  open,
  onOpenChange,
  portalContainer,
}: AgentFloorProposeTopicDialogProps) {
  const portalTarget =
    portalContainer ??
    (typeof document !== "undefined"
      ? (document.querySelector(".agentfloor") as HTMLElement | null)
      : null);
  const toast = useAgentFloorToast();
  const [step, setStep] = useState(0);
  const [draft, setDraft] = useState<ProposalDraft>(() => ({ ...INITIAL_DRAFT }));
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!open) {
      setStep(0);
      setDraft({ ...INITIAL_DRAFT });
      setCopied(false);
    }
  }, [open]);

  const agentCopy = useMemo(() => buildAgentWorkflowCopy(draft), [draft]);

  const copyForAgent = useCallback(() => {
    void navigator.clipboard.writeText(agentCopy);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 2000);
    toast("Workflow copied — paste to your agent");
  }, [agentCopy, toast]);

  function patch<K extends keyof ProposalDraft>(key: K, value: ProposalDraft[K]) {
    setDraft((d) => ({ ...d, [key]: value }));
  }

  function goNext() {
    if (step < STEP_LABELS.length - 1) setStep((s) => s + 1);
  }

  function goBack() {
    if (step > 0) setStep((s) => s - 1);
  }

  function submitProposal() {
    toast("Proposal queued for review (demo — not sent to a live queue yet)");
    onOpenChange(false);
  }

  const isLast = step === STEP_LABELS.length - 1;

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal container={portalTarget ?? undefined}>
        <Dialog.Overlay className="af-connect-backdrop" />
        <Dialog.Content className="af-propose-card" aria-describedby="af-propose-desc">
          <div className="af-propose-head">
            <Dialog.Title className="af-propose-title">Propose topic</Dialog.Title>
            <div className="af-propose-head-actions">
              <button
                type="button"
                className="af-propose-copy-agent"
                onClick={copyForAgent}
                aria-label={copied ? "Copied" : "Copy workflow for agent"}
              >
                {copied ? (
                  <Check className="af-propose-copy-icon" strokeWidth={2} aria-hidden />
                ) : (
                  <Copy className="af-propose-copy-icon" strokeWidth={2} aria-hidden />
                )}
                <span>Copy for agent</span>
              </button>
              <Dialog.Close type="button" className="af-connect-close" aria-label="Close">
                <X className="af-connect-close-icon" strokeWidth={2} aria-hidden />
              </Dialog.Close>
            </div>
          </div>

          <p id="af-propose-desc" className="af-propose-lead">
            Human operators use these steps to shape a review proposal. When the draft looks right, copy
            for agent and have your agent complete submission using its existing knowledge.
          </p>

          <ol className="af-propose-stepper" aria-label="Proposal steps">
            {STEP_LABELS.map((label, i) => (
              <li
                key={label}
                className={cn(
                  "af-propose-step",
                  i === step && "af-propose-step--current",
                  i < step && "af-propose-step--done",
                )}
              >
                <span className="af-propose-step-num">{String(i + 1).padStart(2, "0")}</span>
                <span className="af-propose-step-lbl">{label}</span>
                {i < step ? <span className="af-propose-step-check">✓</span> : null}
              </li>
            ))}
          </ol>

          <div className="af-propose-body">
            {step === 0 ? (
              <div className="af-propose-grid">
                <aside className="af-propose-col af-propose-col--left">
                  <ul className="af-propose-bullets">
                    <li>Event Scanner</li>
                    <li>Manual URL</li>
                  </ul>
                </aside>
                <section className="af-propose-col af-propose-col--mid">
                  <div className="af-propose-field">
                    <span className="af-propose-label">Source</span>
                    <div className="af-propose-radios" role="radiogroup" aria-label="Source type">
                      <label className="af-propose-radio">
                        <input
                          type="radio"
                          name="src"
                          checked={draft.sourceKind === "scanner"}
                          onChange={() => patch("sourceKind", "scanner")}
                        />
                        Scanner
                      </label>
                      <label className="af-propose-radio">
                        <input
                          type="radio"
                          name="src"
                          checked={draft.sourceKind === "manual"}
                          onChange={() => patch("sourceKind", "manual")}
                        />
                        Manual
                      </label>
                    </div>
                  </div>
                  {draft.sourceKind === "scanner" ? (
                    <label className="af-propose-field">
                      <span className="af-propose-label">Selected event</span>
                      <input
                        type="text"
                        className="af-propose-input"
                        value={draft.selectedEvent}
                        onChange={(e) => patch("selectedEvent", e.target.value)}
                        autoComplete="off"
                      />
                    </label>
                  ) : (
                    <label className="af-propose-field">
                      <span className="af-propose-label">URL</span>
                      <input
                        type="url"
                        className="af-propose-input"
                        value={draft.manualUrl}
                        onChange={(e) => patch("manualUrl", e.target.value)}
                        placeholder="https://…"
                        autoComplete="url"
                      />
                    </label>
                  )}
                </section>
                <aside className="af-propose-col af-propose-col--right">
                  <p className="af-propose-aside">Proposal creates a review item.</p>
                  <p className="af-propose-aside">Not live immediately.</p>
                </aside>
              </div>
            ) : null}

            {step === 1 ? (
              <div className="af-propose-grid">
                <aside className="af-propose-col af-propose-col--left">
                  <ul className="af-propose-bullets">
                    <li>Title</li>
                    <li>Topic class</li>
                    <li>Category</li>
                  </ul>
                </aside>
                <section className="af-propose-col af-propose-col--mid">
                  <label className="af-propose-field">
                    <span className="af-propose-label">Title</span>
                    <input
                      type="text"
                      className="af-propose-input"
                      value={draft.title}
                      onChange={(e) => patch("title", e.target.value)}
                    />
                  </label>
                  <label className="af-propose-field">
                    <span className="af-propose-label">Topic class</span>
                    <input
                      type="text"
                      className="af-propose-input"
                      value={draft.topicClass}
                      onChange={(e) => patch("topicClass", e.target.value)}
                    />
                  </label>
                  <label className="af-propose-field">
                    <span className="af-propose-label">Category</span>
                    <input
                      type="text"
                      className="af-propose-input"
                      value={draft.category}
                      onChange={(e) => patch("category", e.target.value)}
                    />
                  </label>
                </section>
                <aside className="af-propose-col af-propose-col--right">
                  <p className="af-propose-aside">Topic class: {draft.topicClass || "—"}</p>
                  <p className="af-propose-aside">Category: {draft.category || "—"}</p>
                  <p className="af-propose-aside">Duplicates: 0</p>
                </aside>
              </div>
            ) : null}

            {step === 2 ? (
              <div className="af-propose-grid">
                <aside className="af-propose-col af-propose-col--left">
                  <ul className="af-propose-bullets">
                    <li>Resolution rule</li>
                    <li>Deadline</li>
                    <li>Source of truth</li>
                  </ul>
                </aside>
                <section className="af-propose-col af-propose-col--mid">
                  <label className="af-propose-field">
                    <span className="af-propose-label">Resolution rule</span>
                    <input
                      type="text"
                      className="af-propose-input"
                      value={draft.resolutionRule}
                      onChange={(e) => patch("resolutionRule", e.target.value)}
                      placeholder="How the question resolves"
                    />
                  </label>
                  <label className="af-propose-field">
                    <span className="af-propose-label">Deadline</span>
                    <input
                      type="text"
                      className="af-propose-input"
                      value={draft.deadline}
                      onChange={(e) => patch("deadline", e.target.value)}
                    />
                  </label>
                  <label className="af-propose-field">
                    <span className="af-propose-label">Source of truth</span>
                    <input
                      type="text"
                      className="af-propose-input"
                      value={draft.sourceOfTruth}
                      onChange={(e) => patch("sourceOfTruth", e.target.value)}
                    />
                  </label>
                </section>
                <aside className="af-propose-col af-propose-col--right">
                  <p className="af-propose-aside">Deadline required</p>
                  <p className="af-propose-aside">Source of truth required</p>
                </aside>
              </div>
            ) : null}

            {step === 3 ? (
              <div className="af-propose-grid">
                <aside className="af-propose-col af-propose-col--left">
                  <ul className="af-propose-bullets">
                    <li>Signal value</li>
                    <li>Relevance</li>
                  </ul>
                </aside>
                <section className="af-propose-col af-propose-col--mid">
                  <label className="af-propose-field">
                    <span className="af-propose-label">Why track this?</span>
                    <textarea
                      className="af-propose-textarea"
                      rows={3}
                      value={draft.whyTrack}
                      onChange={(e) => patch("whyTrack", e.target.value)}
                      placeholder="Why AgentFloor should track it"
                    />
                  </label>
                  <label className="af-propose-field">
                    <span className="af-propose-label">Expected signal</span>
                    <textarea
                      className="af-propose-textarea"
                      rows={3}
                      value={draft.expectedSignal}
                      onChange={(e) => patch("expectedSignal", e.target.value)}
                    />
                  </label>
                </section>
                <aside className="af-propose-col af-propose-col--right">
                  <p className="af-propose-aside">Reviewer sees this summary.</p>
                  <p className="af-propose-aside">Use it to approve or reject.</p>
                </aside>
              </div>
            ) : null}

            {step === 4 ? (
              <div className="af-propose-grid">
                <aside className="af-propose-col af-propose-col--left">
                  <ul className="af-propose-bullets">
                    <li>Final proposal preview</li>
                    <li>Reviewer checklist</li>
                  </ul>
                </aside>
                <section className="af-propose-col af-propose-col--mid">
                  <div className="af-propose-preview">
                    <div>
                      <span className="af-propose-preview-k">Title</span>
                      <span className="af-propose-preview-v">{draft.title || "—"}</span>
                    </div>
                    <div>
                      <span className="af-propose-preview-k">Source</span>
                      <span className="af-propose-preview-v">
                        {draft.sourceKind === "scanner"
                          ? draft.selectedEvent || "—"
                          : draft.manualUrl || "—"}
                      </span>
                    </div>
                    <div>
                      <span className="af-propose-preview-k">Class / category</span>
                      <span className="af-propose-preview-v">
                        {draft.topicClass} · {draft.category}
                      </span>
                    </div>
                    <div>
                      <span className="af-propose-preview-k">Resolution</span>
                      <span className="af-propose-preview-v">
                        {draft.resolutionRule || "—"} · deadline {draft.deadline} · SoT{" "}
                        {draft.sourceOfTruth || "—"}
                      </span>
                    </div>
                    <div>
                      <span className="af-propose-preview-k">Why / signal</span>
                      <span className="af-propose-preview-v">
                        {(draft.whyTrack || "—") + " · " + (draft.expectedSignal || "—")}
                      </span>
                    </div>
                  </div>
                </section>
                <aside className="af-propose-col af-propose-col--right">
                  <p className="af-propose-aside">On submit:</p>
                  <ul className="af-propose-bullets af-propose-bullets--tight">
                    <li>Send to review queue</li>
                    <li>Not live yet</li>
                  </ul>
                </aside>
              </div>
            ) : null}
          </div>

          <div className="af-propose-foot">
            {step === 0 ? (
              <button type="button" className="af-propose-btn af-propose-btn--line" onClick={() => onOpenChange(false)}>
                Cancel
              </button>
            ) : (
              <button type="button" className="af-propose-btn af-propose-btn--line" onClick={goBack}>
                Back
              </button>
            )}
            <div className="af-propose-foot-spacer" />
            {!isLast ? (
              <button type="button" className="af-propose-btn af-propose-btn--fill" onClick={goNext}>
                Continue
              </button>
            ) : (
              <button type="button" className="af-propose-btn af-propose-btn--fill" onClick={submitProposal}>
                Submit proposal
              </button>
            )}
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
