"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import StatusBadge from "./StatusBadge";

const navItems = [
  { href: "/ui/", label: "Home", icon: "📈" },
  { href: "/ui/config/", label: "Config", icon: "⚙️" },
  { href: "/ui/metrics/", label: "Metrics", icon: "📊" },
];

export default function Navbar() {
  const pathname = usePathname();

  return (
    <nav className="border-b border-slate-700 bg-slate-900/80 backdrop-blur-sm sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex h-14 items-center justify-between">
          <div className="flex items-center gap-8">
            <Link
              href="/ui/"
              className="text-lg font-bold text-sky-400 tracking-tight"
            >
              Stock Exporter
            </Link>
            <div className="flex items-center gap-1">
              {navItems.map((item) => {
                const isActive =
                  pathname === item.href ||
                  pathname === item.href.replace(/\/$/, "");
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                      isActive
                        ? "bg-sky-500/20 text-sky-300"
                        : "text-slate-400 hover:text-slate-200 hover:bg-slate-800"
                    }`}
                  >
                    <span className="mr-1.5">{item.icon}</span>
                    {item.label}
                  </Link>
                );
              })}
            </div>
          </div>
          <StatusBadge />
        </div>
      </div>
    </nav>
  );
}
