"use client";

import { useRef, useEffect } from "react";

export function SwarmVisualization() {
    const canvasRef = useRef<HTMLCanvasElement>(null);

    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;

        const ctx = canvas.getContext("2d");
        if (!ctx) return;

        // TypeScript needs this non-null assertion for the closure
        const context = ctx;

        const dpr = window.devicePixelRatio || 1;
        const rect = canvas.getBoundingClientRect();
        canvas.width = rect.width * dpr;
        canvas.height = rect.height * dpr;
        ctx.scale(dpr, dpr);

        const width = rect.width;
        const height = rect.height;

        // Agent particles
        const agents = Array.from({ length: 16 }, (_, i) => ({
            x: Math.random() * width,
            y: Math.random() * height,
            vx: (Math.random() - 0.5) * 2,
            vy: (Math.random() - 0.5) * 2,
            size: Math.random() * 3 + 2,
            status: Math.random() > 0.7 ? "active" : Math.random() > 0.5 ? "completed" : "pending",
        }));

        function animate() {
            context.fillStyle = "rgba(10, 10, 10, 0.1)";
            context.fillRect(0, 0, width, height);

            // Draw connections
            agents.forEach((agent, i) => {
                agents.forEach((other, j) => {
                    if (i >= j) return;
                    const dist = Math.hypot(agent.x - other.x, agent.y - other.y);
                    if (dist < 100) {
                        context.beginPath();
                        context.moveTo(agent.x, agent.y);
                        context.lineTo(other.x, other.y);
                        context.strokeStyle = `rgba(0, 240, 255, ${0.15 * (1 - dist / 100)})`;
                        context.lineWidth = 1;
                        context.stroke();
                    }
                });
            });

            // Draw agents
            agents.forEach((agent) => {
                agent.x += agent.vx;
                agent.y += agent.vy;

                if (agent.x < 0 || agent.x > width) agent.vx *= -1;
                if (agent.y < 0 || agent.y > height) agent.vy *= -1;

                const color =
                    agent.status === "active" ? "#00f0ff" :
                        agent.status === "completed" ? "#00ff94" :
                            "#444";

                context.beginPath();
                context.arc(agent.x, agent.y, agent.size, 0, Math.PI * 2);
                context.fillStyle = color;
                context.fill();

                // Glow effect for active agents
                if (agent.status === "active") {
                    context.beginPath();
                    context.arc(agent.x, agent.y, agent.size * 2, 0, Math.PI * 2);
                    context.fillStyle = "rgba(0, 240, 255, 0.2)";
                    context.fill();
                }
            });

            requestAnimationFrame(animate);
        }

        animate();

        return () => {
            // Cleanup
        };
    }, []);

    return (
        <div className="relative w-full h-48 bg-[#0a0a0a] border border-white/10 overflow-hidden">
            <canvas ref={canvasRef} className="absolute inset-0 w-full h-full" />
            <div className="absolute inset-0 tech-grid opacity-50" />
            <div className="absolute top-3 left-3 flex items-center gap-2">
                <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
                <span className="font-mono-tech text-xs text-gray-500">LIVE_SWARM</span>
            </div>
            <div className="absolute bottom-3 right-3 font-mono-tech text-xs text-gray-600">
                16_AGENTS_ACTIVE
            </div>
        </div>
    );
}
