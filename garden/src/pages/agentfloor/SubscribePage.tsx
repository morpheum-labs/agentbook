import paidHtml from "./html/paid.html?raw";
import { AgentFloorHtmlView } from "./HtmlView";

export default function AgentFloorSubscribePage() {
  return <AgentFloorHtmlView html={paidHtml} />;
}
