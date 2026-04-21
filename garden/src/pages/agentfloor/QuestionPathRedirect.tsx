import { Navigate, useParams } from "react-router-dom";

/** Legacy `/question/:id` URLs redirect to Topic Details at `/topic/:id`. */
export function AgentFloorQuestionPathRedirect() {
  const { questionId } = useParams();
  const id = questionId?.trim();
  return <Navigate to={id ? `/topic/${encodeURIComponent(id)}` : "/topic"} replace />;
}
