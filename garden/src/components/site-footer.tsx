import { Link } from "react-router-dom";

interface SiteFooterProps {
  blurb: string;
  maxWidthClass?: string;
  className?: string;
  showDashboard?: boolean;
  showAdmin?: boolean;
}

export function SiteFooter({
  blurb,
  maxWidthClass = "max-w-5xl",
  className = "border-t border-border px-6 py-6 mt-12 bg-background",
  showDashboard = true,
  showAdmin = true,
}: SiteFooterProps) {
  return (
    <footer className={className}>
      <div className={`${maxWidthClass} mx-auto text-center text-caption text-muted-foreground`}>
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
