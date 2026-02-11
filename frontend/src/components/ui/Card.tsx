import { cn } from '../../utils/cn';

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
    children: React.ReactNode;
    className?: string;
    noPadding?: boolean;
    variant?: 'default' | 'glass' | 'gradient';
}

export const Card = ({ children, className, noPadding = false, variant = 'default', ...props }: CardProps) => {
    const variantStyles = {
        default: "bg-white border border-slate-200/80 shadow-sm",
        glass: "glass border-white/20 shadow-lg",
        gradient: "bg-gradient-to-br from-white to-slate-50 border border-slate-200/60 shadow-md"
    };

    return (
        <div
            className={cn(
                "rounded-2xl overflow-hidden transition-all duration-300",
                "hover:shadow-lg hover:shadow-slate-200/50",
                "hover-lift card-shine",
                variantStyles[variant],
                !noPadding && "p-6",
                className
            )}
            {...props}
        >
            {children}
        </div>
    );
};

export const CardHeader = ({ children, className }: { children: React.ReactNode; className?: string }) => (
    <div className={cn("flex items-center justify-between mb-4", className)}>
        {children}
    </div>
);

export const CardTitle = ({ children, className }: { children: React.ReactNode; className?: string }) => (
    <h3 className={cn("text-lg font-semibold text-slate-800 flex items-center gap-2", className)}>
        {children}
    </h3>
);
