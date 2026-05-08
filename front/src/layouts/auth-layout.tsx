import { Outlet } from "react-router-dom";

export function AuthLayout() {
  return (
    <div className="min-h-screen bg-[#001538]">
      <div className="mx-auto flex min-h-screen w-full max-w-6xl items-center justify-center px-4 py-10">
        <div className="grid w-full max-w-5xl gap-6 rounded-3xl border border-[#0a3575] bg-[#001f52] p-6 shadow-soft md:grid-cols-[1.1fr_0.9fr]">
          <div className="hidden rounded-2xl bg-[#001538] p-8 text-white md:block">
            <p className="text-sm uppercase tracking-[0.3em] text-[#90b8ff]">Diploma MVP</p>
            <h1 className="mt-4 text-3xl font-semibold leading-tight text-white">
              Платформа управления
              <br />
              киберспортивными турнирами
            </h1>
            <p className="mt-4 text-sm text-[#90afd4]">
              Только реальные сценарии под текущий backend: авторизация, турниры, импорт из Google Sheets,
              сетка, матчи, подтверждения и уведомления.
            </p>
          </div>
          <div className="rounded-2xl bg-[#001f52]">
            <Outlet />
          </div>
        </div>
      </div>
    </div>
  );
}