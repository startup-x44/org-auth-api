'use client';

import { Card } from '@/components/ui/card';
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert';
import { 
  AlertTriangle, 
  Terminal,
  CheckCircle2,
  AlertCircle,
  Info,
  XCircle
} from 'lucide-react';

export default function AlertShowcase() {
  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4">
      <div className="max-w-4xl mx-auto space-y-8">
        
        {/* Header */}
        <div className="text-center">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Shadcn/UI Alert Components</h1>
          <p className="text-gray-600">Clean, accessible, and beautiful alert components</p>
        </div>

        {/* Default Alert */}
        <Card className="p-6">
          <h2 className="text-xl font-semibold mb-4">Default Alert</h2>
          <Alert>
            <Terminal className="h-4 w-4" />
            <AlertTitle>Heads up!</AlertTitle>
            <AlertDescription>
              You can add components and dependencies to your app using the cli.
            </AlertDescription>
          </Alert>
        </Card>

        {/* Destructive Alert */}
        <Card className="p-6">
          <h2 className="text-xl font-semibold mb-4">Destructive Alert</h2>
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>
              Your session has expired. Please log in again.
            </AlertDescription>
          </Alert>
        </Card>

        {/* Multiple Examples */}
        <Card className="p-6">
          <h2 className="text-xl font-semibold mb-4">More Examples</h2>
          <div className="space-y-4">
            
            <Alert>
              <Info className="h-4 w-4" />
              <AlertTitle>Information</AlertTitle>
              <AlertDescription>
                This is some important information you should know about.
              </AlertDescription>
            </Alert>

            <Alert variant="destructive">
              <XCircle className="h-4 w-4" />
              <AlertTitle>Something went wrong</AlertTitle>
              <AlertDescription>
                There was a problem processing your request. Please try again later.
              </AlertDescription>
            </Alert>

            <Alert>
              <CheckCircle2 className="h-4 w-4" />
              <AlertTitle>Success!</AlertTitle>
              <AlertDescription>
                Your changes have been saved successfully.
              </AlertDescription>
            </Alert>

            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>Warning</AlertTitle>
              <AlertDescription>
                Please review your changes before submitting.
              </AlertDescription>
            </Alert>

          </div>
        </Card>

        {/* Usage Examples */}
        <Card className="p-6">
          <h2 className="text-xl font-semibold mb-4">Usage</h2>
          <div className="bg-slate-950 text-slate-50 rounded-lg p-4 overflow-x-auto">
            <pre className="text-sm">
{`import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Terminal } from "lucide-react"

// Default Alert
<Alert>
  <Terminal className="h-4 w-4" />
  <AlertTitle>Heads up!</AlertTitle>
  <AlertDescription>
    You can add components to your app using the cli.
  </AlertDescription>
</Alert>

// Destructive Alert
<Alert variant="destructive">
  <AlertCircle className="h-4 w-4" />
  <AlertTitle>Error</AlertTitle>
  <AlertDescription>
    Your session has expired. Please log in again.
  </AlertDescription>
</Alert>`}
            </pre>
          </div>
        </Card>

      </div>
    </div>
  );
}

