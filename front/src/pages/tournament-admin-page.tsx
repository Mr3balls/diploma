import { useEffect, useRef, useState } from "react";
import { useParams, Navigate, Link } from "react-router-dom";
import { ArrowLeft, Settings, Users, FileSpreadsheet, Trophy, Swords, ClipboardList } from "lucide-react";
import { toast } from "sonner";
import { Plus, Shuffle, Play, Trash2 } from "lucide-react";
import { useAuth } from "@/app/providers/auth-provider";
import { useTournamentAdminAccess } from "@/shared/hooks/use-tournament-admin-access";
import {
  useAddManager,
  useChangeTournamentStatus,
  useRemoveManager,
  useTournament,
  useTournamentAudit,
  useUpdateTournament,
  useTournamentParticipants,
  useAddTournamentParticipant,
  useBulkAddTournamentParticipants,
  useRemoveTournamentParticipant,
  useShuffleTournamentParticipants,
  useStartTournamentBracket,
} from "@/features/tournaments/hooks";
import { useAdminCreateTeam, useAdminDeleteTeam, useApproveTeam, useRejectTeam, useTeam, useTournamentAdminTeams, useRemoveMember } from "@/features/teams/hooks";
import { useGenerateBracket, useRegenerateBracket, useReseedBracket, useTournamentBracket, useAdvanceToPlayoff } from "@/features/bracket/hooks";
import {
  useApproveResult,
  useRejectResult,
  useScheduleMatch,
  useAdminSetResult,
  useTournamentAdminMatches,
} from "@/features/matches/hooks";
import { useConnectGoogleSheet, useValidateGoogleSheet, usePreviewImport, useConfirmImport, useTournamentImports } from "@/features/import/hooks";
import { GoogleSheetForm } from "@/features/import/components/google-sheet-form";
import { ImportPreviewTable } from "@/features/import/components/import-preview-table";
import { CreateTournamentForm } from "@/features/tournaments/components/create-tournament-form";
import { TeamDetailsCard } from "@/features/teams/components/team-details-card";
import { TeamsTable } from "@/features/teams/components/teams-table";
import { BracketView } from "@/features/bracket/components/bracket-view";
import { GroupStageView } from "@/features/bracket/components/group-stage-view";
import { GroupDEView } from "@/features/bracket/components/group-de-view";
import { ReseedBoard } from "@/features/bracket/components/reseed-board";
import { MatchesTable } from "@/features/matches/components/matches-table";
import type { Match } from "@/shared/types/api";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { managerSchema, tournamentStatusSchema, type ManagerFormValues, type TournamentFormValues } from "@/features/tournaments/schemas";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { FormField } from "@/shared/ui/form-field";
import { Input } from "@/shared/ui/input";
import { Select } from "@/shared/ui/select";
import { Button } from "@/shared/ui/button";
import { SectionCard } from "@/shared/ui/section";
import { cn } from "@/shared/lib/cn";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
import { Spinner } from "@/shared/ui/spinner";
import { formatDateTime } from "@/shared/lib/date";
import { deriveSeedOrderFromMatches, deriveSeedOrderFromTeams } from "@/shared/lib/bracket";
import { getErrorMessage } from "@/shared/lib/http";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";

const STATUS_LABELS: Record<string, string> = {
  draft: "Черновик · draft",
  registration_open: "Регистрация открыта · registration_open",
  registration_closed: "Регистрация закрыта · registration_closed",
  bracket_generated: "Сетка создана · bracket_generated",
  in_progress: "Идёт · in_progress",
  finished: "Завершён · finished",
  cancelled: "Отменён · cancelled",
  ready: "Готов к старту · ready",
  completed: "Завершён · completed",
};

function ManagerForm({
  onAdd,
  onRemove,
  isBusy,
}: {
  onAdd: (values: ManagerFormValues) => void;
  onRemove: (values: ManagerFormValues) => void;
  isBusy?: boolean;
}) {
  const form = useForm<ManagerFormValues>({
    resolver: zodResolver(managerSchema),
    defaultValues: { user_id: "" },
  });
  const { register, handleSubmit, formState: { errors } } = form;

  return (
    <form className="grid gap-4 md:grid-cols-[1fr_auto_auto]" onSubmit={handleSubmit(onAdd)}>
      <FormField label="UUID пользователя" error={errors.user_id?.message}>
        <Input {...register("user_id")} placeholder="Вставьте UUID пользователя" />
      </FormField>
      <div className="pt-7">
        <Button type="submit" size="sm" disabled={isBusy}>
          Добавить
        </Button>
      </div>
      <div className="pt-7">
        <Button type="button" variant="destructive" size="sm" disabled={isBusy} onClick={handleSubmit(onRemove)}>
          Удалить
        </Button>
      </div>
    </form>
  );
}

