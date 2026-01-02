"use client";

import { motion } from "motion/react";
import { Activity, ArrowRight, CheckCircle, XCircle, Clock } from "lucide-react";
import { useEffect, useState } from "react";

import { Terminal as TerminalComponent, TypingAnimation, AnimatedSpan } from "@/components/ui/terminal";
import { TechBadge } from "@/components/ui/TechBadge";

// Simulated realtime events
const mockEvents = [
  { type: "start", agent: "Agent-001", message: "Initialized mission 'checkout-flow-test'" },
  { type: "navigate", agent: "Agent-001", message: "Navigated to https://example.com/checkout", status: "success" },
  { type: "action", agent: "Agent-002", message: "Clicking 'Add to Cart' button", status: "success" },
  { type: "action", agent: "Agent-003", message: "Filling payment form", status: "pending" },
  { type: "navigate", agent: "Agent-004", message: "Navigated to https://example.com/products", status: "success" },
  { type: "action", agent: "Agent-005", message: "Clicking 'Buy Now' button", status: "error", error: "Element not found" },
  { type: "action", agent: "Agent-002", message: "Filling shipping address", status: "success" },
  { type: "navigate", agent: "Agent-006", message: "Navigated to https://example.com/cart", status: "success" },
  { type: "action", agent: "Agent-003", message: "Payment form submitted", status: "success" },
  { type: "action", agent: "Agent-007", message: "Clicking 'Apply Coupon' button", status: "pending" },
];

function EventIcon({ type, status }: { type: string; status?: string }) {
  if (type === "start") {
    return <Activity className="w-3.5 h-3.5 text-cyan-400" />;
  }
  if (type === "navigate") {
    return <ArrowRight className="w-3.5 h-3.5 text-gray-500" />;
  }
  if (status === "success") {
    return <CheckCircle className="w-3.5 h-3.5 text-green-400" />;
  }
  if (status === "error") {
    return <XCircle className="w-3.5 h-3.5 text-red-400" />;
  }
  return <Clock className="w-3.5 h-3.5 text-yellow-400" />;
}

function EventLine({ event, index }: { event: typeof mockEvents[0]; index: number }) {
  const timestamp = new Date().toISOString().split("T")[1].slice(0, 8);
  const statusColor = event.status === "success" ? "text-green-400" : event.status === "error" ? "text-red-400" : "text-yellow-400";

  return (
    <AnimatedSpan delay={index * 200} className="flex items-start gap-3 py-1">
      <span className="text-gray-600 shrink-0">[{timestamp}]</span>
      <EventIcon type={event.type} status={event.status} />
      <span className="text-cyan-400 shrink-0">{event.agent}:</span>
      <span className="text-gray-400">{event.message}</span>
      {event.error && <span className={statusColor}>{"// "}{event.error}</span>}
    </AnimatedSpan>
  );
}

export function TerminalEventsSection() {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true);
        }
      },
      { threshold: 0.3 }
    );

    const element = document.getElementById("terminal-section");
    if (element) observer.observe(element);

    return () => observer.disconnect();
  }, []);

  return (
    <section
      id="terminal-section"
      className="py-32 px-6 bg-[#080808] border-t border-white/5 relative overflow-hidden"
    >
      {/* Background effects */}
      <div className="absolute inset-0 tech-grid opacity-10 pointer-events-none" />

      <div className="relative z-10 max-w-6xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
          className="grid lg:grid-cols-2 gap-12 items-center"
        >
          {/* Left side - Description */}
          <div>
            <TechBadge>Realtime_Events</TechBadge>
            <h2 className="font-display-tech text-3xl md:text-4xl font-bold mt-6 mb-4">
              <span className="text-gray-100">Watch Your</span>
              <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 to-teal-400 tech-glow">
                Swarm in Action
              </span>
            </h2>
            <p className="font-mono-tech text-sm text-gray-500 leading-relaxed mb-8">
              Every agent action streamed in real-time via WebSocket.
              Debug issues as they happen. See the full picture.
            </p>

            <div className="space-y-4">
              {[
                { icon: <Activity className="w-4 h-4" />, title: "Live Streaming", desc: "WebSocket delivers instant updates" },
                { icon: <CheckCircle className="w-4 h-4" />, title: "Status Tracking", desc: "Success, error, and pending states" },
                { icon: <Clock className="w-4 h-4" />, title: "Timestamped Logs", desc: "Precise timing for analysis" },
              ].map((feature, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, x: -20 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true }}
                  transition={{ duration: 0.4, delay: i * 0.1 }}
                  className="flex items-start gap-4 p-4 bg-[#111] border border-white/5 rounded-lg group hover:border-cyan-500/20 transition-colors"
                >
                  <div className="p-2 bg-cyan-500/10 text-cyan-400 group-hover:bg-cyan-500/20 transition-colors">
                    {feature.icon}
                  </div>
                  <div>
                    <div className="font-display-tech text-sm font-semibold text-gray-200">
                      {feature.title}
                    </div>
                    <div className="font-mono-tech text-xs text-gray-600">
                      {feature.desc}
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          </div>

          {/* Right side - Terminal */}
          <motion.div
            initial={{ opacity: 0, x: 30 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
          >
            <div className="relative">
              {/* Glow effect */}
              <div className="absolute inset-0 bg-cyan-500/10 blur-3xl rounded-xl" />

              <TerminalComponent
                className="relative bg-[#0a0a0a] border border-white/10 rounded-xl overflow-hidden"
                startOnView={isVisible}
              >
                {/* Terminal content */}
                <div className="p-4 font-mono-tech text-sm space-y-1 min-h-[400px]">
                  <AnimatedSpan>
                    <TypingAnimation>
                      swarmtest init checkout-flow-test --agents=10 --rate=50
                    </TypingAnimation>
                  </AnimatedSpan>

                  <AnimatedSpan>
                    <span className="text-cyan-400">→ Mission created: checkout-flow-test</span>
                  </AnimatedSpan>
                  <AnimatedSpan>
                    <span className="text-gray-500">→ Spawning 10 agents with 50 req/min rate limit...</span>
                  </AnimatedSpan>

                  <div className="mt-4 pt-4 border-t border-white/5">
                    <AnimatedSpan>
                      <span className="text-gray-600">{"// Realtime event stream"}</span>
                    </AnimatedSpan>
                  </div>

                  {mockEvents.map((event, index) => (
                    <EventLine key={index} event={event} index={index} />
                  ))}

                  <AnimatedSpan delay={mockEvents.length * 200 + 500}>
                    <div className="flex items-center gap-2 mt-4 pt-4 border-t border-white/5">
                      <span className="text-green-400">$</span>
                      <span className="text-gray-700">_</span>
                      <span className="inline-block w-2 h-4 bg-cyan-400 animate-pulse" />
                    </div>
                  </AnimatedSpan>
                </div>
              </TerminalComponent>

              {/* Floating stats */}
              <div className="absolute -bottom-4 -right-4 bg-[#111] border border-white/10 px-4 py-2 font-mono-tech text-xs flex items-center gap-4">
                <div>
                  <span className="text-gray-600">events</span>
                  <span className="text-cyan-400 ml-2">{mockEvents.length}</span>
                </div>
                <div className="w-px h-4 bg-white/10" />
                <div>
                  <span className="text-gray-600">latency</span>
                  <span className="text-green-400 ml-2">8ms</span>
                </div>
                <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
              </div>
            </div>
          </motion.div>
        </motion.div>
      </div>
    </section>
  );
}
