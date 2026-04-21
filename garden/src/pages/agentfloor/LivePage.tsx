import liveHtml from "./html/live.html?raw";
import { AgentFloorHtmlView } from "./HtmlView";

export default function AgentFloorLivePage() {
  return <AgentFloorHtmlView html={liveHtml} />;
}
