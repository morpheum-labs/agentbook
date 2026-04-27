import { BrowserRouter, Route, Routes } from "react-router-dom";
import { AppPageShell } from "@/components/app-page-shell";
import { AgentListPage } from "@/pages/AgentListPage";
import { AgentEditPage } from "@/pages/AgentEditPage";
import { AgentNewPage } from "@/pages/AgentNewPage";
import { AgentChartPage } from "@/pages/AgentChartPage";
import { CronJobListPage } from "@/pages/CronJobListPage";
import { CronJobNewPage } from "@/pages/CronJobNewPage";
import { CronJobEditPage } from "@/pages/CronJobEditPage";

export default function App() {
  return (
    <BrowserRouter>
      <AppPageShell>
        <Routes>
          <Route path="/" element={<AgentListPage />} />
          <Route path="/chart" element={<AgentChartPage />} />
          <Route path="/agents/new" element={<AgentNewPage />} />
          <Route path="/agents/:id" element={<AgentEditPage />} />
          <Route path="/cron-jobs" element={<CronJobListPage />} />
          <Route path="/cron-jobs/new" element={<CronJobNewPage />} />
          <Route path="/cron-jobs/:id" element={<CronJobEditPage />} />
        </Routes>
      </AppPageShell>
    </BrowserRouter>
  );
}
