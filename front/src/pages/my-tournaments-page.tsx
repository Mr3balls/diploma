import { Link } from "react-router-dom";
import { LayoutList, Trophy, Crown, Shield } from "lucide-react";
import { useMyTournaments } from "@/features/profile/hooks";
import { useLang } from "@/app/providers/lang-provider";
import { Badge } from "@/shared/ui/badge";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
import { Spinner } from "@/shared/ui/spinner";
import { formatDate } from "@/shared/lib/date";
import type { MyTournamentEntry } from "@/shared/types/api";

const STATUS_TONE: Record<string, "default" | "success" | "danger" | "muted" | "warning"> = {
  registration_open: "warning",
  in_progress: "warning",
  finished: "success",
  completed: "success",
  cancelled: "danger",
  draft: "muted",
  registration_closed: "muted",
  bracket_generated: "default",
  ready: "default",
};

const ROLE_TONE: Record<string, "default" | "success" | "danger" | "muted" | "warning"> = {
  organizer: "warning",
  manager: "default",
  participant: "muted",
};

function TournamentRow({ entry }: { entry: MyTournamentEntry }) {
  const { t } = useLang();

  const ROLE_LABEL: Record<string, string> = {
    organizer: t("role.organizer"),
    manager: t("role.manager"),
    participant: t("role.participant"),
  };

  return (
    <Link
      to={`/tournaments/${entry.id}`}
      className="group flex flex-col gap-2 rounded-xl border border-[#2d2d2d] bg-[#111111] px-5 py-4 transition-colors hover:border-[#ff5500]/50 hover:bg-[#1a1a1a] sm:flex-row sm:items-center sm:justify-between"
    >
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-2 mb-1">
          {entry.is_winner && (
            <span className="flex items-center gap-1 text-[10px] font-bold uppercase tracking-wider text-[#f59e0b]">
              <Crown className="h-3 w-3" />
              {t("myTournaments.winner")}
            </span>
          )}
          <span className="text-sm font-semibold text-white group-hover:text-[#ff5500] transition-colors truncate">
            {entry.title}
          </span>
        </div>
        <div className="flex flex-wrap items-center gap-2 text-xs text-[#666666]">
          {entry.discipline && <span>{entry.discipline}</span>}
          <span>·</span>
          <span>{t(`format.${entry.format}`)}</span>
          {entry.start_at && (
            <>
              <span>·</span>
              <span>{formatDate(entry.start_at)}</span>
            </>
          )}
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-2 shrink-0">
        <Badge tone={ROLE_TONE[entry.user_role] ?? "muted"}>
          {entry.user_role === "organizer" && <Shield className="h-3 w-3 mr-1" />}
          {ROLE_LABEL[entry.user_role] ?? entry.user_role}
        </Badge>
        <Badge tone={STATUS_TONE[entry.status] ?? "muted"}>
          {t(`status.${entry.status}`)}
        </Badge>
      </div>
    </Link>
  );
}

export function MyTournamentsPage() {
  const { t } = useLang();
  const query = useMyTournaments();
  const items = query.data?.items ?? [];

  const organized = items.filter((t) => t.user_role === "organizer" || t.user_role === "manager");
  const participated = items.filter((t) => t.user_role === "participant");

  return (
    <div className="grid gap-0">

      {/* Banner */}
      <div
        style={{
          width: "100vw",
          marginLeft: "calc(50% - 50vw)",
          background: "#111111",
          borderBottom: "1px solid #2d2d2d",
        }}
      >
        <div className="mx-auto w-full max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[#ff5500]/10">
                <LayoutList className="h-5 w-5 text-[#ff5500]" />
              </div>
              <div>
                <h1
                  className="font-black uppercase text-white"
                  style={{ fontSize: "clamp(1.5rem, 4vw, 2.5rem)", letterSpacing: "-0.03em" }}
                >
                  {t("myTournaments.title")}
                </h1>
                <p className="text-sm text-[#666666]">
                  {items.length > 0
                    ? t("myTournaments.count", { n: items.length })
                    : t("myTournaments.none")}
                </p>
              </div>
            </div>

            {/* summary chips */}
            {items.length > 0 && (
              <div className="flex flex-wrap gap-3">
                <div className="flex items-center gap-2 rounded-lg border border-[#2d2d2d] bg-[#1a1a1a] px-3 py-2">
                  <Shield className="h-4 w-4 text-[#ff5500]" />
                  <span className="text-sm font-bold text-white">{organized.length}</span>
                  <span className="text-xs text-[#666666]">{t("myTournaments.organizedStat")}</span>
                </div>
                <div className="flex items-center gap-2 rounded-lg border border-[#2d2d2d] bg-[#1a1a1a] px-3 py-2">
                  <LayoutList className="h-4 w-4 text-[#ff5500]" />
                  <span className="text-sm font-bold text-white">{participated.length}</span>
                  <span className="text-xs text-[#666666]">{t("myTournaments.participatedStat")}</span>
                </div>
                <div className="flex items-center gap-2 rounded-lg border border-[#2d2d2d] bg-[#1a1a1a] px-3 py-2">
                  <Trophy className="h-4 w-4 text-[#f59e0b]" />
                  <span className="text-sm font-bold text-white">{items.filter((i) => i.is_winner).length}</span>
                  <span className="text-xs text-[#666666]">{t("myTournaments.winsStat")}</span>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="py-8 grid gap-8">
        {query.isLoading && <Spinner />}
        {query.isError && <ErrorState />}

        {!query.isLoading && !query.isError && items.length === 0 && (
          <EmptyState
            title={t("myTournaments.empty")}
            description={t("myTournaments.emptyDesc")}
          />
        )}

        {organized.length > 0 && (
          <section className="grid gap-3">
            <h2 className="flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-[#666666]">
              <Shield className="h-3.5 w-3.5 text-[#ff5500]" />
              {t("myTournaments.sectionOrg", { n: organized.length })}
            </h2>
            {organized.map((entry) => <TournamentRow key={entry.id} entry={entry} />)}
          </section>
        )}

        {participated.length > 0 && (
          <section className="grid gap-3">
            <h2 className="flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-[#666666]">
              <LayoutList className="h-3.5 w-3.5 text-[#ff5500]" />
              {t("myTournaments.sectionPart", { n: participated.length })}
            </h2>
            {participated.map((entry) => <TournamentRow key={entry.id} entry={entry} />)}
          </section>
        )}
      </div>
    </div>
  );
}