function StatusForm({
  currentStatus,
  onSubmit,
  isBusy,
}: {
  currentStatus: string;
  onSubmit: (status: string) => void;
  isBusy?: boolean;
}) {
  const form = useForm({
    resolver: zodResolver(tournamentStatusSchema),
    defaultValues: { status: currentStatus as any },
  });

  useEffect(() => {
    form.reset({ status: currentStatus as any });
  }, [currentStatus, form]);

  const options = Object.entries(STATUS_LABELS);
  return (
    <form
      className="flex flex-wrap items-end gap-4"
      onSubmit={form.handleSubmit((v) => onSubmit(v.status))}
    >
      <FormField label="Статус турнира">
        <Select {...form.register("status")}>
          {options.map(([value, label]) => (
            <option key={value} value={value}>{label}</option>
          ))}
        </Select>
      </FormField>
      <Button type="submit" size="sm" disabled={isBusy}>
        Обновить
      </Button>
    </form>
  );
}

function AdminCreateTeamForm({ tournamentId }: { tournamentId: string }) {
  const [open, setOpen] = useState(false);
  const [teamName, setTeamName] = useState("");
  const [members, setMembers] = useState(["", "", "", "", ""]);
  const createMutation = useAdminCreateTeam(tournamentId);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!teamName.trim()) { toast.error("Введите название команды"); return; }
    try {
      const filtered = members.map((m) => m.trim()).filter(Boolean);
      await createMutation.mutateAsync({ team_name: teamName.trim(), members: filtered });
      toast.success("Команда добавлена");
      setTeamName("");
      setMembers(["", "", "", "", ""]);
      setOpen(false);
    } catch (err) {
      toast.error(getErrorMessage(err));
    }
  }

  if (!open) {
    return (
      <Button variant="outline" size="sm" onClick={() => setOpen(true)}>
        + Добавить команду вручную
      </Button>
    );
  }

  return (
    <Card className="border-[#2d2d2d]">
      <CardHeader className="pb-2">
        <CardTitle className="text-base">Новая команда</CardTitle>
      </CardHeader>
      <CardContent>
        <form className="grid gap-4" onSubmit={handleSubmit}>
          <div className="grid gap-1">
            <label className="text-sm text-[#9e9e9e]">Название команды</label>
            <Input
              placeholder="Название команды"
              value={teamName}
              onChange={(e) => setTeamName(e.target.value)}
              className="md:max-w-sm"
            />
          </div>
          <div className="grid gap-2">
            <label className="text-sm text-[#9e9e9e]">Игроки (первый — капитан)</label>
            {members.map((m, i) => (
              <div key={i} className="flex gap-2 md:max-w-sm">
                <Input
                  placeholder={i === 0 ? "Капитан (игрок 1)" : `Игрок ${i + 1}`}
                  value={m}
                  onChange={(e) => setMembers((prev) => prev.map((x, j) => j === i ? e.target.value : x))}
                />
                {members.length > 1 && (
                  <Button type="button" variant="outline" size="sm" className="shrink-0"
                    onClick={() => setMembers((prev) => prev.filter((_, j) => j !== i))}>
                    ×
                  </Button>
                )}
              </div>
            ))}
            {members.length < 10 && (
              <Button type="button" variant="outline" size="sm" className="md:max-w-sm"
                onClick={() => setMembers((prev) => [...prev, ""])}>
                + Добавить игрока
              </Button>
            )}
          </div>
          <div className="flex gap-3">
            <Button type="submit" size="sm" disabled={createMutation.isPending}>
              {createMutation.isPending ? "Сохранение..." : "Создать"}
            </Button>
            <Button type="button" variant="outline" size="sm" onClick={() => setOpen(false)}>
              Отмена
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function IndividualParticipantPanel({ tournamentId, hideStart }: { tournamentId: string; hideStart?: boolean }) {
  const [name, setName] = useState("");
  const [bulkText, setBulkText] = useState("");
  const [showBulk, setShowBulk] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const participantsQuery = useTournamentParticipants(tournamentId);
  const addOne = useAddTournamentParticipant(tournamentId);
  const bulkAdd = useBulkAddTournamentParticipants(tournamentId);
  const remove = useRemoveTournamentParticipant(tournamentId);
  const shuffle = useShuffleTournamentParticipants(tournamentId);
  const start = useStartTournamentBracket(tournamentId);

  const participants = participantsQuery.data?.items ?? [];
  const canStart = participants.length >= 2;

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) return;
    try {
      await addOne.mutateAsync(trimmed);
      setName("");
      inputRef.current?.focus();
    } catch (err) {
      toast.error(getErrorMessage(err));
    }
  }

  async function handleBulkAdd() {
    const names = bulkText.split("\n").map((n) => n.trim()).filter(Boolean);
    if (!names.length) return;
    try {
      await bulkAdd.mutateAsync(names);
      setBulkText("");
      setShowBulk(false);
    } catch (err) {
      toast.error(getErrorMessage(err));
    }
  }

  async function handleStart() {
    if (!confirm("Начать турнир и сгенерировать сетку?")) return;
    try {
      await start.mutateAsync();
      toast.success("Сетка сгенерирована, турнир начался");
    } catch (err) {
      toast.error(getErrorMessage(err));
    }
  }

  return (
    <div className="space-y-4">
      <form onSubmit={handleAdd} className="flex gap-2">
        <Input
          ref={inputRef}
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Имя участника"
          className="flex-1"
        />
        <Button type="submit" size="sm" disabled={addOne.isPending || !name.trim()}>
          <Plus className="h-4 w-4" />
        </Button>
      </form>

      <Button variant="ghost" size="sm" className="w-full text-xs" onClick={() => setShowBulk((v) => !v)}>
        {showBulk ? "Скрыть" : "Добавить списком (несколько сразу)"}
      </Button>
      {showBulk && (
        <div className="space-y-2">
          <textarea
            className="w-full rounded-xl border border-[#2d2d2d] bg-[#111111] px-3 py-2 text-sm text-white placeholder-[#666666] focus:outline-none"
            rows={5}
            placeholder={"Один участник на строку:\nАлексей\nМихаил\nСергей"}
            value={bulkText}
            onChange={(e) => setBulkText(e.target.value)}
          />
          <Button size="sm" className="w-full" onClick={handleBulkAdd} disabled={bulkAdd.isPending || !bulkText.trim()}>
            {bulkAdd.isPending ? "Добавление..." : "Добавить всех"}
          </Button>
        </div>
      )}

      {participants.length > 0 && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm text-[#9e9e9e]">Участники: {participants.length}</span>
            <Button
              variant="secondary"
              size="sm"
              disabled={shuffle.isPending || participants.length < 2}
              onClick={() => void shuffle.mutateAsync().catch((err) => toast.error(getErrorMessage(err)))}
            >
              <Shuffle className="mr-1 h-3.5 w-3.5" />
              Перемешать
            </Button>
          </div>
          <ul className="divide-y divide-[#2d2d2d] rounded-xl border border-[#2d2d2d]">
            {[...participants].sort((a, b) => a.seed - b.seed).map((p) => (
              <li key={p.id} className="flex items-center justify-between gap-2 px-3 py-2">
                <div className="flex items-center gap-2 min-w-0">
                  <span className="w-5 text-right text-xs text-[#666666]">{p.seed}</span>
                  <span className="truncate text-sm text-white">{p.name}</span>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-6 w-6 shrink-0 p-0 text-[#666666] hover:text-red-400"
                  disabled={remove.isPending}
                  onClick={() => void remove.mutateAsync(p.id).catch((err) => toast.error(getErrorMessage(err)))}
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </Button>
              </li>
            ))}
          </ul>
        </div>
      )}

      {!hideStart && participants.length < 2 && (
        <p className="rounded-xl bg-yellow-900/20 px-3 py-2 text-xs text-yellow-400">
          Добавьте минимум 2 участника для запуска сетки
        </p>
      )}

      {!hideStart && (
        <Button
          className="w-full"
          disabled={!canStart || start.isPending}
          onClick={handleStart}
        >
          <Play className="mr-2 h-4 w-4" />
          {start.isPending ? "Генерация..." : "Сгенерировать сетку и начать"}
        </Button>
      )}
    </div>
  );
}

