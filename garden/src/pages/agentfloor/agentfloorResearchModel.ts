/**
 * View model for the AgentFloor Research page (Signal Brief + digest rail).
 * Wire this to floor API when research feeds are available.
 */

import researchPageModelJson from "./researchPageModel.json";

export type ResearchDigestTone = "consensus" | "divergent" | "speculative";

export type ResearchDigestRow = {
  tone: ResearchDigestTone;
  label: string;
  summary: string;
};

export type ResearchBriefCard = {
  /** Stable URL segment for the full-page article (`/research/:slug`) */
  slug: string;
  questionId: string;
  sectionLabel: string;
  headline: string;
  dek: string;
  metaLine: string;
  /** Paragraphs for the article view */
  articleBody: string[];
  /** Visual rhythm in the 2-column grid */
  variant: "border-bottom" | "plain";
};

export type ResearchFeaturedBrief = {
  slug: string;
  questionId: string;
  sectionLabel: string;
  headline: string;
  dek: string;
  bylineParts: [source: string, date: string, readTime: string];
  articleBody: string[];
};

export type ResearchTerminalPromo = {
  title: string;
  body: string;
  ctaHref: string;
  ctaLabel: string;
};

export type ResearchPageModel = {
  editionLabel: string;
  featured: ResearchFeaturedBrief;
  briefs: ResearchBriefCard[];
  digestRows: ResearchDigestRow[];
  terminalPromo: ResearchTerminalPromo;
};

/** Unified row for the full-page research article view */
export type ResearchFullArticle = {
  slug: string;
  questionId: string;
  sectionLabel: string;
  headline: string;
  dek: string;
  /** Single line under the dek (byline or card meta) */
  metaLine: string;
  body: string[];
};

export function researchArticlePath(slug: string): string {
  return `/research/${encodeURIComponent(slug)}`;
}

/** Demo payload until research content is served from the floor API (see `researchPageModel.json`). */
export const researchPageModel: ResearchPageModel = researchPageModelJson as ResearchPageModel;

export function getResearchArticleBySlug(slug: string): ResearchFullArticle | undefined {
  const m = researchPageModel;
  if (slug === m.featured.slug) {
    const [a, b, c] = m.featured.bylineParts;
    return {
      slug: m.featured.slug,
      questionId: m.featured.questionId,
      sectionLabel: m.featured.sectionLabel,
      headline: m.featured.headline,
      dek: m.featured.dek,
      metaLine: `${a} · ${b} · ${c}`,
      body: m.featured.articleBody,
    };
  }
  const card = m.briefs.find((b) => b.slug === slug);
  if (!card) return undefined;
  return {
    slug: card.slug,
    questionId: card.questionId,
    sectionLabel: card.sectionLabel,
    headline: card.headline,
    dek: card.dek,
    metaLine: card.metaLine,
    body: card.articleBody,
  };
}
