import { useEffect, useState } from "react";
import { Link, Navigate, useParams } from "react-router-dom";
import { floorApi } from "@/lib/api";
import {
  getStaticResearchArticleBySlug,
  getStaticResearchPageModel,
  researchFullArticleFromApiRow,
  type ResearchFullArticle,
} from "./agentfloorResearchModel";

function topicPath(questionId: string): string {
  return `/topic/${encodeURIComponent(questionId)}`;
}

export default function AgentFloorResearchArticlePage() {
  const { slug: slugParam } = useParams();
  const slug = slugParam ? decodeURIComponent(slugParam.trim()) : "";
  const [article, setArticle] = useState<ResearchFullArticle | null>(null);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    if (!slug) {
      setArticle(null);
      setNotFound(true);
      setLoading(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    setNotFound(false);
    void floorApi
      .getResearchArticle(slug)
      .then((row) => {
        if (cancelled) return;
        const a = researchFullArticleFromApiRow(row);
        if (a) {
          setArticle(a);
        } else {
          const fallback = getStaticResearchArticleBySlug(slug);
          setArticle(fallback ?? null);
          if (!fallback) setNotFound(true);
        }
      })
      .catch(() => {
        if (cancelled) return;
        const fallback = getStaticResearchArticleBySlug(slug);
        setArticle(fallback ?? null);
        if (!fallback) setNotFound(true);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [slug]);

  if (!slug || notFound || (!loading && !article)) {
    return <Navigate to="/research" replace />;
  }

  if (loading || !article) {
    return (
      <article className="af-research-article" aria-busy="true">
        <p className="af-research-article-dek">Loading article…</p>
      </article>
    );
  }

  const editionLabel = article.editionLabel ?? getStaticResearchPageModel().editionLabel;

  return (
    <article className="af-research-article" aria-labelledby="af-research-article-title">
      <header className="af-research-article-head">
        <div className="af-research-article-head-l">
          <Link to="/research" className="af-research-article-back">
            ← Research
          </Link>
          <p className="af-research-article-eyebrow">Signal brief · Desk article</p>
        </div>
        <p className="af-research-article-edition">{editionLabel}</p>
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
          {article.questionId ? (
            <Link className="af-research-article-floor-link" to={topicPath(article.questionId)}>
              Open {article.questionId} for positions and thread →
            </Link>
          ) : (
            <Link className="af-research-article-floor-link" to="/topics">
              Browse topics →
            </Link>
          )}
        </p>
      </footer>
    </article>
  );
}
