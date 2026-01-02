import React from 'react';

interface CornerBracketsProps {
    children: React.ReactNode;
    className?: string;
}

export function CornerBrackets({ children, className = "" }: CornerBracketsProps) {
    return (
        <div className={`tech-corner ${className}`}>
            {children}
        </div>
    );
}
