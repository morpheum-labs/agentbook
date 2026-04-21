import { Link, Navigate, useParams } from "react-router-dom";
import {
  getResearchArticleBySlug,
  researchPageModel,
} from "./agentfloorResearchModel";

function topicPath(questionId: string): string {
  return `/topic/${encodeURIComponent(questionId)}`;
}

export default function AgentFloorResearchArticlePage() {
  const { slug: slugParam } = useParams();
  const slug = slugParam ? decodeURIComponent(slugParam.trim()) : "";
  const article = slug ? getResearchArticleBySlug(slug) : undefined;

  if (!article) {
    return <Navigate to="/research" replace />;
  }

  return (
    <article className="af-research-article" aria-labelledby="af-research-article-title">
      <header className="af-research-article-head">
        <div className="af-research-article-head-l">
          <Link to="/research" className="af-research-article-back">
            ← Research
          </Link>
          <p className="af-research-article-eyebrow">Signal brief · Desk article</p>
        </div>
        <p className="af-research-article-edition">{researchPageModel.editionLabel}</p>
      </header>

      <p className="af-research-article-kicker">{article.sectionLabel}</p>
      <h1 id="af-research-article-title" className="af-research-article-title">
        {article.headline}
      </h1>
      <p className="af-research-article-dek">{article.dek}</p>
      <p className="af-research-article-meta">{article.metaLine}</p>

      <div className="af-research-article-body">
        {article.body.map((p, i) => (
          <p key={i} className="af-research-article-p">
            {p}
          </p>
        ))}
      </div>

      <footer className="af-research-article-foot">
        <p className="af-research-article-foot-note">
          This is a Research desk write-up, not the live topic card.{" "}
          <Link className="af-research-article-floor-link" to={topicPath(article.questionId)}>
            Open {article.questionId} for positions and thread →
          </Link>
        </p>
      </footer>
    </article>
  );
}
