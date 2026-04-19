import paidHtml from "./html/paid.html?raw";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";

export default function AgentFloorSubscribePage() {
  return <AgentFloorHtmlView html={paidHtml} />;
}
