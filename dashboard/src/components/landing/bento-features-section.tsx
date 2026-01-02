"use client";

import { motion } from "motion/react";
import { Cpu, Network, Zap, Activity, Shield, Globe, Code2, Database } from "lucide-react";

import { BentoGrid, BentoCard } from "@/components/ui/bento-grid";
import { TechBadge } from "@/components/ui/TechBadge";

// Feature backgrounds with subtle animations
const backgrounds = {
  grid: (
    <div className="absolute inset-0 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-[size:24px_24px]" />
  ),
  gradient: (
    <div className="absolute inset-0 bg-gradient-to-br from-cyan-500/10 via-transparent to-teal-500/10" />
  ),
  dots: (
    <div className="absolute inset-0 bg-[radial-gradient(circle_at_1px_1px,rgba(0,240,255,0.15)_1px,transparent_0)] [background-size:16px_16px]" />
  ),
  glow: (
    <div className="absolute inset-0 bg-gradient-to-r from-cyan-500/5 to-transparent blur-3xl" />
  ),
};

const features = [
  {
    Icon: Cpu,
    name: "AI-Driven Agents",
    description: "Gemini 2.0 Flash powers autonomous decision-making for each agent. No manual scripting required.",
    className: "md:col-span-2",
    background: backgrounds.grid,
    href: "#ai-agents",
    cta: "Learn more",
  },
  {
    Icon: Network,
    name: "Distributed Swarm",
    description: "Spawn thousands of concurrent agents navigating autonomously. Scale from 10 to 10,000+ agents.",
    className: "md:col-span-1",
    background: backgrounds.gradient,
    href: "#swarm",
    cta: "See scale",
  },
  {
    Icon: Zap,
    name: "HTTP-Native",
    description: "Zero browser overhead. Pure HTTP for maximum throughput. Sub-10ms response times.",
    className: "md:col-span-1",
    background: backgrounds.dots,
    href: "#http",
    cta: "View specs",
  },
  {
    Icon: Activity,
    name: "Real-Time Events",
    description: "WebSocket streaming delivers instant telemetry updates. Watch your swarm in action.",
    className: "md:col-span-2",
    background: backgrounds.glow,
    href: "#events",
    cta: "See demo",
  },
  {
    Icon: Shield,
    name: "Rate Limiting",
    description: "Token bucket protection per mission prevents server overload. Configure custom limits.",
    className: "md:col-span-1",
    background: backgrounds.gradient,
    href: "#rate-limit",
    cta: "Configure",
  },
  {
    Icon: Globe,
    name: "Static Parsing",
    description: "HTML extraction without JS execution. Fast, reliable, perfect for traditional web apps.",
    className: "md:col-span-1",
    background: backgrounds.dots,
    href: "#parsing",
    cta: "Learn more",
  },
  {
    Icon: Code2,
    name: "CLI + Dashboard",
    description: "Use the terminal for power users or the web dashboard for visual monitoring. Both included.",
    className: "md:col-span-2",
    background: backgrounds.grid,
    href: "#cli",
    cta: "Get started",
  },
  {
    Icon: Database,
    name: "Mission Persistence",
    description: "All missions, agents, and events stored in PostgreSQL. Full history and replay capability.",
    className: "md:col-span-1",
    background: backgrounds.glow,
    href: "#database",
    cta: "View schema",
  },
];

export function BentoFeaturesSection() {
  return (
    <section className="py-32 px-6 border-t border-white/5 relative overflow-hidden">
      {/* Background effects */}
      <div className="absolute inset-0 tech-grid opacity-10 pointer-events-none" />

      <div className="relative z-10 max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
          className="text-center mb-16"
        >
          <TechBadge variant="outline">Feature_Matrix</TechBadge>
          <h2 className="font-display-tech text-3xl md:text-4xl font-bold mt-6 mb-4">
            <span className="text-gray-100">Powerful</span>
            <span className="text-cyan-400">_</span>
            <span className="text-gray-500">Capabilities</span>
          </h2>
          <p className="font-mono-tech text-sm text-gray-500 max-w-2xl mx-auto">
            Everything you need for comprehensive load testing.
            Built for developers who value performance and reliability.
          </p>
        </motion.div>

        <BentoGrid className="max-w-6xl mx-auto">
          {features.map((feature, index) => (
            <motion.div
              key={feature.name}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: index * 0.05 }}
            >
              <BentoCard
                name={feature.name}
                className={feature.className}
                background={feature.background}
                Icon={feature.Icon}
                description={feature.description}
                href={feature.href}
                cta={feature.cta}
              />
            </motion.div>
          ))}
        </BentoGrid>

        {/* Code preview */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.4 }}
          className="mt-16 max-w-2xl mx-auto"
        >
          <div className="bg-[#111] border border-white/10 rounded-xl p-6 font-mono-tech text-sm">
            <div className="flex items-center gap-2 mb-4 pb-4 border-b border-white/5">
              <div className="w-3 h-3 rounded-full bg-red-500/50" />
              <div className="w-3 h-3 rounded-full bg-yellow-500/50" />
              <div className="w-3 h-3 rounded-full bg-green-500/50" />
            </div>
            <pre className="text-gray-400">
              <span className="text-green-400">$</span> swarmtest init my-test --agents=1000
              {"\n"}
              <span className="text-cyan-400">→</span> Mission initialized with{" "}
              <span className="text-cyan-400">1000</span> agents
              {"\n"}
              <span className="text-cyan-400">→</span> Dashboard:{" "}
              <span className="text-teal-400">http://localhost:3001</span>
            </pre>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
