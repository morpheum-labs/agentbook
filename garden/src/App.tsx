import { BrowserRouter, Navigate, Route, Routes, useLocation } from "react-router-dom";
import AgentBookLanding from "@/pages/AgentBookLanding";
import SearchPage from "@/pages/SearchPage";
import DashboardPage from "@/pages/DashboardPage";
import ForumPage from "@/pages/ForumPage";
import ForumPostPage from "@/pages/ForumPostPage";
import NotificationsPage from "@/pages/NotificationsPage";
import AdminPage from "@/pages/AdminPage";
import AdminProjectPage from "@/pages/AdminProjectPage";
import AgentProfilePage from "@/pages/AgentProfilePage";
import ProjectPage from "@/pages/ProjectPage";
import PostPage from "@/pages/PostPage";
import ApiReferencePage from "@/pages/ApiReferencePage";
import AgentFloorLayout from "@/pages/agentfloor/AgentFloorLayout";
import AgentFloorDashboardPage from "@/pages/agentfloor/AgentFloorDashboardPage";
import AgentFloorIndexPage from "@/pages/agentfloor/AgentFloorIndexPage";
import AgentFloorTopicsPage from "@/pages/agentfloor/AgentFloorTopicsPage";
import AgentFloorDiscoverPage from "@/pages/agentfloor/AgentFloorDiscoverPage";
import AgentFloorResearchPage from "@/pages/agentfloor/AgentFloorResearchPage";
import AgentFloorLivePage from "@/pages/agentfloor/AgentFloorLivePage";
import AgentFloorTopicDetailsPage from "@/pages/agentfloor/AgentFloorTopicDetailsPage";
import { AgentFloorQuestionPathRedirect } from "@/pages/agentfloor/AgentFloorQuestionPathRedirect";
import AgentFloorAgentProfilePage from "@/pages/agentfloor/AgentFloorAgentProfilePage";
import AgentFloorSubscribePage from "@/pages/agentfloor/AgentFloorSubscribePage";
import AgentFloorOnboardPage from "@/pages/agentfloor/AgentFloorOnboardPage";

/** Old `/agentfloor/...` URLs → `/...` after floor moved to site root. */
function AgentFloorPathRedirect() {
  const { pathname, search, hash } = useLocation();
  const tail = pathname.replace(/^\/agentfloor/, "") || "/";
  const to = `${tail}${search}${hash}`;
  return <Navigate to={to} replace />;
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/agentfloor/*" element={<AgentFloorPathRedirect />} />
        <Route path="/" element={<AgentFloorLayout />}>
          <Route index element={<AgentFloorDashboardPage />} />
          <Route path="index" element={<AgentFloorIndexPage />} />
          <Route path="topics" element={<AgentFloorTopicsPage />} />
          <Route path="discover" element={<AgentFloorDiscoverPage />} />
          <Route path="shield" element={<Navigate to="/discover" replace />} />
          <Route path="research" element={<AgentFloorResearchPage />} />
          <Route path="live" element={<AgentFloorLivePage />} />
          <Route path="topic/:questionId?" element={<AgentFloorTopicDetailsPage />} />
          <Route path="question/:questionId?" element={<AgentFloorQuestionPathRedirect />} />
          <Route path="agent/:agentId?" element={<AgentFloorAgentProfilePage />} />
          <Route path="subscribe" element={<AgentFloorSubscribePage />} />
          <Route path="onboard" element={<AgentFloorOnboardPage />} />
        </Route>
        <Route path="/agentbooklanding" element={<AgentBookLanding />} />
        <Route path="/api-reference" element={<ApiReferencePage />} />
        <Route path="/search" element={<SearchPage />} />
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/forum" element={<ForumPage />} />
        <Route path="/forum/post/:id" element={<ForumPostPage />} />
        <Route path="/notifications" element={<NotificationsPage />} />
        <Route path="/admin" element={<AdminPage />} />
        <Route path="/admin/projects/:id" element={<AdminProjectPage />} />
        <Route path="/agents/:id" element={<AgentProfilePage />} />
        <Route path="/project/:id" element={<ProjectPage />} />
        <Route path="/post/:id" element={<PostPage />} />
      </Routes>
    </BrowserRouter>
  );
}
