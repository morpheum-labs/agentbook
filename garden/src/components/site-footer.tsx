import { Link } from "react-router-dom";
import { cn } from "@/lib/utils";

interface SiteFooterProps {
  blurb: string;
  /** Optional extra classes on the inner content wrapper (e.g. align with a narrow main column). */
  innerClassName?: string;
  className?: string;
  showDashboard?: boolean;
  showAdmin?: boolean;
}

export function SiteFooter({
  blurb,
  innerClassName,
  className = "mt-12 border-t border-border bg-background py-6",
  showDashboard = true,
  showAdmin = true,
}: SiteFooterProps) {
  return (
    <footer className={className}>
      <div
        className={cn(
          "container-app text-center text-caption text-muted-foreground",
          innerClassName
        )}
      >
        <p>{blurb}</p>
        <p className="mt-4 flex flex-wrap items-center justify-center gap-x-6 gap-y-2">
          {showDashboard && (
            <Link to="/dashboard" className="text-link underline underline-offset-4 hover:opacity-90">
              Dashboard
            </Link>
          )}
          {showAdmin && (
            <Link to="/admin" className="text-link underline underline-offset-4 hover:opacity-90">
              Admin
            </Link>
          )}
          <Link to="/api-reference" className="text-link underline underline-offset-4 hover:opacity-90">
            API reference
          </Link>
        </p>
      </div>
    </footer>
  );
}
