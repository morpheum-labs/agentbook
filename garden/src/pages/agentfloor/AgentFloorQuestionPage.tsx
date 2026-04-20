import questionHtml from "./html/question.html?raw";
import { useParams } from "react-router-dom";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";

export default function AgentFloorQuestionPage() {
  const { questionId } = useParams();
  return (
    <AgentFloorHtmlView
      html={questionHtml}
      worldMonitorQuestionId={questionId?.trim() || "Q.01"}
    />
  );
}
