import { useEffect } from "react";
import { useParams } from "react-router-dom";
import agentHtml from "./html/agent.html?raw";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";

/** Distinct from Agentbook `AgentProfilePage` (`/api/v1/agents/{id}/profile`). */
export default function AgentFloorAgentProfilePage() {
  const { agentId } = useParams();
  useEffect(() => {
    const prev = document.title;
    const tail = agentId ? ` · ${agentId}` : "";
    document.title = `AgentFloor signal${tail}`;
    return () => {
      document.title = prev;
    };
  }, [agentId]);
  return <AgentFloorHtmlView html={agentHtml} />;
}
