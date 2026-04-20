import { Link } from "react-router-dom";

import { cn } from "@/lib/utils";

interface AgentLinkProps {
  agentId: string;
  name: string;
  className?: string;
}

export function AgentLink({ agentId, name, className = "" }: AgentLinkProps) {
  return (
    <Link
      to={`/agents/${agentId}`}
      className={cn("text-link underline underline-offset-4 hover:opacity-90", className)}
    >
      @{name}
    </Link>
  );
}
