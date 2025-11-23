# Alert Components

Standard shadcn/ui alert components for the authentication service frontend.

## Components

### Basic Alert
```tsx
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert';

<Alert variant="success" dismissible onDismiss={() => console.log('dismissed')}>
  <AlertTitle>Success!</AlertTitle>
  <AlertDescription>Your operation completed successfully.</AlertDescription>
</Alert>
```

### Pre-configured Alerts
```tsx
import { SuccessAlert, ErrorAlert, WarningAlert, InfoAlert } from '@/components/ui/alert';

<SuccessAlert title="Success!" dismissible>
  Operation completed successfully.
</SuccessAlert>

<ErrorAlert title="Error Occurred" dismissible>
  Something went wrong. Please try again.
</ErrorAlert>

<WarningAlert title="Warning" dismissible>
  Please review your input before proceeding.
</WarningAlert>

<InfoAlert title="Information" dismissible>
  Here's some helpful information.
</InfoAlert>
```

### Alert Dialogs
```tsx
import { ConfirmDialog, DeleteDialog } from '@/components/ui/alert-dialog';

<ConfirmDialog
  open={showDialog}
  onOpenChange={setShowDialog}
  title="Confirm Action"
  description="Are you sure you want to proceed?"
  onConfirm={handleConfirm}
  variant="warning"
  loading={isLoading}
/>

<DeleteDialog
  open={showDeleteDialog}
  onOpenChange={setShowDeleteDialog}
  title="Delete Item"
  description="This action cannot be undone."
  onConfirm={handleDelete}
  loading={isDeleting}
/>
```

## Props

### Alert Props
- `variant`: "default" | "destructive" | "success" | "warning" | "info"
- `size`: "default" | "sm" | "lg"
- `dismissible`: boolean - Shows dismiss button
- `onDismiss`: () => void - Callback when dismissed
- `icon`: React.ReactNode - Custom icon
- `showIcon`: boolean - Show/hide icon (default: true)

### Alert Dialog Props
- `open`: boolean - Dialog open state
- `onOpenChange`: (open: boolean) => void - State change handler
- `title`: string - Dialog title
- `description`: string - Dialog description
- `variant`: "default" | "destructive" | "success" | "warning" | "info"
- `confirmText`: string - Confirm button text
- `cancelText`: string - Cancel button text
- `onConfirm`: () => void - Confirm callback
- `onCancel`: () => void - Cancel callback
- `loading`: boolean - Loading state

## Features

- 🎨 **Multiple variants** - Success, error, warning, info, and default styles
- 📱 **Responsive design** - Works on all screen sizes
- ♿ **Accessible** - Full ARIA support and keyboard navigation
- 🔧 **Customizable** - Custom icons, sizes, and styling
- ✨ **Animations** - Smooth transitions and hover effects
- 🎭 **Themeable** - Dark mode support
- 📦 **Pre-configured** - Ready-to-use components for common scenarios
- 🚫 **Dismissible** - Optional dismiss functionality

## Examples

Visit `/alerts-demo` to see all components in action with interactive examples.