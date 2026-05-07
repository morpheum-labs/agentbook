import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

/**
 * Hyperlink presentation for actions (preview-dark.html <a> parity):
 * underline + accent color, hover fills like global anchor hover glow + primary wash.
 */
const buttonVariants = cva(
  [
    "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-none border border-transparent",
    "text-ui-semi shadow-none outline-none transition-[color,background-color,box-shadow,text-decoration-color]",
    "disabled:pointer-events-none disabled:opacity-50",
    "[&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0",
    "focus-visible:ring-[3px] focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40",
  ].join(" "),
  {
    variants: {
      variant: {
        default:
          "bg-transparent text-link underline underline-offset-[3px] decoration-link/75 hover:bg-primary hover:text-primary-foreground hover:no-underline hover:shadow-[0_0_14px_color-mix(in_srgb,var(--color-link)_35%,transparent)]",
        destructive:
          "bg-transparent text-destructive underline underline-offset-[3px] decoration-destructive/65 hover:bg-destructive hover:text-destructive-foreground hover:no-underline hover:shadow-none focus-visible:ring-destructive/35",
        outline:
          "bg-transparent text-link underline underline-offset-[3px] decoration-dotted decoration-link/85 hover:bg-primary hover:text-primary-foreground hover:no-underline hover:shadow-[0_0_14px_color-mix(in_srgb,var(--color-link)_35%,transparent)]",
        secondary:
          "bg-transparent text-foreground underline underline-offset-[3px] decoration-primary/40 hover:bg-primary hover:text-primary-foreground hover:no-underline hover:shadow-[0_0_14px_color-mix(in_srgb,var(--color-link)_35%,transparent)]",
        ghost:
          "bg-transparent text-foreground no-underline hover:text-link hover:underline hover:underline-offset-[3px] hover:shadow-[0_0_14px_color-mix(in_srgb,var(--color-link)_35%,transparent)]",
        link:
          "bg-transparent text-link underline underline-offset-[3px] decoration-link/75 hover:bg-primary hover:text-primary-foreground hover:no-underline hover:shadow-[0_0_14px_color-mix(in_srgb,var(--color-link)_35%,transparent)]",
      },
      size: {
        default: "min-h-10 px-3 py-2 has-[>svg]:px-2.5",
        xs: "min-h-8 gap-1 px-2 py-1.5 text-caption has-[>svg]:px-1.5 [&_svg:not([class*='size-'])]:size-3",
        sm: "min-h-9 gap-1.5 px-2.5 py-2 text-caption has-[>svg]:px-2",
        lg: "min-h-11 gap-2 px-4 py-2.5 has-[>svg]:px-3",
        icon: "size-10 min-h-10 min-w-10 gap-0 p-0 [&_svg:not([class*='size-'])]:size-4",
        "icon-xs": "size-8 min-h-8 min-w-8 gap-0 p-0 [&_svg:not([class*='size-'])]:size-3",
        "icon-sm": "size-9 min-h-9 min-w-9 gap-0 p-0 [&_svg:not([class*='size-'])]:size-4",
        "icon-lg": "size-11 min-h-11 min-w-11 gap-0 p-0 [&_svg:not([class*='size-'])]:size-5",
      },
    },
    compoundVariants: [
      {
        size: ["icon", "icon-xs", "icon-sm", "icon-lg"],
        class: "underline-none decoration-transparent hover:no-underline hover:shadow-none",
      },
      {
        variant: "ghost",
        size: ["icon", "icon-xs", "icon-sm", "icon-lg"],
        class:
          "text-muted-foreground hover:text-foreground hover:bg-accent/65 dark:hover:bg-[rgba(var(--terminal-bg-rgb),0.22)]",
      },
      {
        variant: "destructive",
        size: ["icon", "icon-xs", "icon-sm", "icon-lg"],
        class: "text-destructive hover:bg-destructive hover:text-destructive-foreground",
      },
    ],
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

function Button({
  className,
  variant = "default",
  size = "default",
  asChild = false,
  ...props
}: React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean
  }) {
  const Comp = asChild ? Slot : "button"

  return (
    <Comp
      data-slot="button"
      data-variant={variant}
      data-size={size}
      className={cn(buttonVariants({ variant, size }), className)}
      {...props}
    />
  )
}

export { Button, buttonVariants }