export function TournamentAdminPage() {
  const { id = "" } = useParams();
  const { user } = useAuth();

  const tournamentQuery = useTournament(id);
  const access = useTournamentAdminAccess(id, tournamentQuery.data);

  const teamsQuery = useTournamentAdminTeams(id, access.canAccessAdmin);
  const bracketQuery = useTournamentBracket(id);
  const matchesQuery = useTournamentAdminMatches(id, access.canAccessAdmin);
  const participantsQuery = useTournamentParticipants(id);
  const auditQuery = useTournamentAudit(id, access.canAccessAdmin);

  const updateTournamentMutation = useUpdateTournament(id);
  const changeStatusMutation = useChangeTournamentStatus(id);
  const addManagerMutation = useAddManager(id);
  const removeManagerMutation = useRemoveManager(id);

  const approveTeamMutation = useApproveTeam(id);
  const rejectTeamMutation = useRejectTeam(id);
  const removeMemberMutation = useRemoveMember(id);
  const deleteTeamMutation = useAdminDeleteTeam(id);

  const generateBracketMutation = useGenerateBracket(id);
  const regenerateBracketMutation = useRegenerateBracket(id);
  const reseedBracketMutation = useReseedBracket(id);
  const advanceToPlayoffMutation = useAdvanceToPlayoff(id);

  const approveResultMutation = useApproveResult(id);
  const rejectResultMutation = useRejectResult(id);
  const scheduleMatchMutation = useScheduleMatch(id);
  const adminSetResultMutation = useAdminSetResult(id);

  const connectSheetMutation = useConnectGoogleSheet(id);
  const validateSheetMutation = useValidateGoogleSheet(id);
  const previewImportMutation = usePreviewImport(id);
  const confirmImportMutation = useConfirmImport(id);
  const importsQuery = useTournamentImports(id, access.canAccessAdmin);
  const [importPreview, setImportPreview] = useState<import("@/shared/types/api").ImportPreviewResponse | null>(null);

  const [activeSection, setActiveSection] = useState("settings");
  const [selectedTeamId, setSelectedTeamId] = useState<string | null>(null);
  const selectedTeamQuery = useTeam(selectedTeamId ?? undefined, Boolean(selectedTeamId) && access.canAccessAdmin);

  const [reseedItems, setReseedItems] = useState<{ id: string; label: string }[]>([]);

  const [scheduleMatch, setScheduleMatch] = useState<Match | null>(null);
  const [scheduleAt, setScheduleAt] = useState("");
  const [resultMatch, setResultMatch] = useState<Match | null>(null);
  const [resultWinnerId, setResultWinnerId] = useState("");
  const [resultScore, setResultScore] = useState("");

  useEffect(() => {
    const fromMatches =
      bracketQuery.data?.matches?.length
        ? deriveSeedOrderFromMatches(bracketQuery.data.matches).map((teamId) => ({
            id: teamId,
            label: teamsQuery.data?.items.find((t) => t.id === teamId)?.name || teamId,
          }))
        : [];
    const fallback =
      teamsQuery.data?.items?.length
        ? deriveSeedOrderFromTeams(teamsQuery.data.items).map((teamId) => ({
            id: teamId,
            label: teamsQuery.data?.items.find((t) => t.id === teamId)?.name || teamId,
          }))
        : [];
    setReseedItems(fromMatches.length ? fromMatches : fallback);
  }, [bracketQuery.data?.matches, teamsQuery.data?.items]);

  if (tournamentQuery.isLoading || access.isLoading) return <Spinner />;
  if (tournamentQuery.isError || !tournamentQuery.data) return <ErrorState />;
  if (!user || !access.canAccessAdmin) {
    return <Navigate to={`/tournaments/${id}`} replace />;
  }

  const tournament = tournamentQuery.data;
  const isIndividual = tournament.registration_mode === "individual";
  const canReseed =
    tournament.status !== "in_progress" &&
    tournament.status !== "finished" &&
    tournament.status !== "cancelled";
  const teamsById = new Map((teamsQuery.data?.items ?? []).map((t) => [t.id, t]));
  const participantsById = new Map((participantsQuery.data?.items ?? []).map((p) => [p.id, p]));

  function toDatetimeLocal(iso: string) {
    const d = new Date(iso);
    const pad = (n: number) => String(n).padStart(2, "0");
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
  }

  function handleScheduleOpen(match: Match) {
    setScheduleMatch(match);
    setScheduleAt(match.scheduled_at ? toDatetimeLocal(match.scheduled_at) : "");
  }

  async function handleScheduleSubmit() {
    if (!scheduleMatch || !scheduleAt) return;
    try {
      const isoDate = new Date(scheduleAt).toISOString();
      await scheduleMatchMutation.mutateAsync({ matchId: scheduleMatch.id, payload: { scheduled_at: isoDate } });
      toast.success("Время установлено");
      setScheduleMatch(null);
    } catch (err) {
      toast.error(getErrorMessage(err));
    }
  }

  function handleResultOpen(match: Match) {
    setResultMatch(match);
    setResultWinnerId("");
    setResultScore("");
  }

  async function handleResultSubmit() {
    if (!resultMatch || !resultWinnerId) return;
    try {
      const payload = isIndividual
        ? { winner_participant_id: resultWinnerId, score_text: resultScore || undefined }
        : { winner_team_id: resultWinnerId, score_text: resultScore || undefined };
      await adminSetResultMutation.mutateAsync({ matchId: resultMatch.id, payload });
      toast.success("Победитель установлен");
      setResultMatch(null);
    } catch (err) {
      toast.error(getErrorMessage(err));
    }
  }

  async function handleUpdateTournament(values: TournamentFormValues) {
    try {
      await updateTournamentMutation.mutateAsync(values);
      toast.success("Настройки сохранены");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleStatusUpdate(status: string) {
    try {
      await changeStatusMutation.mutateAsync({ status: status as any });
      toast.success("Статус обновлён");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleAddManager(values: ManagerFormValues) {
    try {
      await addManagerMutation.mutateAsync(values);
      toast.success("Со-организатор добавлен");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleRemoveManager(values: ManagerFormValues) {
    try {
      await removeManagerMutation.mutateAsync(values.user_id);
      toast.success("Со-организатор удалён");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleApproveTeam(teamId: string) {
    try {
      await approveTeamMutation.mutateAsync(teamId);
      toast.success("Команда одобрена");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleRejectTeam(teamId: string) {
    const reason = window.prompt("Укажите причину отклонения:");
    if (!reason || reason.trim().length < 2) return;
    try {
      await rejectTeamMutation.mutateAsync({ teamId, reason: reason.trim() });
      toast.success("Команда отклонена");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleRemoveMember(memberId: string) {
    if (!selectedTeamId) return;
    try {
      await removeMemberMutation.mutateAsync({ teamId: selectedTeamId, memberId });
      toast.success("Участник удалён");
      await selectedTeamQuery.refetch();
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleDeleteTeam(teamId: string) {
    if (!window.confirm("Удалить команду? Это действие нельзя отменить.")) return;
    try {
      await deleteTeamMutation.mutateAsync(teamId);
      if (selectedTeamId === teamId) setSelectedTeamId(null);
      toast.success("Команда удалена");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleGenerateBracket() {
    try {
      await generateBracketMutation.mutateAsync();
      toast.success("Сетка создана");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleRegenerateBracket() {
    try {
      await regenerateBracketMutation.mutateAsync();
      toast.success("Сетка пересоздана");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleAdvanceToPlayoff() {
    if (!confirm("Перейти к плей-офф? Групповые матчи будут завершены.")) return;
    try {
      await advanceToPlayoffMutation.mutateAsync();
      toast.success("Плей-офф начался!");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleSaveReseed() {
    try {
      await reseedBracketMutation.mutateAsync({ ordered_team_ids: reseedItems.map((item) => item.id) });
      toast.success("Порядок посева сохранён");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleApproveResult(matchId: string) {
    try {
      await approveResultMutation.mutateAsync(matchId);
      toast.success("Результат подтверждён");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleRejectResult(matchId: string) {
    try {
      await rejectResultMutation.mutateAsync({ matchId });
      toast.success("Результат отклонён");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleConnectSheet(values: import("@/features/import/schemas").GoogleSheetFormValues) {
    try {
      await connectSheetMutation.mutateAsync(values);
      toast.success("Таблица привязана");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleValidateSheet(values: import("@/features/import/schemas").GoogleSheetFormValues) {
    try {
      await validateSheetMutation.mutateAsync(values);
      toast.success("Таблица прошла проверку");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handlePreviewImport(values: import("@/features/import/schemas").GoogleSheetFormValues) {
    try {
      const result = await previewImportMutation.mutateAsync(values);
      setImportPreview(result);
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleConfirmImport(batchId: string) {
    try {
      await confirmImportMutation.mutateAsync({ batch_id: batchId });
      toast.success("Импорт подтверждён");
      setImportPreview(null);
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  const sheetIsBusy =
    connectSheetMutation.isPending ||
    validateSheetMutation.isPending ||
    previewImportMutation.isPending ||
    confirmImportMutation.isPending;

  const winnerName = tournament.winner_team_id
    ? (teamsQuery.data?.items.find((t) => t.id === tournament.winner_team_id)?.name ?? null)
    : tournament.winner_participant_id
    ? (participantsQuery.data?.items.find((p) => p.id === tournament.winner_participant_id)?.name ?? null)
    : null;

  const NAV = [
    { key: "settings",     label: "Настройки",   icon: Settings },
    ...(isIndividual
      ? [{ key: "participants", label: "Участники", icon: Users }]
      : [
          { key: "teams",  label: "Команды",  icon: Users },
          { key: "import", label: "Импорт",   icon: FileSpreadsheet },
        ]
    ),
    { key: "bracket", label: "Сетка",   icon: Trophy },
    { key: "matches", label: "Матчи",   icon: Swords },
    { key: "audit",   label: "Аудит",   icon: ClipboardList },
  ];

  return (
    <div className="grid gap-0">

      {/* ── Banner ──────────────────────────────────────────────── */}
      <div style={{ width: "100vw", marginLeft: "calc(50% - 50vw)", background: "#111111", borderBottom: "1px solid #2d2d2d" }}>
        <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-6">
          <Link to={`/tournaments/${id}`} className="inline-flex items-center gap-1.5 text-xs text-[#666666] hover:text-[#ff5500] transition-colors">
            <ArrowLeft className="h-3.5 w-3.5" /> Вернуться к турниру
          </Link>
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div>
              <p className="mb-1 text-xs font-bold uppercase tracking-[0.35em] text-[#ff5500]">Управление</p>
              <h1 className="font-black uppercase text-white" style={{ fontSize: "clamp(1.5rem, 4vw, 2.5rem)", letterSpacing: "-0.03em" }}>
                {tournament.title}
              </h1>
            </div>
            {winnerName && (
              <div className="flex items-center gap-2 rounded-xl border border-yellow-500/30 bg-yellow-500/10 px-4 py-2.5">
                <Trophy className="h-4 w-4 text-yellow-400" />
                <div>
                  <p className="text-[10px] font-bold uppercase tracking-wide text-yellow-500">Победитель</p>
                  <p className="text-sm font-semibold text-white">{winnerName}</p>
                </div>
              </div>
            )}
          </div>

          {/* tab nav */}
          <div className="flex gap-0 border-b border-[#2d2d2d] overflow-x-auto">
            {NAV.map(({ key, label, icon: Icon }) => (
              <button
                key={key}
                onClick={() => setActiveSection(key)}
                className={cn(
                  "relative flex items-center gap-2 px-4 py-3 text-sm font-semibold whitespace-nowrap transition-colors",
                  activeSection === key
                    ? "text-white after:absolute after:bottom-0 after:left-0 after:right-0 after:h-0.5 after:bg-[#ff5500]"
                    : "text-[#666666] hover:text-[#9e9e9e]",
                )}
              >
                <Icon className="h-4 w-4" />
                {label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* ── Content ─────────────────────────────────────────────── */}
      <div className="py-8 grid gap-6">

      {activeSection === "settings" && <SectionCard title="Основные настройки" description="Редактирование параметров и смена статуса.">
        <div className="grid gap-6">
          <CreateTournamentForm
            defaultValues={{
              title: tournament.title,
              description: tournament.description ?? "",
              rules: tournament.rules ?? "",
              discipline: tournament.discipline ?? "",
              max_teams: tournament.max_teams ?? 8,
              visibility: tournament.visibility,
            }}
            submitLabel="Сохранить"
            onSubmit={handleUpdateTournament}
            isSubmitting={updateTournamentMutation.isPending}
            showAdvanced
          />
          <StatusForm
            currentStatus={tournament.status}
            onSubmit={handleStatusUpdate}
            isBusy={changeStatusMutation.isPending}
          />
        </div>
      </SectionCard>}

      {activeSection === "settings" && <SectionCard
        title="Со-организаторы"
        description="Добавьте пользователей, которые смогут управлять турниром наравне с вами."
      >
        <ManagerForm
          onAdd={handleAddManager}
          onRemove={handleRemoveManager}
          isBusy={addManagerMutation.isPending || removeManagerMutation.isPending}
        />
      </SectionCard>}

      {tournament.registration_mode === "individual" ? (
        <>
          {activeSection === "participants" && <SectionCard
            title="Участники"
            description="Добавьте участников по имени, перемешайте посев и запустите сетку."
          >
            <IndividualParticipantPanel tournamentId={id} />
          </SectionCard>}
          {activeSection === "bracket" && (bracketQuery.data?.matches ?? []).length > 0 && (
            <SectionCard title="Сетка" description="Турнирная сетка по участникам.">
              <BracketView
                matches={bracketQuery.data?.matches ?? []}
                participants={participantsQuery.data?.items ?? []}
                adminMode
                tournamentId={id}
              />
            </SectionCard>
          )}
        </>
      ) : (
        <>
          {activeSection === "import" && <SectionCard
            title="Импорт из Google Sheets"
            description="Загрузите команды из публичной таблицы Google Sheets."
          >
            <div className="grid gap-4">
              <GoogleSheetForm
                onConnect={handleConnectSheet}
                onValidate={handleValidateSheet}
                onPreview={handlePreviewImport}
                isBusy={sheetIsBusy}
              />
              {importPreview && (
                <ImportPreviewTable
                  preview={importPreview}
                  onConfirm={handleConfirmImport}
                  isConfirming={confirmImportMutation.isPending}
                />
              )}
              {importsQuery.data?.items.length ? (
                <div className="text-xs text-[#9e9e9e]">
                  Последних импортов: {importsQuery.data.items.length}
                </div>
              ) : null}
            </div>
          </SectionCard>}

          {activeSection === "teams" && <SectionCard title="Команды" description="Одобрение и управление зарегистрированными командами.">
            <div className="grid gap-4">
              <AdminCreateTeamForm tournamentId={id} />
              {teamsQuery.isLoading ? (
                <Spinner />
              ) : teamsQuery.isError ? (
                <ErrorState />
              ) : teamsQuery.data?.items.length ? (
                <TeamsTable
                  teams={teamsQuery.data.items}
                  withActions
                  onOpen={setSelectedTeamId}
                  onApprove={handleApproveTeam}
                  onReject={handleRejectTeam}
                  onDelete={handleDeleteTeam}
                />
              ) : (
                <EmptyState title="Команд нет" description="Пока никто не зарегистрировался." />
              )}
              {selectedTeamId && selectedTeamQuery.isLoading ? <Spinner /> : null}
              {selectedTeamQuery.data && selectedTeamId ? (
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <p className="text-xs text-[#666666]">Состав команды</p>
                    <Button variant="ghost" size="sm" className="text-xs text-[#666666]" onClick={() => setSelectedTeamId(null)}>
                      Закрыть
                    </Button>
                  </div>
                  <TeamDetailsCard
                    data={selectedTeamQuery.data}
                    allowAdminActions
                    onRemoveMember={handleRemoveMember}
                  />
                </div>
              ) : null}
            </div>
          </SectionCard>}

          {activeSection === "bracket" && <SectionCard
            title="Сетка"
            description="Генерация сетки по зарегистрированным командам. Перегенерация сбрасывает текущую сетку."
            actions={
              <>
                <Button onClick={handleGenerateBracket} size="sm" disabled={generateBracketMutation.isPending}>
                  Сгенерировать
                </Button>
                <Button variant="outline" size="sm" onClick={handleRegenerateBracket} disabled={regenerateBracketMutation.isPending}>
                  Пересоздать
                </Button>
              </>
            }
          >
            {bracketQuery.isLoading ? (
              <Spinner />
            ) : bracketQuery.isError ? (
              <ErrorState />
            ) : bracketQuery.data?.bracket?.format === "group_stage" ? (
              <div className="space-y-4">
                {bracketQuery.data.bracket.status === "playoff" && (bracketQuery.data.matches ?? []).filter((m) => !m.group_id).length > 0 && (
                  <div className="space-y-2">
                    <h3 className="text-sm font-semibold uppercase tracking-wide text-[#9e9e9e]">Плей-офф</h3>
                    <BracketView
                      matches={(bracketQuery.data.matches ?? []).filter((m) => !m.group_id)}
                      teams={teamsQuery.data?.items ?? []}
                      adminMode
                      tournamentId={id}
                    />
                  </div>
                )}
                {(bracketQuery.data.groups ?? []).length > 0 ? (
                  <GroupStageView
                    groups={bracketQuery.data.groups ?? []}
                    matches={bracketQuery.data.matches ?? []}
                    teams={teamsQuery.data?.items ?? []}
                    adminMode
                    tournamentId={id}
                  />
                ) : (
                  <EmptyState title="Сетка не сгенерирована" description="Нажмите «Сгенерировать» для создания группового этапа." />
                )}
                {bracketQuery.data.bracket.status !== "playoff" && (bracketQuery.data.groups ?? []).length > 0 && (
                  <Button
                    className="w-full"
                    disabled={advanceToPlayoffMutation.isPending}
                    onClick={() => void handleAdvanceToPlayoff()}
                  >
                    {advanceToPlayoffMutation.isPending ? "Переход..." : "Перейти к плей-офф →"}
                  </Button>
                )}
              </div>
            ) : bracketQuery.data?.bracket?.format === "group_de" ? (
              <div className="space-y-4">
                {/* Playoff bracket after advancement */}
                {bracketQuery.data.bracket.status === "playoff" && (bracketQuery.data.matches ?? []).filter((m) => !m.group_id).length > 0 && (
                  <div className="space-y-2">
                    <h3 className="text-sm font-semibold uppercase tracking-wide text-[#9e9e9e]">Плей-офф</h3>
                    <BracketView
                      matches={(bracketQuery.data.matches ?? []).filter((m) => !m.group_id)}
                      teams={teamsQuery.data?.items ?? []}
                      adminMode
                      tournamentId={id}
                    />
                  </div>
                )}
                {/* Per-group DE brackets */}
                {(bracketQuery.data.groups ?? []).length > 0 ? (
                  <GroupDEView
                    groups={bracketQuery.data.groups ?? []}
                    matches={bracketQuery.data.matches ?? []}
                    teams={teamsQuery.data?.items ?? []}
                    adminMode
                    tournamentId={id}
                  />
                ) : (
                  <EmptyState title="Сетка не сгенерирована" description="Нажмите «Сгенерировать» для создания группового DE этапа." />
                )}
                {/* Advance to playoff once all groups have 3 qualified teams each */}
                {bracketQuery.data.bracket.status !== "playoff" && (bracketQuery.data.groups ?? []).length > 0 && (
                  <Button
                    className="w-full"
                    disabled={advanceToPlayoffMutation.isPending}
                    onClick={() => void handleAdvanceToPlayoff()}
                  >
                    {advanceToPlayoffMutation.isPending ? "Переход..." : "Перейти к плей-офф →"}
                  </Button>
                )}
              </div>
            ) : (
              <BracketView
                matches={bracketQuery.data?.matches ?? []}
                teams={teamsQuery.data?.items ?? []}
                adminMode
                tournamentId={id}
              />
            )}
          </SectionCard>}

          {activeSection === "bracket" && <SectionCard title="Посев" description="Перетащите команды для изменения порядка посева.">
            <ReseedBoard
              items={reseedItems}
              onChange={setReseedItems}
              onSave={handleSaveReseed}
              disabled={!canReseed}
              saving={reseedBracketMutation.isPending}
            />
          </SectionCard>}
        </>
      )}

      {activeSection === "matches" && <SectionCard title="Матчи" description="Просмотр и управление результатами матчей.">
        {scheduleMatch && (
          <Card className="mb-4 border-[#2d2d2d]">
            <CardHeader className="pb-2">
              <CardTitle className="text-base">Время матча — Раунд {scheduleMatch.round_number ?? "—"} · Слот {scheduleMatch.slot_index ?? "—"}</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4">
              <div className="grid gap-1">
                <label className="text-sm text-[#9e9e9e]">Дата и время</label>
                <Input type="datetime-local" value={scheduleAt} onChange={(e) => setScheduleAt(e.target.value)} className="md:max-w-sm" />
              </div>
              <div className="flex gap-3">
                <Button size="sm" disabled={scheduleMatchMutation.isPending || !scheduleAt} onClick={() => void handleScheduleSubmit()}>
                  {scheduleMatchMutation.isPending ? "Сохранение..." : "Сохранить"}
                </Button>
                <Button size="sm" variant="outline" onClick={() => setScheduleMatch(null)}>Отмена</Button>
              </div>
            </CardContent>
          </Card>
        )}
        {resultMatch && (
          <Card className="mb-4 border-[#2d2d2d]">
            <CardHeader className="pb-2">
              <CardTitle className="text-base">Победитель — Раунд {resultMatch.round_number ?? "—"} · Слот {resultMatch.slot_index ?? "—"}</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4">
              <div className="grid gap-2">
                <p className="text-sm text-[#9e9e9e]">Выберите победителя:</p>
                {isIndividual ? (
                  <>
                    {resultMatch.participant1_id && (
                      <Button
                        size="sm"
                        variant={resultWinnerId === resultMatch.participant1_id ? "default" : "outline"}
                        className="justify-start"
                        onClick={() => setResultWinnerId(resultMatch.participant1_id!)}
                      >
                        {participantsById.get(resultMatch.participant1_id)?.name ?? resultMatch.participant1_id.slice(0, 8)}
                      </Button>
                    )}
                    {resultMatch.participant2_id && (
                      <Button
                        size="sm"
                        variant={resultWinnerId === resultMatch.participant2_id ? "default" : "outline"}
                        className="justify-start"
                        onClick={() => setResultWinnerId(resultMatch.participant2_id!)}
                      >
                        {participantsById.get(resultMatch.participant2_id)?.name ?? resultMatch.participant2_id.slice(0, 8)}
                      </Button>
                    )}
                  </>
                ) : (
                  <>
                    {resultMatch.team1_id && (
                      <Button
                        size="sm"
                        variant={resultWinnerId === resultMatch.team1_id ? "default" : "outline"}
                        className="justify-start"
                        onClick={() => setResultWinnerId(resultMatch.team1_id!)}
                      >
                        {teamsById.get(resultMatch.team1_id)?.name ?? resultMatch.team1_id.slice(0, 8)}
                      </Button>
                    )}
                    {resultMatch.team2_id && (
                      <Button
                        size="sm"
                        variant={resultWinnerId === resultMatch.team2_id ? "default" : "outline"}
                        className="justify-start"
                        onClick={() => setResultWinnerId(resultMatch.team2_id!)}
                      >
                        {teamsById.get(resultMatch.team2_id)?.name ?? resultMatch.team2_id.slice(0, 8)}
                      </Button>
                    )}
                  </>
                )}
              </div>
              <div className="grid gap-1">
                <label className="text-sm text-[#9e9e9e]">Счёт (опционально)</label>
                <Input placeholder="Напр. 2:1" value={resultScore} onChange={(e) => setResultScore(e.target.value)} className="md:max-w-sm" />
              </div>
              <div className="flex gap-3">
                <Button size="sm" disabled={adminSetResultMutation.isPending || !resultWinnerId} onClick={() => void handleResultSubmit()}>
                  {adminSetResultMutation.isPending ? "Сохранение..." : "Установить победителя"}
                </Button>
                <Button size="sm" variant="outline" onClick={() => setResultMatch(null)}>Отмена</Button>
              </div>
            </CardContent>
          </Card>
        )}
        {matchesQuery.isLoading ? (
          <Spinner />
        ) : matchesQuery.isError ? (
          <ErrorState />
        ) : matchesQuery.data?.items.length ? (
          <MatchesTable
            matches={matchesQuery.data.items}
            teams={teamsQuery.data?.items ?? []}
            participants={participantsQuery.data?.items ?? []}
            adminMode
            onSchedule={handleScheduleOpen}
            onSubmitResult={handleResultOpen}
            onApprove={(match) => void handleApproveResult(match.id)}
            onReject={(match) => void handleRejectResult(match.id)}
          />
        ) : (
          <EmptyState title="Матчей нет" description="После генерации сетки матчи появятся здесь." />
        )}
      </SectionCard>}

      {activeSection === "audit" && <SectionCard title="Журнал аудита" description="История действий по турниру.">
        {auditQuery.isLoading ? (
          <Spinner />
        ) : auditQuery.isError ? (
          <ErrorState />
        ) : auditQuery.data?.items.length ? (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Время</TableHead>
                  <TableHead>Действие</TableHead>
                  <TableHead>Пользователь</TableHead>
                  <TableHead>Детали</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {auditQuery.data.items.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>{formatDateTime(item.created_at)}</TableCell>
                    <TableCell>{item.action_type}</TableCell>
                    <TableCell>{item.actor_user_id || "—"}</TableCell>
                    <TableCell className="max-w-[320px] whitespace-pre-wrap break-words text-xs">
                      {item.metadata_json ? JSON.stringify(item.metadata_json, null, 2) : item.description || "—"}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        ) : (
          <EmptyState title="Журнал пуст" description="Действий пока не было." />
        )}
      </SectionCard>}

      </div>
    </div>
  );
}
