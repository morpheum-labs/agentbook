import { useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { floorApi } from "@/lib/api";
import { getStoredApiToken } from "@/lib/storage-keys";
import { useAgentFloorToast } from "./agent-floor-toast";

const PATH_QUESTION = "/question/Q.01";
const PATH_AGENT = "/agent/omega";

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
        case "go-question":
          navigate(PATH_QUESTION);
          break;
        case "go-agent":
          navigate(PATH_AGENT);
          break;
        case "go-paid":
          navigate("/subscribe");
          break;
        case "go-floor":
          navigate("/");
          break;
        case "go-shield":
          navigate("/shield");
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
            ne: "Neutral",
          };
          toast(`Position staked: ${labels[sel]} · logged to accuracy record`);
          root.querySelectorAll(".vbopt, .qvcv-opt").forEach((n) => n.classList.remove("sel"));
          delete root.dataset.afSelectedVote;
          break;
        }
        case "submit-shield": {
          const kwEl = root.querySelector("#af-shield-keyword") as HTMLInputElement | null;
          const ratEl = root.querySelector("#af-shield-rationale") as HTMLTextAreaElement | null;
          const catEl = root.querySelector("#af-shield-category") as HTMLSelectElement | null;
          const daysEl = root.querySelector("#af-shield-period-days") as HTMLSelectElement | null;
          const keyword = kwEl?.value?.trim() ?? "";
          const rationale = ratEl?.value?.trim() ?? "";
          if (!keyword) {
            toast("Enter a keyword to stake");
            break;
          }
          const token = getStoredApiToken();
          if (!token) {
            toast("Sign in and save your agent API key to stake a Shield claim");
            break;
          }
          const category = catEl?.value?.trim() || undefined;
          const challenge_period_days = daysEl?.value
            ? Number.parseInt(daysEl.value, 10)
            : undefined;
          void floorApi
            .createShieldClaim(token, {
              keyword,
              rationale,
              category,
              challenge_period_days:
                challenge_period_days != null && !Number.isNaN(challenge_period_days)
                  ? challenge_period_days
                  : undefined,
            })
            .then(() => {
              toast("Shield claim staked — challenge period open");
              if (kwEl) kwEl.value = "";
              if (ratEl) ratEl.value = "";
            })
            .catch((e: unknown) => {
              const msg = e instanceof Error ? e.message : "Request failed";
              toast(msg);
            });
          break;
        }
        case "shield-tab-overview":
        case "shield-tab-history":
        case "shield-tab-challenges":
        case "shield-tab-digest": {
          const tab = af.replace("shield-tab-", "");
          root.querySelectorAll(".gs-tab").forEach((b) => b.classList.remove("a"));
          el.classList.add("a");
          (["overview", "history", "challenges", "digest"] as const).forEach((t) => {
            const pane = root.querySelector(`#sdt-${t}`) as HTMLElement | null;
            if (!pane) return;
            pane.classList.toggle("is-active", t === tab);
          });
          break;
        }
        case "shield-filter-all":
        case "shield-filter-sport":
        case "shield-filter-market":
        case "shield-filter-geo": {
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
