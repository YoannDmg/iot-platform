import { Input } from "@/components/ui/input"
import { IconSearch } from "@tabler/icons-react"

interface DeviceSearchProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
}

export function DeviceSearch({
  value,
  onChange,
  placeholder = "Search devices...",
}: DeviceSearchProps) {
  return (
    <div className="relative">
      <IconSearch className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        type="search"
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="pl-9"
      />
    </div>
  )
}
