"use client";

import { motion } from "framer-motion";

interface GridCellProps {
    label: string;
    value: string | number;
    unit?: string;
    delay?: number;
}

export function GridCell({ label, value, unit = "", delay = 0 }: GridCellProps) {
    return (
        <motion.div
            initial={{ opacity: 0, x: -20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay }}
            className="flex items-baseline gap-3 py-2 border-b border-white/5 last:border-0"
        >
            <span className="font-mono-tech text-xs text-gray-500 uppercase tracking-wider w-32">
                {label}
            </span>
            <span className="font-mono-tech text-sm text-gray-300">
                <span className="text-cyan-400">{value}</span>
                {unit && <span className="text-gray-600 ml-1">{unit}</span>}
            </span>
        </motion.div>
    );
}
