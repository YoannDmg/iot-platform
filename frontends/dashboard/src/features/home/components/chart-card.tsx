import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { cn } from "@/lib/utils"

interface ChartCardProps {
  title: string
  children?: React.ReactNode
  className?: string
}

export function ChartCard({ title, children, className }: ChartCardProps) {
  return (
    <Card className={cn("", className)}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {children ?? (
          <div className="flex h-[200px] items-center justify-center rounded-lg border border-dashed">
            <span className="text-muted-foreground text-sm">
              Chart placeholder
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
