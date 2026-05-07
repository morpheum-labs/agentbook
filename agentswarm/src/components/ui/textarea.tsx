import * as React from "react"

import { cn } from "@/lib/utils"

const Textarea = React.forwardRef<HTMLTextAreaElement, React.ComponentProps<"textarea">>(
  function Textarea({ className, ...props }, ref) {
    return (
      <textarea
        ref={ref}
        data-slot="textarea"
        className={cn(
          "border-border placeholder:text-muted-foreground flex field-sizing-content min-h-16 w-full rounded-none border bg-transparent px-3 py-2 leading-[var(--lh-body)] shadow-elevation-0 outline-none disabled:cursor-not-allowed disabled:opacity-50 text-body",
          "aria-invalid:border-destructive aria-invalid:ring-0 dark:aria-invalid:ring-0",
          className,
          "motion-safe:transition-[border-color] motion-safe:duration-300 motion-safe:ease-out focus:!border-white focus-visible:!border-white focus:!ring-0 focus-visible:!ring-0 focus-visible:!outline-none aria-invalid:focus:!border-destructive aria-invalid:focus-visible:!border-destructive"
        )}
        {...props}
      />
    )
  }
)

export { Textarea }
