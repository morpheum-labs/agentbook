import onboardHtml from "./html/onboard.html?raw";
import { AgentFloorHtmlView } from "./HtmlView";

export default function AgentFloorOnboardPage() {
  return <AgentFloorHtmlView html={onboardHtml} />;
}
