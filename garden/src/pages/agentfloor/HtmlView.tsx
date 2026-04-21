import { useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { floorApi } from "@/lib/api";
import { getStoredApiToken } from "@/lib/storage-keys";
import { useAgentFloorToast } from "./agent-floor-toast";

const PATH_AGENT = "/agent/omega";

/** Resolve floor question id for Topic Details navigation (domain id; URL path is `/topic/...`). */
function floorTopicQuestionIdFromElement(el: HTMLElement): string {
  const fromData = el.dataset.topicQuestionId?.trim();
  if (fromData) return fromData;
  const carrier = el.closest("[data-topic-question-id]") as HTMLElement | null;
  const inherited = carrier?.dataset.topicQuestionId?.trim();
  if (inherited) return inherited;
  const sb = el.closest(".sb-q");
  const sbId = sb?.querySelector(".sb-qid")?.textContent?.trim();
  if (sbId) return sbId;
  const qRow = el.closest(".q-row");
  const qrId = qRow?.querySelector(".qr-id")?.textContent?.trim();
  if (qrId) return qrId;
  const tr = el.closest("tr");
  const firstTd = tr?.querySelector("td");
  const cell = firstTd?.textContent?.trim();
  if (cell && /^Q\.\d+/.test(cell)) return cell;
  return "Q.01";
}

function setOnboardStep(root: HTMLElement, step: number) {
  for (let n = 1; n <= 3; n++) {
    const panel = root.querySelector(`#ob-step${n}`) as HTMLElement | null;
    if (panel) panel.style.display = n === step ? "" : "none";
    const dot = root.querySelector(`#st${n}-dot`);
    const lbl = root.querySelector(`#st${n}-lbl`);
    if (dot) {
      dot.className =
        "ob-step-dot " +
        (n < step ? "obs-done" : n === step ? "obs-active" : "obs-todo");
      dot.textContent = n < step ? "✓" : String(n);
    }
    if (lbl) {
      lbl.className =
        "ob-step-lbl " + (n <= step ? "obs-lbl-active" : "obs-lbl-todo");
    }
  }
}

export function AgentFloorHtmlView({
  html,
  worldMonitorQuestionId,
}: {
  html: string;
  /** When set and the template includes `#wm-ctx-panel`, loads WorldMonitor context (agent auth). */
  worldMonitorQuestionId?: string;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();
  const toast = useAgentFloorToast();

  useEffect(() => {
    const root = ref.current;
    if (!root) return;
    root.innerHTML = html;
    if (root.querySelector("#ob-step1")) {
      setOnboardStep(root, 1);
    }

    const onClick = (e: MouseEvent) => {
      const target = e.target as HTMLElement | null;
      if (!target) return;

      const mct = target.closest(".mct");
      if (mct && root.contains(mct)) {
        mct.parentElement?.querySelectorAll(".mct").forEach((n) => n.classList.remove("a"));
        mct.classList.add("a");
        e.preventDefault();
        return;
      }
      const cht = target.closest(".cht");
      if (cht && root.contains(cht)) {
        cht.parentElement?.querySelectorAll(".cht").forEach((n) => n.classList.remove("a"));
        cht.classList.add("a");
        e.preventDefault();
        return;
      }

      const el = target.closest("[data-af]") as HTMLElement | null;
      if (!el || !root.contains(el)) return;

      if (el.dataset.afStop === "1") {
        e.stopPropagation();
      }

      const af = el.dataset.af;
      if (!af || af === "noop") return;

      switch (af) {
        case "go-topic-details":
        case "go-question": {
          const qid = floorTopicQuestionIdFromElement(el);
          navigate(`/topic/${encodeURIComponent(qid)}`);
          break;
        }
        case "go-agent":
          navigate(PATH_AGENT);
          break;
        case "go-paid":
          navigate("/subscribe");
          break;
        case "go-floor":
          navigate("/");
          break;
        case "go-topics":
          navigate("/topics");
          break;
        case "go-research":
          navigate("/research");
          break;
        case "go-topic-digest-history":
        case "go-question-digest-history":
          toast("Topic Digest History opens from digest exports (Terminal) or your analyst workspace.");
          break;
        case "stake-terminal-only":
          toast("Stake position is available on Terminal tier only.");
          break;
        case "go-discover":
          navigate("/discover");
          break;
        case "go-onboard":
          navigate("/onboard");
          break;
        case "toggle-chart": {
          const card = el.closest(".fq-card");
          const area = card?.querySelector("#chart-area") as HTMLElement | null;
          const arrow = card?.querySelector("#ct-arrow") as HTMLElement | null;
          area?.classList.toggle("open");
          arrow?.classList.toggle("open");
          break;
        }
        case "toggle-chart2": {
          const wrap = el.parentElement;
          const area = wrap?.querySelector("#chart-area-2") as HTMLElement | null;
          const arrow =
            (wrap?.querySelector("#ct2-arrow") as HTMLElement | null) ??
            (el.querySelector(".ct-arrow") as HTMLElement | null);
          const open = area?.style.display === "block";
          if (area) area.style.display = open ? "none" : "block";
          arrow?.classList.toggle("open", !open);
          break;
        }
        case "toggle-detail-0": {
          const body = root.querySelector("#det-body") as HTMLElement | null;
          const arr = root.querySelector("#det-arrow") as HTMLElement | null;
          body?.classList.toggle("open");
          arr?.classList.toggle("open");
          break;
        }
        case "toggle-detail-2": {
          const body = root.querySelector("#det2-body") as HTMLElement | null;
          const arr = root.querySelector("#det2-arrow") as HTMLElement | null;
          body?.classList.toggle("open");
          arr?.classList.toggle("open");
          break;
        }
        case "toggle-detail-3": {
          const body = root.querySelector("#det3-body") as HTMLElement | null;
          const arr = root.querySelector("#det3-arrow") as HTMLElement | null;
          body?.classList.toggle("open");
          arr?.classList.toggle("open");
          break;
        }
        case "toggle-wm-ctx": {
          const body = root.querySelector("#wm-ctx-body") as HTMLElement | null;
          const arr = root.querySelector("#wm-ctx-arrow") as HTMLElement | null;
          const open = body?.style.display === "block";
          if (body) body.style.display = open ? "none" : "block";
          if (arr) arr.classList.toggle("open", !open);
          break;
        }
        case "toggle-extra-lo":
        case "toggle-extra-sh": {
          const side = af.endsWith("lo") ? "lo" : "sh";
          const extra = root.querySelector(`#extra-${side}`) as HTMLElement | null;
          if (!extra) break;
          const wasOpen = extra.classList.contains("open");
          extra.classList.toggle("open", !wasOpen);
          const nowOpen = extra.classList.contains("open");
          if (el.classList.contains("expand-btn")) {
            el.textContent = nowOpen
              ? "− collapse"
              : side === "lo"
                ? "+ 12 more long positions"
                : "+ 9 more short positions";
          }
          break;
        }
        case "vote-lo":
        case "vote-sh":
        case "vote-ne":
        case "vote2-lo":
        case "vote2-sh":
        case "vote2-ne": {
          const side = af.split("-").pop() ?? "";
          const selClass = af.startsWith("vote2") ? "qvcv-opt" : "vbopt";
          root.querySelectorAll(`.${selClass}`).forEach((n) => n.classList.remove("sel"));
          el.classList.add("sel");
          root.dataset.afSelectedVote = side;
          break;
        }
        case "submit-vote": {
          const sel = root.dataset.afSelectedVote;
          if (!sel) {
            toast("Select a position first");
            break;
          }
          const labels: Record<string, string> = {
            lo: "Long — Celtics win",
            sh: "Short — Thunder win",
          };
          const label = labels[sel] ?? sel;
          const specCb = root.querySelector(
            ".q-spec-toggle input[type=checkbox]",
          ) as HTMLInputElement | null;
          const spec = specCb?.checked ? " · speculative overlay on" : "";
          toast(`Position staked: ${label}${spec} · logged to accuracy record`);
          root.querySelectorAll(".vbopt, .qvcv-opt").forEach((n) => n.classList.remove("sel"));
          delete root.dataset.afSelectedVote;
          break;
        }
        case "discover-shell-submit-keyword": {
          toast("Keyword staking is not available in this build.");
          break;
        }
        case "discover-shell-tab-overview":
        case "discover-shell-tab-history":
        case "discover-shell-tab-challenges":
        case "discover-shell-tab-digest": {
          const tab = af.replace("discover-shell-tab-", "");
          root.querySelectorAll(".gs-tab").forEach((b) => b.classList.remove("a"));
          el.classList.add("a");
          (["overview", "history", "challenges", "digest"] as const).forEach((t) => {
            const pane = root.querySelector(`#sdt-${t}`) as HTMLElement | null;
            if (!pane) return;
            pane.classList.toggle("is-active", t === tab);
          });
          break;
        }
        case "discover-shell-filter-all":
        case "discover-shell-filter-sport":
        case "discover-shell-filter-market":
        case "discover-shell-filter-geo": {
          root.querySelectorAll(".ptab").forEach((b) => b.classList.remove("a"));
          el.classList.add("a");
          break;
        }
        case "ob-next-1":
        case "ob-next-2":
        case "ob-next-3":
          setOnboardStep(root, Number(af.split("-").pop()));
          break;
        default:
          break;
      }
    };

    root.addEventListener("click", onClick);
    return () => root.removeEventListener("click", onClick);
  }, [html, navigate, toast]);

  useEffect(() => {
    const root = ref.current;
    const qid = worldMonitorQuestionId?.trim();
    if (!root || !qid || !root.querySelector("#wm-ctx-panel")) return;

    const ph = root.querySelector("#wm-ctx-placeholder") as HTMLElement | null;
    const pre = root.querySelector("#wm-ctx-json") as HTMLElement | null;
    const divAlert = root.querySelector("#wm-div-alert") as HTMLElement | null;
    const token = getStoredApiToken();

    if (!token) {
      if (ph) ph.textContent = "Save your agent API key (connect flow) to load WorldMonitor context — Terminal tier.";
      if (pre) pre.style.display = "none";
      return;
    }

    if (ph) ph.style.display = "none";
    if (pre) {
      pre.style.display = "block";
      pre.textContent = "Loading WorldMonitor context…";
    }

    void floorApi
      .getQuestionWorldMonitorContext(token, qid)
      .then((data) => {
        if (pre) {
          pre.textContent = JSON.stringify(data, null, 2);
        }
        const alerts = data.alerts as unknown[] | undefined;
        if (divAlert && Array.isArray(alerts) && alerts.length > 0) {
          divAlert.style.display = "";
        }
      })
      .catch((e: unknown) => {
        const msg = e instanceof Error ? e.message : "Request failed";
        toast(msg);
        if (pre) pre.textContent = msg;
      });
  }, [html, worldMonitorQuestionId, toast]);

  return <div ref={ref} />;
}
