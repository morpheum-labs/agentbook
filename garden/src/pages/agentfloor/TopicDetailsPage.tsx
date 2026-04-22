import { useEffect, useMemo, useState } from "react";
import { Navigate, useParams, useSearchParams } from "react-router-dom";
import { floorApi } from "@/lib/api";
import { AgentFloorHtmlView } from "./HtmlView";
import {
  buildRegionalDetailHtml,
  defaultRegionalDetailPageModel,
  parseRegionalFiltersFromSearchParams,
  regionalDetailPageFromApiPayload,
  type RegionalDetailPageModel,
} from "./agentfloorRegionalDetailModel";
import {
  buildTopicDetailsHtml,
  topicDetailsPageFromApi,
  type TopicDetailsPageModel,
} from "./agentfloorTopicDetailsModel";

type TopicLoadState = "loading" | "ready" | "error" | "not_found";

function isNotFoundMessage(msg: string): boolean {
  const m = msg.toLowerCase();
  return (
    m.includes("404") ||
    m.includes("not found") ||
    m.includes("question not found") ||
    m.includes("not_found")
  );
}

export default function AgentFloorTopicDetailsPage() {
  const { questionId } = useParams();
  const [searchParams] = useSearchParams();
  const id = questionId?.trim() || "Q.01";
  const regionalMode = searchParams.get("view") === "regional";
  const filters = useMemo(
    () => parseRegionalFiltersFromSearchParams(searchParams),
    [searchParams],
  );
  const queryKey = searchParams.toString();

  const [apiRegional, setApiRegional] = useState<RegionalDetailPageModel | null>(null);

  const [topicModel, setTopicModel] = useState<TopicDetailsPageModel | null>(null);
  const [topicLoad, setTopicLoad] = useState<TopicLoadState>(() => (regionalMode ? "ready" : "loading"));

  useEffect(() => {
    if (!regionalMode) {
      setApiRegional(null);
      return;
    }
    let cancelled = false;
    const apiQs = new URLSearchParams(searchParams);
    apiQs.delete("view");
    const suffix = apiQs.toString();
    void floorApi
      .getTopicRegional(id, suffix)
      .then((raw) => {
        if (cancelled) return;
        const parsed = regionalDetailPageFromApiPayload(raw as Record<string, unknown>);
        if (parsed && parsed.rows.length > 0) {
          setApiRegional(parsed);
        } else {
          setApiRegional(null);
        }
      })
      .catch(() => {
        if (!cancelled) setApiRegional(null);
      });
    return () => {
      cancelled = true;
    };
  }, [id, regionalMode, queryKey, searchParams]);

  useEffect(() => {
    if (regionalMode) {
      setTopicLoad("ready");
      return;
    }
    let cancelled = false;
    setTopicLoad("loading");
    setTopicModel(null);
    const run = async () => {
      try {
        const [qRaw, posRaw, histRaw] = await Promise.all([
          floorApi.getTopicDetails(id, "include=digest"),
          floorApi.listQuestionPositions(id, "limit=40").catch(() => [] as Record<string, unknown>[]),
          floorApi.listTopicDigestHistory(id, "limit=8").catch(() => [] as Record<string, unknown>[]),
        ]);
        if (cancelled) return;
        const model = topicDetailsPageFromApi(qRaw, posRaw, histRaw);
        if (!model) {
          setTopicModel(null);
          setTopicLoad("not_found");
          return;
        }
        setTopicModel(model);
        setTopicLoad("ready");
      } catch (e: unknown) {
        if (cancelled) return;
        const msg = e instanceof Error ? e.message : "";
        setTopicModel(null);
        setTopicLoad(isNotFoundMessage(msg) ? "not_found" : "error");
      }
    };
    void run();
    return () => {
      cancelled = true;
    };
  }, [id, regionalMode]);

  const html = useMemo(() => {
    if (regionalMode) {
      const model = apiRegional ?? defaultRegionalDetailPageModel(id, filters);
      return buildRegionalDetailHtml(model);
    }
    if (topicModel) return buildTopicDetailsHtml(topicModel);
    return "";
  }, [id, regionalMode, apiRegional, filters, topicModel]);

  if (!regionalMode && topicLoad === "loading") {
    return (
      <div className="q-wrap" role="status" aria-busy="true">
        <p className="q-lower-muted" style={{ padding: "2rem 1rem" }}>
          Loading topic…
        </p>
      </div>
    );
  }

  if (!regionalMode && (topicLoad === "not_found" || topicLoad === "error")) {
    return <Navigate to="/topics" replace />;
  }

  return (
    <AgentFloorHtmlView
      html={html}
      worldMonitorQuestionId={regionalMode || topicLoad !== "ready" || !topicModel ? undefined : id}
    />
  );
}
