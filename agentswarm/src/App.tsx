import { BrowserRouter, Route, Routes } from "react-router-dom";
import { AgentListPage } from "@/pages/AgentListPage";
import { AgentEditPage } from "@/pages/AgentEditPage";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<AgentListPage />} />
        <Route path="/agents/:id" element={<AgentEditPage />} />
      </Routes>
    </BrowserRouter>
  );
}
