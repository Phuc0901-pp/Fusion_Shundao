import { cn } from '../../utils/cn';

type BadgeVariant = 'default' | 'success' | 'warning' | 'error' | 'info' | 'outline';

interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
    variant?: BadgeVariant;
    children: React.ReactNode;
}

const variants: Record<BadgeVariant, string> = {
    default: "bg-slate-700 text-slate-200",
    success: "bg-green-500/10 text-green-400 border border-green-500/20",
    warning: "bg-amber-500/10 text-amber-400 border border-amber-500/20",
    error: "bg-red-500/10 text-red-400 border border-red-500/20",
    info: "bg-blue-500/10 text-blue-400 border border-blue-500/20",
    outline: "bg-transparent border border-slate-600 text-slate-400",
};

export const Badge = ({ className, variant = 'default', ...props }: BadgeProps) => {
    return (
        <span
            className={cn(
                "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
                variants[variant],
                className
            )}
            {...props}
        />
    );
};
