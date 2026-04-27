import * as React from "react"
import * as DialogPrimitive from "@radix-ui/react-dialog"
import { X } from "lucide-react"

import { cn } from "@/lib/utils"

const Dialog = DialogPrimitive.Root
const DialogTrigger = DialogPrimitive.Trigger
const DialogPortal = DialogPrimitive.Portal
const DialogClose = DialogPrimitive.Close

const DialogOverlay = React.forwardRef<
  React.ComponentRef<typeof DialogPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Overlay
    ref={ref}
    className={cn("fixed inset-0 z-[var(--z-modal)] bg-black/50 backdrop-blur-[1px]", className)}
    {...props}
  />
))
DialogOverlay.displayName = DialogPrimitive.Overlay.displayName

const DialogContent = React.forwardRef<
  React.ComponentRef<typeof DialogPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content> & {
    /** Hide the built-in close control */
    hideCloseButton?: boolean
  }
>(({ className, children, hideCloseButton, ...props }, ref) => (
  <DialogPortal>
    <DialogOverlay />
    <DialogPrimitive.Content
      ref={ref}
      className={cn(
        "fixed top-[50%] left-[50%] z-[var(--z-modal)] w-[calc(100%-2rem)] max-w-lg translate-x-[-50%] translate-y-[-50%]",
        "rounded-2xl border border-border bg-card p-0 shadow-elevation-3",
        "outline-none focus:outline-none",
        className
      )}
      {...props}
    >
      {children}
      {!hideCloseButton && (
        <DialogClose
          className="ring-offset-background absolute top-3 right-3 z-10 rounded-md p-1.5 text-muted-foreground opacity-80 transition-opacity hover:opacity-100 hover:bg-muted focus:ring-2 focus:ring-ring focus:ring-offset-2 focus-visible:outline-none disabled:pointer-events-none [&_svg]:size-4"
          aria-label="Close"
        >
          <X />
        </DialogClose>
      )}
    </DialogPrimitive.Content>
  </DialogPortal>
))
DialogContent.displayName = DialogPrimitive.Content.displayName

const DialogHeader = ({ className, ...props }: React.ComponentProps<"div">) => (
  <div
    className={cn("flex flex-col space-y-1.5 border-b border-border/60 px-5 pt-5 pb-4", className)}
    {...props}
  />
)
DialogHeader.displayName = "DialogHeader"

const DialogTitle = React.forwardRef<
  React.ComponentRef<typeof DialogPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Title>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Title
    ref={ref}
    className={cn("text-subheading-lg text-foreground pr-8 font-medium leading-none tracking-tight", className)}
    {...props}
  />
))
DialogTitle.displayName = DialogPrimitive.Title.displayName

const DialogDescription = React.forwardRef<
  React.ComponentRef<typeof DialogPrimitive.Description>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Description>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Description
    ref={ref}
    className={cn("text-caption-body text-muted-foreground pt-1", className)}
    {...props}
  />
))
DialogDescription.displayName = DialogPrimitive.Description.displayName

const DialogBody = ({ className, ...props }: React.ComponentProps<"div">) => (
  <div className={cn("max-h-[min(70vh,32rem)] overflow-y-auto px-5 py-4", className)} {...props} />
)
DialogBody.displayName = "DialogBody"

const DialogFooter = ({ className, ...props }: React.ComponentProps<"div">) => (
  <div
    className={cn("flex flex-col-reverse gap-2 border-t border-border/60 sm:flex-row sm:justify-end px-5 py-4", className)}
    {...props}
  />
)
DialogFooter.displayName = "DialogFooter"

export {
  Dialog,
  DialogBody,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogOverlay,
  DialogPortal,
}
