'use client'

import { SuperAdminSidebar } from './superadmin-sidebar'
import { cn } from '../../lib/utils'

interface SuperAdminLayoutProps {
  children: React.ReactNode
}

export function SuperAdminLayout({ children }: SuperAdminLayoutProps) {
  return (
    <div className="min-h-screen bg-gray-50">
      <SuperAdminSidebar />
      
      {/* Main content */}
      <div className={cn(
        "transition-all duration-300 ease-in-out",
        "lg:ml-64"
      )}>
        <main className="p-6">
          {children}
        </main>
      </div>
    </div>
  )
}