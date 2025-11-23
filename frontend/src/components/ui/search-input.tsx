'use client'

import * as React from "react"
import { Search, X } from "lucide-react"
import { cn } from "../../lib/utils"
import { Input } from "./input"
import { useDebounce } from "../../hooks/useDebounce"

interface SearchInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'onChange'> {
  onSearch: (value: string) => void
  debounceMs?: number
  clearable?: boolean
  loading?: boolean
}

const SearchInput = React.forwardRef<HTMLInputElement, SearchInputProps>(
  ({ 
    className, 
    onSearch, 
    debounceMs = 300, 
    clearable = true, 
    loading = false,
    value: propValue,
    defaultValue,
    ...props 
  }, ref) => {
    const [internalValue, setInternalValue] = React.useState(
      propValue?.toString() || defaultValue?.toString() || ''
    )
    
    const debouncedValue = useDebounce(internalValue, debounceMs)

    // Handle controlled vs uncontrolled
    const value = propValue !== undefined ? propValue.toString() : internalValue

    React.useEffect(() => {
      onSearch(debouncedValue)
    }, [debouncedValue, onSearch])

    React.useEffect(() => {
      if (propValue !== undefined) {
        setInternalValue(propValue.toString())
      }
    }, [propValue])

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const newValue = e.target.value
      if (propValue === undefined) {
        setInternalValue(newValue)
      }
    }

    const handleClear = () => {
      const newValue = ''
      if (propValue === undefined) {
        setInternalValue(newValue)
      }
      onSearch(newValue)
    }

    return (
      <div className="relative">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
        <Input
          ref={ref}
          type="text"
          value={value}
          onChange={handleChange}
          className={cn("pl-10 pr-10", className)}
          {...props}
        />
        {loading && (
          <div className="absolute right-8 top-1/2 transform -translate-y-1/2">
            <div className="animate-spin rounded-full h-4 w-4 border-2 border-gray-200 border-t-gray-600" />
          </div>
        )}
        {clearable && value && !loading && (
          <button
            type="button"
            onClick={handleClear}
            className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>
    )
  }
)
SearchInput.displayName = "SearchInput"

export { SearchInput }