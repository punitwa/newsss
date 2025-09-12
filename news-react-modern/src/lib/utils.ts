import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatTimeAgo(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSecs < 60) {
    return 'Just now';
  } else if (diffMins < 60) {
    return `${diffMins}m ago`;
  } else if (diffHours < 24) {
    return `${diffHours}h ago`;
  } else if (diffDays < 7) {
    return `${diffDays}d ago`;
  } else {
    return date.toLocaleDateString();
  }
}

export function getCategoryIcon(category: string): string {
  const icons: { [key: string]: string } = {
    technology: '💻',
    business: '💼',
    sports: '⚽',
    politics: '🏛️',
    health: '🏥',
    science: '🔬',
    entertainment: '🎬',
    world: '🌍',
    general: '📰',
  };
  return icons[category] || icons.general;
}

export function getCategoryColor(category: string): string {
  const colors: { [key: string]: string } = {
    technology: 'bg-blue-500',
    business: 'bg-green-500',
    sports: 'bg-orange-500',
    politics: 'bg-red-500',
    health: 'bg-purple-500',
    science: 'bg-cyan-500',
    entertainment: 'bg-pink-500',
    world: 'bg-emerald-500',
    general: 'bg-gray-500',
  };
  return colors[category] || colors.general;
}

export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength).trim() + '...';
}
