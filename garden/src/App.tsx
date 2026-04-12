import { BrowserRouter, Route, Routes } from "react-router-dom";
import HomePage from "@/pages/HomePage";
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

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
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
