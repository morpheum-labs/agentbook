import * as React from "react"

import { cn } from "@/lib/utils"

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
  const resolvedType = type ?? "text"
  const isTextField = resolvedType === "text"

  return (
    <input
      type={type}
      data-slot="input"
      className={cn(
        "file:text-foreground placeholder:text-muted-foreground h-10 w-full min-w-0 rounded-none border border-border bg-transparent px-3 py-2 leading-[var(--lh-body)] shadow-elevation-0 outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-caption file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 text-body",
        "aria-invalid:border-destructive aria-invalid:ring-0 dark:aria-invalid:ring-0",
        className,
        isTextField
          ? "motion-safe:transition-[border-color] motion-safe:duration-300 motion-safe:ease-out focus:!border-white focus-visible:!border-white focus:!ring-0 focus-visible:!ring-0 focus-visible:!outline-none aria-invalid:focus:!border-destructive aria-invalid:focus-visible:!border-destructive"
          : "transition-[color,box-shadow] focus-visible:border-ring focus-visible:outline-none focus-visible:ring-ring/50 focus-visible:ring-[3px]"
      )}
      {...props}
    />
  )
}

export { Input }
