
import { Award } from 'lucide-react'

interface TierBadgeProps {
  nameAr: string
  color: string
  sortOrder: number
  size?: 'sm' | 'md' | 'lg'
}

export function TierBadge({ nameAr, color, sortOrder, size = 'md' }: TierBadgeProps) {
  const sizes = {
    sm: 'px-2 py-0.5 text-[10px] gap-1',
    md: 'px-2.5 py-1 text-xs gap-1',
    lg: 'px-3 py-1.5 text-sm gap-1.5',
  }
  const iconSizes = {
    sm: 'w-2.5 h-2.5',
    md: 'w-3 h-3',
    lg: 'w-3.5 h-3.5',
  }
  return (
    <span
      className={`inline-flex items-center rounded-full font-bold ${sizes[size]}`}
      style={{ backgroundColor: color, color: sortOrder >= 4 ? '#fff' : sortOrder === 1 ? '#374151' : '#fff' }}
    >
      <Award className={iconSizes[size]} />
      {nameAr}
    </span>
  )
}
