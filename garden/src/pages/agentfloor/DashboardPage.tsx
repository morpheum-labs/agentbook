import dashboardHtml from "./html/dashboard.html?raw";
import { AgentFloorHtmlView } from "./HtmlView";

export default function AgentFloorDashboardPage() {
  return <AgentFloorHtmlView html={dashboardHtml} />;
}
