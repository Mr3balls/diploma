import { Link, Outlet } from "react-router-dom";
import { Trophy, Swords, Users, Shield } from "lucide-react";

const FEATURES = [
  { icon: Trophy,  text: "Турниры Single / Double / Group" },
  { icon: Swords,  text: "Сетка с SVG-коннекторами" },
  { icon: Users,   text: "Командная и индивидуальная регистрация" },
  { icon: Shield,  text: "Роли: организатор, менеджер, участник" },
];

export function AuthLayout() {
  return (
    <div className="flex min-h-screen bg-[#111111]">

      {/* ── Left branding panel ───────────────────────────────── */}
      <div className="relative hidden flex-col justify-between overflow-hidden border-r border-[#2d2d2d] bg-[#0d0d0d] p-10 lg:flex lg:w-[480px] xl:w-[540px] shrink-0">
        {/* orange glow */}
        <div
          aria-hidden
          style={{
            position: "absolute",
            top: "-80px",
            left: "-80px",
            width: "500px",
            height: "500px",
            background: "radial-gradient(ellipse at center, rgba(255,85,0,0.12) 0%, transparent 65%)",
            pointerEvents: "none",
          }}
        />

        {/* logo */}
        <Link to="/" className="flex items-center gap-3 z-10">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-[#ff5500]">
            <Trophy className="h-5 w-5 text-white" />
          </div>
          <span className="text-lg font-black uppercase tracking-widest text-white">ACE</span>
        </Link>

        {/* center content */}
        <div className="z-10 space-y-8">
          <div className="space-y-4">
            <p className="text-xs font-bold uppercase tracking-[0.35em] text-[#ff5500]">
              Esports Platform
            </p>
            <h2
              className="font-black uppercase text-white leading-none"
              style={{ fontSize: "clamp(2.2rem, 4vw, 3rem)", letterSpacing: "-0.03em" }}
            >
              Управляй
              <br />
              турнирами.
              <br />
              <span className="text-[#ff5500]">Побеждай.</span>
            </h2>
          </div>

          <ul className="space-y-3">
            {FEATURES.map(({ icon: Icon, text }) => (
              <li key={text} className="flex items-center gap-3 text-sm text-[#9e9e9e]">
                <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-[#1a1a1a] border border-[#2d2d2d]">
                  <Icon className="h-3.5 w-3.5 text-[#ff5500]" />
                </span>
                {text}
              </li>
            ))}
          </ul>
        </div>

        {/* bottom */}
        <p className="z-10 text-xs text-[#444444]">© 2025 ACE Esports Platform</p>
      </div>

      {/* ── Right form panel ──────────────────────────────────── */}
      <div className="flex flex-1 items-center justify-center px-4 py-12 sm:px-8">
        <div className="w-full max-w-md">
          {/* mobile logo */}
          <Link to="/" className="mb-8 flex items-center gap-2 lg:hidden">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-[#ff5500]">
              <Trophy className="h-4 w-4 text-white" />
            </div>
            <span className="text-sm font-black uppercase tracking-widest text-white">ACE</span>
          </Link>

          <Outlet />
        </div>
      </div>
    </div>
  );
}
