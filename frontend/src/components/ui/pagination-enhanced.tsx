'use client'

import * as React from "react"
import { cn } from "../../lib/utils"
import {
  Pagination as ShadcnPagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "./pagination"

interface PaginationProps {
  currentPage: number
  totalPages: number
  onPageChange: (page: number) => void
  showFirstLast?: boolean
  showPrevNext?: boolean
  maxVisiblePages?: number
  className?: string
}

const Pagination: React.FC<PaginationProps> = ({
  currentPage,
  totalPages,
  onPageChange,
  showFirstLast = true,
  showPrevNext = true,
  maxVisiblePages = 5,
  className
}) => {
  const getVisiblePages = () => {
    if (totalPages <= maxVisiblePages) {
      return Array.from({ length: totalPages }, (_, i) => i + 1)
    }

    const halfVisible = Math.floor(maxVisiblePages / 2)
    let start = Math.max(1, currentPage - halfVisible)
    let end = Math.min(totalPages, start + maxVisiblePages - 1)

    if (end - start + 1 < maxVisiblePages) {
      start = Math.max(1, end - maxVisiblePages + 1)
    }

    const pages = Array.from({ length: end - start + 1 }, (_, i) => start + i)
    
    // Add ellipsis logic
    const visiblePages: (number | 'ellipsis')[] = []
    
    if (start > 1) {
      visiblePages.push(1)
      if (start > 2) {
        visiblePages.push('ellipsis')
      }
    }
    
    visiblePages.push(...pages)
    
    if (end < totalPages) {
      if (end < totalPages - 1) {
        visiblePages.push('ellipsis')
      }
      visiblePages.push(totalPages)
    }
    
    return visiblePages
  }

  const visiblePages = getVisiblePages()

  if (totalPages <= 1) return null

  return (
    <ShadcnPagination className={cn(className)}>
      <PaginationContent>
        {showFirstLast && currentPage > 1 && (
          <PaginationItem>
            <PaginationLink 
              onClick={() => onPageChange(1)}
              className="cursor-pointer"
            >
              First
            </PaginationLink>
          </PaginationItem>
        )}
        
        {showPrevNext && currentPage > 1 && (
          <PaginationItem>
            <PaginationPrevious
              onClick={() => onPageChange(currentPage - 1)}
              className="cursor-pointer"
            />
          </PaginationItem>
        )}

        {visiblePages.map((page, index) => {
          if (page === 'ellipsis') {
            return (
              <PaginationItem key={`ellipsis-${index}`}>
                <PaginationEllipsis />
              </PaginationItem>
            )
          }

          return (
            <PaginationItem key={page}>
              <PaginationLink
                onClick={() => onPageChange(page)}
                isActive={currentPage === page}
                className="cursor-pointer"
              >
                {page}
              </PaginationLink>
            </PaginationItem>
          )
        })}

        {showPrevNext && currentPage < totalPages && (
          <PaginationItem>
            <PaginationNext
              onClick={() => onPageChange(currentPage + 1)}
              className="cursor-pointer"
            />
          </PaginationItem>
        )}

        {showFirstLast && currentPage < totalPages && (
          <PaginationItem>
            <PaginationLink
              onClick={() => onPageChange(totalPages)}
              className="cursor-pointer"
            >
              Last
            </PaginationLink>
          </PaginationItem>
        )}
      </PaginationContent>
    </ShadcnPagination>
  )
}

interface PaginationInfoProps {
  currentPage: number
  totalItems: number
  itemsPerPage: number
  className?: string
}

const PaginationInfo: React.FC<PaginationInfoProps> = ({
  currentPage,
  totalItems,
  itemsPerPage,
  className
}) => {
  const startItem = (currentPage - 1) * itemsPerPage + 1
  const endItem = Math.min(currentPage * itemsPerPage, totalItems)

  return (
    <div className={cn("text-sm text-muted-foreground", className)}>
      Showing {startItem} to {endItem} of {totalItems} results
    </div>
  )
}

interface PaginationContainerProps {
  children: React.ReactNode
  pagination: React.ReactNode
  info?: React.ReactNode
  className?: string
}

const PaginationContainer: React.FC<PaginationContainerProps> = ({
  children,
  pagination,
  info,
  className
}) => {
  return (
    <div className={cn("space-y-4", className)}>
      {children}
      <div className="flex items-center justify-between">
        <div className="flex-1">
          {info}
        </div>
        <div className="flex-1 flex justify-end">
          {pagination}
        </div>
      </div>
    </div>
  )
}

export { Pagination, PaginationInfo, PaginationContainer }