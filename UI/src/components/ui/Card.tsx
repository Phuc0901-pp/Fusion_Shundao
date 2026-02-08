import { cn } from '../../utils/cn';

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
    children: React.ReactNode;
    className?: string;
    noPadding?: boolean;
}

export const Card = ({ children, className, noPadding = false, ...props }: CardProps) => {
    return (
        <div
            className={cn(
                "bg-white border border-slate-200 shadow-sm rounded-2xl overflow-hidden transition-all duration-300 hover:shadow-md",
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
    <h3 className={cn("text-lg font-semibold text-slate-900 flex items-center gap-2", className)}>
        {children}
    </h3>
);
