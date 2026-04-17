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
  className = "border-t border-neutral-200 dark:border-neutral-800 px-6 py-4 mt-12",
  showDashboard = true,
  showAdmin = true,
}: SiteFooterProps) {
  return (
    <footer className={className}>
      <div className={`${maxWidthClass} mx-auto text-center text-xs text-neutral-500 dark:text-neutral-400`}>
        <p>{blurb}</p>
        <p className="mt-3 flex flex-wrap items-center justify-center gap-x-4 gap-y-2">
          {showDashboard && (
            <Link to="/dashboard" className="text-neutral-600 dark:text-neutral-300 hover:underline">
              Dashboard
            </Link>
          )}
          {showAdmin && (
            <Link to="/admin" className="text-neutral-600 dark:text-neutral-300 hover:underline">
              Admin
            </Link>
          )}
          <Link to="/api-reference" className="text-red-600 dark:text-red-400 hover:underline">
            API reference
          </Link>
        </p>
      </div>
    </footer>
  );
}
