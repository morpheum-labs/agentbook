import { type ReactNode } from "react";
import { Link } from "react-router-dom";
import { CalendarClock, Keyboard, type LucideIcon } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { cn } from "@/lib/utils";

const tips: { icon: LucideIcon; text: string }[] = [
  { icon: CalendarClock, text: "Use a cron string or a label your runner understands." },
  { icon: Keyboard, text: "Prompt is what the agent sees when the job fires — keep it action-oriented." },
];

type CronJobFormLayoutProps = {
  title: string;
  description: string;
  children: ReactNode;
  /** Extra content in the card header (e.g. id line, errors) */
  headerExtra?: ReactNode;
  className?: string;
};

/**
 * Form shell: left = Dribbble-style “insight” panel, right = card + fields.
 */
export function CronJobFormLayout({
  title,
  description,
  children,
  headerExtra,
  className,
}: CronJobFormLayoutProps) {
  return (
    <div className={className}>
      <div className="container-app max-w-5xl py-8 sm:py-10">
        <p className="text-caption text-muted-foreground mb-3">
          <Link
            to="/cron-jobs"
            className="text-link hover:underline underline-offset-2 font-medium"
          >
            Cron jobs
          </Link>
          <span className="text-border mx-2" aria-hidden>
            /
          </span>
          <span className="text-foreground">Editor</span>
        </p>
        <div className="grid gap-6 lg:grid-cols-12 lg:gap-8">
          <aside className="lg:col-span-4">
            <div
              className={cn(
                "relative overflow-hidden rounded-2xl border border-border/70",
                "bg-gradient-to-b from-[var(--mysteria-purple)]/8 via-card to-card",
                "p-5 sm:p-6 shadow-elevation-1",
                "dark:from-[var(--mysteria-purple)]/25 dark:via-card dark:to-card"
              )}
            >
              <div
                className="absolute -right-8 top-0 size-32 rounded-full bg-[var(--amethyst-link)]/15 blur-2xl"
                aria-hidden
              />
              <p className="text-micro text-muted-foreground relative uppercase">Design note</p>
              <h2 className="text-subheading-lg text-foreground relative mt-1 font-medium">Rhythm & intent</h2>
              <p className="text-caption-body text-muted-foreground relative mt-2 leading-relaxed">
                Treat each job as a small contract: <strong className="text-foreground/90">who</strong> runs,{" "}
                <strong className="text-foreground/90">when</strong>, and{" "}
                <strong className="text-foreground/90">what to do</strong>.
              </p>
              <ul className="relative mt-6 space-y-3" role="list">
                {tips.map(({ icon: Icon, text }) => (
                  <li
                    key={text}
                    className="flex gap-3 rounded-xl border border-border/50 bg-background/60 p-3 text-caption text-muted-foreground backdrop-blur-sm"
                  >
                    <span
                      className={cn(
                        "flex size-8 shrink-0 items-center justify-center rounded-lg",
                        "bg-gradient-to-br from-[var(--mysteria-purple)] to-[var(--amethyst-link)]",
                        "text-white shadow-sm"
                      )}
                    >
                      <Icon className="size-4" aria-hidden />
                    </span>
                    <span className="pt-0.5 leading-snug">{text}</span>
                  </li>
                ))}
              </ul>
            </div>
          </aside>
          <div className="lg:col-span-8">
            <Card className="overflow-hidden border-border/80 shadow-elevation-2">
              <div
                className="h-1 w-full bg-gradient-to-r from-[var(--mysteria-purple)] via-[var(--amethyst-link)] to-[var(--lavender-glow)]"
                aria-hidden
              />
              <CardHeader className="space-y-1">
                <CardTitle className="text-subheading-lg text-foreground">{title}</CardTitle>
                <CardDescription className="text-caption-body text-pretty">{description}</CardDescription>
                {headerExtra}
              </CardHeader>
              <CardContent className="pt-0">{children}</CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
