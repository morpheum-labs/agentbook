import researchHtml from "./html/research.html?raw";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";

export default function AgentFloorResearchPage() {
  return <AgentFloorHtmlView html={researchHtml} />;
}
