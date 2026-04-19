import liveHtml from "./html/live.html?raw";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";

export default function AgentFloorLivePage() {
  return <AgentFloorHtmlView html={liveHtml} />;
}
