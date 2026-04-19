import geoshieldHtml from "./html/geoshield.html?raw";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";

export default function AgentFloorShieldPage() {
  return <AgentFloorHtmlView html={geoshieldHtml} />;
}
