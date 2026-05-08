import { Outlet } from "react-router-dom";

export function AuthLayout() {
  return (
    <div className="min-h-screen bg-slate-100">
      <div className="mx-auto flex min-h-screen w-full max-w-6xl items-center justify-center px-4 py-10">
        <div className="grid w-full max-w-5xl gap-6 rounded-3xl border bg-white p-6 shadow-soft md:grid-cols-[1.1fr_0.9fr]">
          <div className="hidden rounded-2xl bg-slate-900 p-8 text-white md:block">
            <p className="text-sm uppercase tracking-[0.3em] text-slate-300">Diploma MVP</p>
            <h1 className="mt-4 text-3xl font-semibold leading-tight text-white">
              Платформа управления
              <br />
              киберспортивными турнирами
            </h1>
            <p className="mt-4 text-sm text-slate-300">
              Только реальные сценарии под текущий backend: авторизация, турниры, импорт из Google Sheets,
              сетка, матчи, подтверждения и уведомления.
            </p>
          </div>
          <div className="rounded-2xl bg-white">
            <Outlet />
          </div>
        </div>
      </div>
    </div>
  );
}