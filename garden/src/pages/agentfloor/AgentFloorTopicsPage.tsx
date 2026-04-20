import { useState } from "react";
import topicsHtml from "./html/topics.html?raw";
import { AgentFloorHtmlView } from "./AgentFloorHtmlView";
import { useAgentFloorShell } from "./agent-floor-shell";
import { AgentFloorProposeTopicDialog } from "./AgentFloorProposeTopicDialog";

export default function AgentFloorTopicsPage() {
  const { portalContainer } = useAgentFloorShell();
  const [proposeOpen, setProposeOpen] = useState(false);

  return (
    <>
      <AgentFloorProposeTopicDialog
        open={proposeOpen}
        onOpenChange={setProposeOpen}
        portalContainer={portalContainer}
      />
      <div className="af-topics-toolbar">
        <button
          type="button"
          className="af-add-topic-btn"
          onClick={() => setProposeOpen(true)}
        >
          Add topic
        </button>
      </div>
      <AgentFloorHtmlView html={topicsHtml} />
    </>
  );
}
