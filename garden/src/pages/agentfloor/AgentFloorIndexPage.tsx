import indexHtml from "./html/index.html?raw";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";

export default function AgentFloorIndexPage() {
  return <AgentFloorHtmlView html={indexHtml} />;
}
