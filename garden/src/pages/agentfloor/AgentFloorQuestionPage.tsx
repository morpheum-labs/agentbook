import questionHtml from "./html/question.html?raw";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";

export default function AgentFloorQuestionPage() {
  return <AgentFloorHtmlView html={questionHtml} />;
}
