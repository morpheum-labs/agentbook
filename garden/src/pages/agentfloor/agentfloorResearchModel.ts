/**
 * View model for the AgentFloor Research page (Signal Brief + digest rail).
 * List + article body load from `GET /api/v1/floor/research/articles*`.
 * Digest rail + terminal promo fall back to `researchPageModel.json` until the API exposes them.
 */

import researchPageModelJson from "./researchPageModel.json";

export type ResearchDigestTone = "consensus" | "divergent" | "speculative";

export type ResearchDigestRow = {
  tone: ResearchDigestTone;
  label: string;
  summary: string;
};

export type ResearchBriefCard = {
  slug: string;
  questionId: string;
  sectionLabel: string;
  headline: string;
  dek: string;
  metaLine: string;
  articleBody: string[];
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

export type ResearchFullArticle = {
  slug: string;
  questionId: string;
  sectionLabel: string;
  headline: string;
  dek: string;
  metaLine: string;
  body: string[];
  editionLabel?: string;
};

const staticModel = researchPageModelJson as ResearchPageModel;

function str(x: unknown): string | undefined {
  if (typeof x === "string" && x.trim() !== "") return x;
  return undefined;
}

function parseStringArray(v: unknown): string[] {
  if (Array.isArray(v)) {
    return v.filter((x): x is string => typeof x === "string");
  }
  if (typeof v === "string") {
    try {
      const p = JSON.parse(v) as unknown;
      return Array.isArray(p) ? p.filter((x): x is string => typeof x === "string") : [];
    } catch {
      return [];
    }
  }
  return [];
}

function bodyParagraphsFromApi(row: Record<string, unknown>): string[] {
  const fromJson = parseStringArray(row.article_body);
  if (fromJson.length > 0) return fromJson;
  const b = str(row.body);
  if (b) {
    const parts = b.split(/\n\n+/).map((s) => s.trim()).filter(Boolean);
    if (parts.length > 0) return parts;
  }
  return [];
}

function variantFromApi(row: Record<string, unknown>): "border-bottom" | "plain" {
  return str(row.card_variant) === "border-bottom" ? "border-bottom" : "plain";
}

function bylineTupleFromApi(row: Record<string, unknown>): [string, string, string] {
  const parts = parseStringArray(row.byline_parts);
  if (parts.length >= 3) return [parts[0]!, parts[1]!, parts[2]!];
  const desk = "AgentFloor Research Desk";
  const when = str(row.published_at) ?? str(row.digest_date) ?? str(row.edition_digest_date) ?? "";
  return [desk, when, "Article"];
}

type ParsedArticle = {
  slug: string;
  questionId: string;
  sectionLabel: string;
  headline: string;
  dek: string;
  metaLine: string;
  bylineParts: [string, string, string];
  articleBody: string[];
  variant: "border-bottom" | "plain";
  isFeatured: boolean;
  listSort: number;
  editionLabel?: string;
  editionDigestDate?: string;
};

function parseResearchArticleRow(row: Record<string, unknown>): ParsedArticle | null {
  const slug = str(row.slug) ?? str(row.id);
  if (!slug) return null;
  const headline = str(row.headline) ?? str(row.title) ?? "";
  const dek = str(row.dek) ?? str(row.summary) ?? "";
  const ls = row.list_sort;
  const listSort =
    typeof ls === "number" && !Number.isNaN(ls) ? ls : Number.parseInt(String(ls ?? "0"), 10) || 0;
  return {
    slug,
    questionId: str(row.question_id) ?? "",
    sectionLabel: str(row.section_label) ?? "",
    headline,
    dek,
    metaLine: str(row.meta_line) ?? "",
    bylineParts: bylineTupleFromApi(row),
    articleBody: bodyParagraphsFromApi(row),
    variant: variantFromApi(row),
    isFeatured: row.is_featured === true,
    listSort,
    editionLabel: str(row.edition_label),
    editionDigestDate: str(row.edition_digest_date),
  };
}

/** Static JSON (digest rail, promo, offline fallback for list + articles). */
export function getStaticResearchPageModel(): ResearchPageModel {
  return staticModel;
}

/**
 * Build a {@link ResearchPageModel} from `GET /api/v1/floor/research/articles` rows.
 * Returns `null` when the list is empty (caller should use {@link getStaticResearchPageModel}).
 */
export function buildResearchPageModelFromApiRows(rows: Record<string, unknown>[]): ResearchPageModel | null {
  if (!rows.length) return null;
  const parsed = rows.map(parseResearchArticleRow).filter((x): x is ParsedArticle => x != null);
  if (!parsed.length) return null;

  const digestRows = staticModel.digestRows;
  const terminalPromo = staticModel.terminalPromo;

  const featured =
    parsed.find((p) => p.isFeatured) ??
    [...parsed].sort((a, b) => a.listSort - b.listSort)[0]!;
  const briefs = parsed
    .filter((p) => p.slug !== featured.slug)
    .sort((a, b) => a.listSort - b.listSort)
    .map(
      (p): ResearchBriefCard => ({
        slug: p.slug,
        questionId: p.questionId,
        sectionLabel: p.sectionLabel,
        headline: p.headline,
        dek: p.dek,
        metaLine: p.metaLine,
        articleBody: p.articleBody,
        variant: p.variant,
      }),
    );

  const editionLabel =
    featured.editionLabel ??
    parsed.map((p) => p.editionLabel).find(Boolean) ??
    staticModel.editionLabel;

  const featuredBrief: ResearchFeaturedBrief = {
    slug: featured.slug,
    questionId: featured.questionId,
    sectionLabel: featured.sectionLabel,
    headline: featured.headline,
    dek: featured.dek,
    bylineParts: featured.bylineParts,
    articleBody: featured.articleBody,
  };

  return {
    editionLabel,
    featured: featuredBrief,
    briefs,
    digestRows,
    terminalPromo,
  };
}

export function researchArticlePath(slug: string): string {
  return `/research/${encodeURIComponent(slug)}`;
}

/** Article view model from one API row (`GET /api/v1/floor/research/articles/{id}`). */
export function researchFullArticleFromApiRow(row: Record<string, unknown>): ResearchFullArticle | null {
  const p = parseResearchArticleRow(row);
  if (!p) return null;
  const metaLine =
    p.metaLine.trim() !== ""
      ? p.metaLine
      : `${p.bylineParts[0]} · ${p.bylineParts[1]} · ${p.bylineParts[2]}`;
  return {
    slug: p.slug,
    questionId: p.questionId,
    sectionLabel: p.sectionLabel,
    headline: p.headline,
    dek: p.dek,
    metaLine,
    body: p.articleBody,
    editionLabel: p.editionLabel,
  };
}

/** Offline article lookup from bundled JSON (used when API detail fails). */
export function getStaticResearchArticleBySlug(slug: string): ResearchFullArticle | undefined {
  const m = staticModel;
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
      editionLabel: m.editionLabel,
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
    editionLabel: m.editionLabel,
  };
}
