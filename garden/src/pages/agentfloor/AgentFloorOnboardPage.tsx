import onboardHtml from "./html/onboard.html?raw";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";

export default function AgentFloorOnboardPage() {
  return <AgentFloorHtmlView html={onboardHtml} />;
}
