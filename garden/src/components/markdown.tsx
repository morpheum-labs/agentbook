"use client";

import ReactMarkdown from "react-markdown";

interface MarkdownProps {
  content: string;
  className?: string;
  mentions?: string[]; // Valid mentions to highlight
}

function processMentions(content: string, mentions: string[] = []): string {
  if (!mentions.length) return content;

  let processed = content;
  for (const name of mentions) {
    const regex = new RegExp(`@${name}\\b`, "g");
    processed = processed.replace(regex, `**[@${name}](/u/${name})**`);
  }
  return processed;
}

export function Markdown({ content, className = "", mentions = [] }: MarkdownProps) {
  const processedContent = processMentions(content, mentions);

  return (
    <div className={`max-w-none text-body text-foreground ${className}`}>
      <ReactMarkdown
        components={{
          h1: ({ children }) => (
            <h1 className="text-lead font-medium text-foreground mt-6 mb-3 first:mt-0">{children}</h1>
          ),
          h2: ({ children }) => (
            <h2 className="text-lead font-medium text-foreground mt-5 mb-2">{children}</h2>
          ),
          h3: ({ children }) => (
            <h3 className="text-card-title text-foreground mt-4 mb-2">{children}</h3>
          ),
          p: ({ children }) => <p className="text-foreground mb-4 leading-[var(--lh-body)]">{children}</p>,
          ul: ({ children }) => (
            <ul className="list-disc list-inside text-foreground mb-4 space-y-2">{children}</ul>
          ),
          ol: ({ children }) => (
            <ol className="list-decimal list-inside text-foreground mb-4 space-y-2">{children}</ol>
          ),
          li: ({ children }) => <li className="text-foreground leading-[var(--lh-body)]">{children}</li>,
          a: ({ href, children }) => (
            <a
              href={href}
              className="text-link underline underline-offset-4 hover:opacity-90"
              target="_blank"
              rel="noopener noreferrer"
            >
              {children}
            </a>
          ),
          code: ({ className, children }) => {
            const isInline = !className;
            if (isInline) {
              return (
                <code className="bg-muted text-foreground px-1.5 py-0.5 rounded-sm text-caption-body">
                  {children}
                </code>
              );
            }
            return (
              <code className="block bg-muted text-foreground p-4 rounded-lg border border-border overflow-x-auto text-caption-body leading-[var(--lh-body)]">
                {children}
              </code>
            );
          },
          pre: ({ children }) => (
            <pre className="bg-muted text-foreground p-4 rounded-lg border border-border overflow-x-auto mb-4 text-caption-body">
              {children}
            </pre>
          ),
          blockquote: ({ children }) => (
            <blockquote className="border-l-2 border-chart-5 pl-4 text-muted-foreground italic my-4 leading-[var(--lh-body)]">
              {children}
            </blockquote>
          ),
          strong: ({ children }) => (
            <strong className="text-foreground font-semibold">{children}</strong>
          ),
          em: ({ children }) => <em className="text-foreground italic">{children}</em>,
          hr: () => <hr className="border-border my-8" />,
          table: ({ children }) => (
            <div className="overflow-x-auto mb-4">
              <table className="min-w-full border border-border text-caption-body">{children}</table>
            </div>
          ),
          th: ({ children }) => (
            <th className="border border-border px-3 py-2 bg-muted text-foreground font-semibold text-left">
              {children}
            </th>
          ),
          td: ({ children }) => (
            <td className="border border-border px-3 py-2 text-foreground">{children}</td>
          ),
        }}
      >
        {processedContent}
      </ReactMarkdown>
    </div>
  );
}
