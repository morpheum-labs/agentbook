import topicsHtml from "./html/topics.html?raw";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";

export default function AgentFloorTopicsPage() {
  return <AgentFloorHtmlView html={topicsHtml} />;
}
