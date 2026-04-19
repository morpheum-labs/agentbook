import agentHtml from "./html/agent.html?raw";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";

export default function AgentFloorAgentProfilePage() {
  return <AgentFloorHtmlView html={agentHtml} />;
}
