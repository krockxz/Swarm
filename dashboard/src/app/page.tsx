"use client";

import { motion } from "motion/react";
import Link from "next/link";
import {
  Terminal,
  ArrowRight,
  Github,
  Play,
  ChevronDown,
} from "lucide-react";

import { TechBadge } from "@/components/ui/TechBadge";
import { CornerBrackets } from "@/components/ui/CornerBrackets";
import { GridCell } from "@/components/ui/GridCell";
import { HeroSection } from "@/components/landing/hero-section";
import { TestimonialsSection } from "@/components/landing/testimonials-section";
import { TerminalEventsSection } from "@/components/landing/terminal-events-section";
import { BentoFeaturesSection } from "@/components/landing/bento-features-section";

// ============================================================================
// MAIN LANDING PAGE
// ============================================================================

export default function LandingPage() {
  const specs = [
    { label: "Backend", value: "Go", hint: "High-performance concurrency" },
    { label: "Frontend", value: "Next.js 15", hint: "React 19, App Router" },
    { label: "AI Engine", value: "Gemini 2.0", hint: "Flash decision-making" },
    { label: "Database", value: "PostgreSQL", hint: "Supabase hosted" },
    { label: "Protocol", value: "WebSocket", hint: "Real-time streaming" },
    { label: "License", value: "MIT", hint: "Open source" },
  ];

  return (
    <div className="min-h-screen bg-[#0a0a0a] text-gray-100 tech-noise">
      {/* Grid background overlay */}
      <div className="fixed inset-0 tech-grid pointer-events-none opacity-30" />

      {/* ============================================================================
          NAVIGATION
      ============================================================================ */}
      <motion.nav
        initial={{ y: -20, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        transition={{ duration: 0.5 }}
        className="fixed top-0 left-0 right-0 z-50 bg-[#0a0a0a]/80 backdrop-blur-md border-b border-white/5"
      >
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          <Link href="/" className="flex items-center gap-3 group">
            <div className="w-8 h-8 bg-cyan-500/10 border border-cyan-500/30 flex items-center justify-center group-hover:border-cyan-500/60 transition-colors">
              <Terminal className="w-4 h-4 text-cyan-400" />
            </div>
            <span className="font-mono-tech text-sm font-semibold tracking-tight">
              SwarmTest<span className="text-cyan-400">_</span>
            </span>
          </Link>

          <div className="flex items-center gap-8">
            <a
              href="#features"
              className="font-mono-tech text-xs text-gray-500 hover:text-gray-300 transition-colors"
            >
              [FEATURES]
            </a>
            <a
              href="#specs"
              className="font-mono-tech text-xs text-gray-500 hover:text-gray-300 transition-colors"
            >
              [SPECS]
            </a>
            <Link
              href="/dashboard"
              className="inline-flex items-center gap-2 px-4 py-2 bg-cyan-500/10 border border-cyan-500/30 font-mono-tech text-xs text-cyan-400 hover:bg-cyan-500/20 hover:border-cyan-500/50 transition-all tech-glitch"
            >
              <Terminal className="w-3.5 h-3.5" />
              LAUNCH_DASHBOARD
            </Link>
          </div>
        </div>
      </motion.nav>

      {/* ============================================================================
          HERO SECTION - Premium with Retro Grid Background
      ============================================================================ */}
      <HeroSection />

      {/* ============================================================================
          TESTIMONIALS SECTION - Infinite Marquee
      ============================================================================ */}
      <TestimonialsSection />

      {/* ============================================================================
          TERMINAL EVENTS SECTION - Animated Terminal
      ============================================================================ */}
      <TerminalEventsSection />

      {/* ============================================================================
          BENTO FEATURES SECTION - Hover Animations
      ============================================================================ */}
      <BentoFeaturesSection />

      {/* ============================================================================
          SPECS SECTION
      ============================================================================ */}
      <section id="specs" className="py-32 px-6 border-t border-white/5">
        <div className="max-w-7xl mx-auto">
          <div className="grid lg:grid-cols-2 gap-16">
            {/* Left side */}
            <motion.div
              initial={{ opacity: 0, x: -30 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5 }}
            >
              <TechBadge>System_Specs</TechBadge>
              <h2 className="font-display-tech text-3xl md:text-4xl font-bold mt-6 mb-4">
                <span className="text-gray-100">Technical</span>
                <span className="text-cyan-400">/</span>
                <span className="text-gray-500">Specs</span>
              </h2>
              <p className="font-mono-tech text-sm text-gray-500 leading-relaxed mb-8">
                Built with modern tools for modern problems.
                No legacy baggage. No unnecessary abstractions.
              </p>

              <div className="space-y-0">
                {specs.map((spec, index) => (
                  <GridCell
                    key={spec.label}
                    label={spec.label}
                    value={spec.value}
                    delay={index * 0.05}
                  />
                ))}
              </div>
            </motion.div>

            {/* Right side - Architecture diagram */}
            <motion.div
              initial={{ opacity: 0, x: 30 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5 }}
            >
              <CornerBrackets className="bg-[#111] border border-white/5 p-8 h-full">
                <h3 className="font-mono-tech text-xs text-gray-500 uppercase tracking-wider mb-6">
                  [DATA_FLOW]
                </h3>

                <div className="space-y-4">
                  {/* Flow steps */}
                  {[
                    { icon: <Terminal className="w-4 h-4" />, label: "CLI / Dashboard", active: true },
                    { icon: <Github className="w-4 h-4" />, label: "REST API → PostgreSQL", active: true },
                    { icon: <Play className="w-4 h-4" />, label: "Agent Spawner (Go)", active: true },
                    { icon: <ArrowRight className="w-4 h-4" />, label: "HTTP Target", active: false },
                  ].map((step, i) => (
                    <div key={step.label} className="flex items-center gap-4">
                      <div className={`p-2 ${step.active ? "bg-cyan-500/20 text-cyan-400" : "bg-white/5 text-gray-600"}`}>
                        {step.icon}
                      </div>
                      <div className="flex-1">
                        <div className="font-mono-tech text-xs text-gray-400">{step.label}</div>
                        <div className="h-px bg-white/5 mt-2" />
                      </div>
                      {i < 3 && (
                        <div className="text-cyan-500/50 font-mono-tech text-xs">↓</div>
                      )}
                    </div>
                  ))}

                  {/* WebSocket side channel */}
                  <div className="ml-6 pl-8 border-l border-dashed border-cyan-500/20 mt-4">
                    <div className="flex items-center gap-2 text-cyan-400/70 font-mono-tech text-xs">
                      <Play className="w-3 h-3" />
                      WebSocket (real-time events)
                    </div>
                  </div>
                </div>
              </CornerBrackets>
            </motion.div>
          </div>
        </div>
      </section>

      {/* ============================================================================
          CTA SECTION
      ============================================================================ */}
      <section className="py-32 px-6 bg-gradient-to-b from-[#0a0a0a] to-[#080808] border-t border-white/5">
        <div className="max-w-4xl mx-auto text-center">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
          >
            <motion.div
              className="inline-block mb-6"
              animate={{ opacity: [0.5, 1, 0.5] }}
              transition={{ duration: 2, repeat: Infinity }}
            >
              <TechBadge variant="outline">Ready_to_Deploy</TechBadge>
            </motion.div>

            <h2 className="font-display-tech text-4xl md:text-5xl font-bold mb-6">
              <span className="text-gray-100">Initialize Your</span>
              <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 to-teal-400 tech-glow">
                First Swarm
              </span>
            </h2>

            <p className="font-mono-tech text-sm text-gray-500 mb-10 max-w-xl mx-auto">
              Self-hosted. Open source. No vendor lock-in.
              Deploy locally or in your cloud environment.
            </p>

            <div className="flex flex-wrap justify-center gap-4">
              <Link
                href="/dashboard"
                className="group inline-flex items-center gap-3 px-8 py-4 bg-gradient-to-r from-cyan-500 to-teal-500 text-black font-mono-tech text-sm font-semibold hover:from-cyan-400 hover:to-teal-400 transition-all relative overflow-hidden"
              >
                <span className="absolute inset-0 bg-white/20 translate-y-full group-hover:translate-y-0 transition-transform duration-300" />
                <span className="relative flex items-center gap-3">
                  <Play className="w-4 h-4" />
                  LAUNCH_DASHBOARD
                  <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
                </span>
              </Link>

              <a
                href="https://github.com"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-3 px-8 py-4 bg-white/5 border border-white/10 font-mono-tech text-sm text-gray-300 hover:bg-white/10 hover:border-white/20 transition-colors"
              >
                <Github className="w-4 h-4" />
                CLONE_REPO
              </a>
            </div>

            <div className="mt-12 pt-8 border-t border-white/5">
              <code className="font-mono-tech text-xs text-gray-600">
                $ git clone https://github.com/swarmtest/swarmtest.git
              </code>
            </div>
          </motion.div>
        </div>
      </section>

      {/* ============================================================================
          FOOTER
      ============================================================================ */}
      <footer className="py-12 px-6 border-t border-white/5">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row items-center justify-between gap-6">
          <div className="flex items-center gap-3">
            <div className="w-6 h-6 bg-cyan-500/10 border border-cyan-500/30 flex items-center justify-center">
              <Terminal className="w-3 h-3 text-cyan-400" />
            </div>
            <span className="font-mono-tech text-sm text-gray-500">
              SwarmTest<span className="text-cyan-400">_</span>
            </span>
          </div>

          <div className="flex items-center gap-8 font-mono-tech text-xs text-gray-600">
            <a href="#features" className="hover:text-gray-300 transition-colors">
              [FEATURES]
            </a>
            <a href="#specs" className="hover:text-gray-300 transition-colors">
              [SPECS]
            </a>
            <a href="#" className="hover:text-gray-300 transition-colors">
              [DOCS]
            </a>
          </div>

          <div className="font-mono-tech text-xs text-gray-700">
            © 2025 :: MIT_LICENSE
          </div>
        </div>
      </footer>
    </div>
  );
}
