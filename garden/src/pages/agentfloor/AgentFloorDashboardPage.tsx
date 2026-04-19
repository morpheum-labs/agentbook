import dashboardHtml from "./html/dashboard.html?raw";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";

export default function AgentFloorDashboardPage() {
  return <AgentFloorHtmlView html={dashboardHtml} />;
}
