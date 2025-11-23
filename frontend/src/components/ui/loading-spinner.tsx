import * as React from "react"
import { cn } from "../../lib/utils"
import { Spinner } from "./spinner"

/* -------------------------------------------
 * Loading Spinner - Wrapper around shadcn Spinner
 * ------------------------------------------- */

interface LoadingSpinnerProps extends Omit<React.ComponentProps<"svg">, 'size'> {
  size?: "sm" | "md" | "lg" | "xl"
  variant?: "default" | "primary" | "secondary"
  fullScreen?: boolean
}

const LoadingSpinner = React.forwardRef<SVGSVGElement, LoadingSpinnerProps>(
  ({ className, size = "md", variant = "default", fullScreen = false, ...props }, ref) => {
    
    const sizeClasses = {
      sm: "size-4",
      md: "size-8", 
      lg: "size-12",
      xl: "size-16",
    }

    const variantClasses = {
      default: "text-muted-foreground",
      primary: "text-primary",
      secondary: "text-muted-foreground/70",
    }

    const spinner = (
      <Spinner
        ref={ref}
        className={cn(
          sizeClasses[size],
          variantClasses[variant],
          className
        )}
        {...props}
      />
    )

    if (fullScreen) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-background">
          {spinner}
        </div>
      )
    }

    return spinner
  }
)

LoadingSpinner.displayName = "LoadingSpinner"

/* -------------------------------------------
 * Loading Overlay
 * ------------------------------------------- */

interface LoadingOverlayProps {
  children?: React.ReactNode
  loading: boolean
  spinnerSize?: "sm" | "md" | "lg" | "xl"
  overlay?: boolean
}

const LoadingOverlay: React.FC<LoadingOverlayProps> = ({
  children,
  loading,
  spinnerSize = "lg",
  overlay = true,
}) => {
  if (!loading) return <>{children}</>

  return (
    <div className="relative">
      {children}

      {overlay ? (
        <div className="absolute inset-0 bg-background/70 backdrop-blur-sm z-10 flex items-center justify-center">
          <LoadingSpinner size={spinnerSize} variant="primary" />
        </div>
      ) : (
        <div className="flex items-center justify-center py-8">
          <LoadingSpinner size={spinnerSize} variant="primary" />
        </div>
      )}
    </div>
  )
}

export { LoadingSpinner, LoadingOverlay }
