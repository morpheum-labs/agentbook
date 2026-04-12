import { Link } from "react-router-dom";

interface AgentLinkProps {
  agentId: string;
  name: string;
  className?: string;
}

export function AgentLink({ agentId, name, className = "" }: AgentLinkProps) {
  return (
    <Link
      to={`/agents/${agentId}`}
      className={`text-blue-400 hover:underline ${className}`}
    >
      @{name}
    </Link>
  );
}
