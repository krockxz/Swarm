"use client";

import { motion } from "motion/react";
import { ArrowRight, Github, Play, ChevronDown } from "lucide-react";
import Link from "next/link";

import { RetroGrid } from "@/components/ui/retro-grid";
import { TechBadge } from "@/components/ui/TechBadge";
import { SwarmVisualization } from "@/components/viz/SwarmVisualization";
import { CornerBrackets } from "@/components/ui/CornerBrackets";

export function HeroSection() {
  return (
    <section className="relative min-h-screen flex items-center justify-center px-6 pt-16 overflow-hidden">
      {/* Retro Grid Background - subtle cyberpunk effect */}
      <div className="absolute inset-0 opacity-20 pointer-events-none">
        <RetroGrid
          className="absolute inset-0 [mask-image:radial-gradient(ellipse_at_center,black_40%,transparent_80%)]"
          cellSize={60}
          opacity={0.3}
          lightLineColor="rgba(0, 240, 255, 0.3)"
          darkLineColor="rgba(0, 240, 255, 0.1)"
          angle={65}
        />
      </div>

      {/* Gradient overlay for depth */}
      <div className="absolute inset-0 bg-gradient-to-b from-transparent via-[#0a0a0a]/50 to-[#0a0a0a] pointer-events-none" />

      {/* Ambient glow effect */}
      <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-cyan-500/10 rounded-full blur-[128px] pointer-events-none" />
      <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-teal-500/10 rounded-full blur-[128px] pointer-events-none" />

      {/* Scanlines overlay */}
      <div className="absolute inset-0 tech-scanlines pointer-events-none opacity-30" />

      <div className="relative z-10 max-w-6xl mx-auto w-full">
        {/* Status bar */}
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="flex items-center justify-between mb-8 pb-4 border-b border-white/5"
        >
          <div className="flex items-center gap-6">
            <TechBadge>System Online</TechBadge>
            <span className="font-mono-tech text-xs text-gray-600">
              v1.0.0 :: build.2025
            </span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
            <span className="font-mono-tech text-xs text-gray-600">ALL SYSTEMS NOMINAL</span>
          </div>
        </motion.div>

        {/* Main headline */}
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.2 }}
          >
            <div className="inline-flex items-center gap-2 px-3 py-1 bg-cyan-500/10 border border-cyan-500/20 rounded-full mb-6">
              <span className="w-1.5 h-1.5 bg-cyan-400 rounded-full animate-pulse" />
              <span className="font-mono-tech text-xs text-cyan-400">
                AI-POWERED LOAD TESTING
              </span>
            </div>

            <h1 className="font-display-tech text-5xl md:text-6xl lg:text-7xl font-bold leading-tight mb-6">
              <span className="text-gray-100">Distributed</span>
              <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 via-teal-400 to-emerald-400 tech-glow">
                Load Testing
              </span>
              <br />
              <span className="text-gray-100">Reimagined</span>
            </h1>

            <p className="font-mono-tech text-sm text-gray-500 mb-8 max-w-md leading-relaxed">
              Spawn thousands of AI-powered autonomous agents.
              Navigate. Test. Analyze. All in real-time.
              <span className="text-cyan-400 animate-pulse">_</span>
            </p>

            {/* CTA Buttons */}
            <div className="flex flex-wrap gap-4">
              <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }}>
                <Link
                  href="/dashboard"
                  className="group inline-flex items-center gap-3 px-6 py-3 bg-gradient-to-r from-cyan-500 to-teal-500 text-black font-mono-tech text-sm font-semibold hover:from-cyan-400 hover:to-teal-400 transition-all relative overflow-hidden"
                >
                  <span className="absolute inset-0 bg-white/20 translate-y-full group-hover:translate-y-0 transition-transform duration-300" />
                  <span className="relative flex items-center gap-3">
                    <Play className="w-4 h-4" />
                    START_TESTING
                    <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
                  </span>
                </Link>
              </motion.div>

              <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }}>
                <a
                  href="https://github.com"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-3 px-6 py-3 bg-white/5 border border-white/10 font-mono-tech text-sm text-gray-300 hover:bg-white/10 hover:border-white/20 transition-all group"
                >
                  <Github className="w-4 h-4 group-hover:rotate-12 transition-transform" />
                  VIEW_SOURCE
                </a>
              </motion.div>
            </div>

            {/* Quick stats */}
            <div className="mt-10 flex items-center gap-8">
              <div>
                <div className="text-2xl font-bold text-gray-100">10K+</div>
                <div className="font-mono-tech text-xs text-gray-600">Concurrent Agents</div>
              </div>
              <div className="w-px h-10 bg-white/10" />
              <div>
                <div className="text-2xl font-bold text-gray-100">&lt;10ms</div>
                <div className="font-mono-tech text-xs text-gray-600">Response Latency</div>
              </div>
              <div className="w-px h-10 bg-white/10" />
              <div>
                <div className="text-2xl font-bold text-gray-100">100%</div>
                <div className="font-mono-tech text-xs text-gray-600">Open Source</div>
              </div>
            </div>
          </motion.div>

          {/* Visualization */}
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.6, delay: 0.4 }}
            className="relative"
          >
            <CornerBrackets className="block">
              <SwarmVisualization />
            </CornerBrackets>

            {/* Floating stats card */}
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.8 }}
              className="absolute -bottom-6 -left-6 bg-[#111]/90 backdrop-blur border border-white/10 px-4 py-3 font-mono-tech text-xs"
            >
              <div className="flex items-center gap-4">
                <div>
                  <span className="text-gray-600">throughput</span>
                  <span className="text-cyan-400 ml-2">1000+</span>
                  <span className="text-gray-600 ml-1">req/s</span>
                </div>
                <div className="w-px h-4 bg-white/10" />
                <div>
                  <span className="text-gray-600">agents</span>
                  <span className="text-green-400 ml-2">unlimited</span>
                </div>
              </div>
            </motion.div>
          </motion.div>
        </div>
      </div>

      {/* Scroll indicator */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.5, delay: 1 }}
        className="absolute bottom-8 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2 text-gray-600"
      >
        <span className="font-mono-tech text-xs">SCROLL</span>
        <ChevronDown className="w-4 h-4 animate-bounce" />
      </motion.div>
    </section>
  );
}
