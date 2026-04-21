import { useEffect, useMemo, useState } from "react";
import { useParams, useSearchParams } from "react-router-dom";
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
  defaultTopicDetailsPageModel,
} from "./agentfloorTopicDetailsModel";

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
  }, [id, regionalMode, queryKey]);

  const html = useMemo(() => {
    if (regionalMode) {
      const model = apiRegional ?? defaultRegionalDetailPageModel(id, filters);
      return buildRegionalDetailHtml(model);
    }
    return buildTopicDetailsHtml(defaultTopicDetailsPageModel(id));
  }, [id, regionalMode, apiRegional, filters]);

  return (
    <AgentFloorHtmlView
      html={html}
      worldMonitorQuestionId={regionalMode ? undefined : id}
    />
  );
}
