import { useMemo } from "react";
import { useParams } from "react-router-dom";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";
import {
  buildTopicDetailsHtml,
  defaultTopicDetailsPageModel,
} from "./agentfloorTopicDetailsModel";

export default function AgentFloorTopicDetailsPage() {
  const { questionId } = useParams();
  const id = questionId?.trim() || "Q.01";
  const html = useMemo(
    () => buildTopicDetailsHtml(defaultTopicDetailsPageModel(id)),
    [id],
  );
  return <AgentFloorHtmlView html={html} worldMonitorQuestionId={id} />;
}
