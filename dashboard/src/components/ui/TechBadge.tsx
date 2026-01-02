import React from 'react';

interface TechBadgeProps {
    children: React.ReactNode;
    variant?: "default" | "outline";
}

export function TechBadge({ children, variant = "default" }: TechBadgeProps) {
    return (
        <span
            className={`inline-flex items-center gap-2 px-3 py-1 font-mono-tech text-xs uppercase tracking-wider ${variant === "outline"
                    ? "border border-cyan-500/30 text-cyan-400"
                    : "bg-cyan-500/10 text-cyan-400"
                }`}
        >
            <span className="w-1.5 h-1.5 bg-cyan-400 rounded-full animate-pulse" />
            {children}
        </span>
    );
}
