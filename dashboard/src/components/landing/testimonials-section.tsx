"use client";

import { motion } from "motion/react";
import { Star, Quote } from "lucide-react";

import { Marquee } from "@/components/ui/marquee";
import { TechBadge } from "@/components/ui/TechBadge";

const testimonials = [
  {
    content: "SwarmTest revolutionized our load testing. We went from manual scripts to AI-powered agents in minutes. The real-time visibility is incredible.",
    author: "Sarah Chen",
    role: "DevOps Lead",
    company: "TechFlow Inc.",
    rating: 5,
  },
  {
    content: "Finally, a load testing tool that actually makes sense. The autonomous agents find edge cases we never would have caught manually.",
    author: "Marcus Rodriguez",
    role: "Backend Engineer",
    company: "StripeScale",
    rating: 5,
  },
  {
    content: "We tested our API with 10,000 concurrent agents. SwarmTest handled it gracefully and the WebSocket streaming made debugging a breeze.",
    author: "Emily Zhang",
    role: "SRE Manager",
    company: "CloudNative Co",
    rating: 5,
  },
  {
    content: "The cyberpunk aesthetic is just the cherry on top. Underneath is a serious tool that delivers real results.",
    author: "James Wilson",
    role: "CTO",
    company: "StartupXYZ",
    rating: 5,
  },
  {
    content: "Deployed in 5 minutes, finding bugs within 10. SwarmTest pays for itself in the first week.",
    author: "Aisha Patel",
    role: "Engineering Manager",
    company: "FinTech Pro",
    rating: 5,
  },
  {
    content: "Our team loves the terminal interface. It feels like using a tool from the future.",
    author: "David Kim",
    role: "Full Stack Developer",
    company: "DevHouse",
    rating: 5,
  },
];

function TestimonialCard({
  testimonial,
}: {
  testimonial: (typeof testimonials)[0];
}) {
  return (
    <motion.div
      whileHover={{ y: -4 }}
      className="relative group"
    >
      <div className="absolute inset-0 bg-gradient-to-r from-cyan-500/20 to-teal-500/20 rounded-xl blur-xl opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
      <div className="relative bg-[#111] border border-white/10 rounded-xl p-6 h-full flex flex-col">
        {/* Quote icon */}
        <div className="absolute top-4 right-4 text-cyan-500/20">
          <Quote className="w-8 h-8" />
        </div>

        {/* Rating */}
        <div className="flex gap-1 mb-4">
          {Array.from({ length: testimonial.rating }).map((_, i) => (
            <Star
              key={i}
              className="w-4 h-4 fill-cyan-400 text-cyan-400"
            />
          ))}
        </div>

        {/* Content */}
        <p className="font-mono-tech text-sm text-gray-400 leading-relaxed mb-6 flex-1">
          {testimonial.content}
        </p>

        {/* Author */}
        <div className="flex items-center gap-3 border-t border-white/5 pt-4">
          <div className="w-10 h-10 bg-gradient-to-br from-cyan-500/20 to-teal-500/20 rounded-full flex items-center justify-center">
            <span className="font-mono-tech text-sm font-semibold text-cyan-400">
              {testimonial.author
                .split(" ")
                .map((n) => n[0])
                .join("")}
            </span>
          </div>
          <div>
            <div className="font-display-tech text-sm font-semibold text-gray-200">
              {testimonial.author}
            </div>
            <div className="font-mono-tech text-xs text-gray-600">
              {testimonial.role} @ {testimonial.company}
            </div>
          </div>
        </div>
      </div>
    </motion.div>
  );
}

export function TestimonialsSection() {
  return (
    <section className="py-32 px-6 border-t border-white/5 relative overflow-hidden">
      {/* Background effects */}
      <div className="absolute inset-0 tech-grid opacity-20 pointer-events-none" />

      <div className="relative z-10 max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
          className="text-center mb-16"
        >
          <TechBadge variant="outline">Social_Proof</TechBadge>
          <h2 className="font-display-tech text-3xl md:text-4xl font-bold mt-6 mb-4">
            <span className="text-gray-100">Trusted by</span>
            <span className="text-cyan-400">_</span>
            <span className="text-gray-500">Developers</span>
          </h2>
          <p className="font-mono-tech text-sm text-gray-500 max-w-2xl mx-auto">
            See what engineering teams are saying about SwarmTest.
            Real feedback from real users.
          </p>
        </motion.div>

        {/* Marquee with testimonials */}
        <div className="relative">
          {/* Fade edges */}
          <div className="absolute left-0 top-0 bottom-0 w-20 bg-gradient-to-r from-[#0a0a0a] to-transparent z-10 pointer-events-none" />
          <div className="absolute right-0 top-0 bottom-0 w-20 bg-gradient-to-l from-[#0a0a0a] to-transparent z-10 pointer-events-none" />

          {/* First row - left to right */}
          <Marquee pauseOnHover className="mb-6" duration={40}>
            {testimonials.map((testimonial) => (
              <div
                key={`${testimonial.author}-1`}
                className="w-[400px] px-3"
              >
                <TestimonialCard testimonial={testimonial} />
              </div>
            ))}
          </Marquee>

          {/* Second row - right to left */}
          <Marquee pauseOnHover reverse className="mb-6" duration={45}>
            {testimonials.map((testimonial) => (
              <div
                key={`${testimonial.author}-2`}
                className="w-[400px] px-3"
              >
                <TestimonialCard testimonial={testimonial} />
              </div>
            ))}
          </Marquee>

          {/* Third row - left to right */}
          <Marquee pauseOnHover duration={50}>
            {testimonials.map((testimonial) => (
              <div
                key={`${testimonial.author}-3`}
                className="w-[400px] px-3"
              >
                <TestimonialCard testimonial={testimonial} />
              </div>
            ))}
          </Marquee>
        </div>

        {/* Stats bar */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.2 }}
          className="mt-16 flex items-center justify-center gap-8 md:gap-16 flex-wrap"
        >
          {[
            { value: "10K+", label: "Active Users" },
            { value: "1M+", label: "Tests Run" },
            { value: "99.9%", label: "Uptime" },
            { value: "4.9★", label: "Average Rating" },
          ].map((stat) => (
            <div key={stat.label} className="text-center">
              <div className="font-display-tech text-2xl md:text-3xl font-bold text-cyan-400">
                {stat.value}
              </div>
              <div className="font-mono-tech text-xs text-gray-600 uppercase tracking-wider">
                {stat.label}
              </div>
            </div>
          ))}
        </motion.div>
      </div>
    </section>
  );
}
