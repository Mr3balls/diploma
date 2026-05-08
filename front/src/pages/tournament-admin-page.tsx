import { useEffect, useState } from "react";
import { useParams, Navigate } from "react-router-dom";
import { toast } from "sonner";
import { useAuth } from "@/app/providers/auth-provider";
import { useTournamentAdminAccess } from "@/shared/hooks/use-tournament-admin-access";
import {
  useAddManager,
  useChangeTournamentStatus,
  useRemoveManager,
  useTournament,
  useTournamentAudit,
  useUpdateTournament,
} from "@/features/tournaments/hooks";
import { useTournamentImports, useConfirmImport, useImportBatch, useConnectGoogleSheet, usePreviewImport, useValidateGoogleSheet } from "@/features/import/hooks";
import { useApproveTeam, useRejectTeam, useTeam, useTournamentAdminTeams, useRemoveMember } from "@/features/teams/hooks";
import { useGenerateBracket, useRegenerateBracket, useReseedBracket, useTournamentBracket } from "@/features/bracket/hooks";
import {
  useApproveResult,
  useConfirmReady,
  useRejectResult,
  useReportIssue,
  useRequestReschedule,
  useScheduleMatch,
  useSubmitResult,
  useTournamentAdminMatches,
} from "@/features/matches/hooks";
import { CreateTournamentForm } from "@/features/tournaments/components/create-tournament-form";
import { GoogleSheetForm } from "@/features/import/components/google-sheet-form";
import { ImportHistoryTable } from "@/features/import/components/import-history-table";
import { ImportPreviewTable } from "@/features/import/components/import-preview-table";
import { TeamDetailsCard } from "@/features/teams/components/team-details-card";
import { TeamsTable } from "@/features/teams/components/teams-table";
import { BracketView } from "@/features/bracket/components/bracket-view";
import { ReseedBoard } from "@/features/bracket/components/reseed-board";
import { MatchesTable } from "@/features/matches/components/matches-table";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { managerSchema, tournamentStatusSchema, type ManagerFormValues, type TournamentFormValues } from "@/features/tournaments/schemas";
import { requestReasonSchema, scheduleMatchSchema, submitResultSchema, type RequestReasonValues, type ScheduleMatchValues, type SubmitResultValues } from "@/features/matches/schemas";
import { FormField } from "@/shared/ui/form-field";
import { Input } from "@/shared/ui/input";
import { Select } from "@/shared/ui/select";
import { Textarea } from "@/shared/ui/textarea";
import { Button } from "@/shared/ui/button";
import { PageHeader } from "@/shared/ui/page-header";
import { SectionCard } from "@/shared/ui/section";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
import { Spinner } from "@/shared/ui/spinner";
import { formatDateTime } from "@/shared/lib/date";
import { deriveSeedOrderFromMatches, deriveSeedOrderFromTeams } from "@/shared/lib/bracket";
import { getErrorMessage } from "@/shared/lib/http";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";

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
    defaultValues: {
      user_id: "",
    },
  });

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = form;

  return (
    <form className="grid gap-4 md:grid-cols-[1fr_auto_auto]" onSubmit={handleSubmit(onAdd)}>
      <FormField label="user_id менеджера" error={errors.user_id?.message}>
        <Input {...register("user_id")} placeholder="UUID пользователя" />
      </FormField>
      <div className="pt-7">
        <Button type="submit" disabled={isBusy}>
          Добавить
        </Button>
      </div>
      <div className="pt-7">
        <Button type="button" variant="destructive" disabled={isBusy} onClick={handleSubmit(onRemove)}>
          Удалить
        </Button>
      </div>
    </form>
  );
}

function TournamentStatusForm({
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

  return (
    <form
      className="flex flex-wrap items-end gap-4"
      onSubmit={form.handleSubmit((values) => onSubmit(values.status))}
    >
      <FormField label="Статус">
        <Select {...form.register("status")}>
          <option value="draft">draft</option>
          <option value="registration_open">registration_open</option>
          <option value="registration_closed">registration_closed</option>
          <option value="bracket_generated">bracket_generated</option>
          <option value="in_progress">in_progress</option>
          <option value="finished">finished</option>
          <option value="cancelled">cancelled</option>
        </Select>
      </FormField>
      <Button type="submit" disabled={isBusy}>
        Обновить статус
      </Button>
    </form>
  );
}

function MatchScheduleInline({
  onSubmit,
  isBusy,
}: {
  onSubmit: (values: ScheduleMatchValues) => void;
  isBusy?: boolean;
}) {
  const form = useForm<ScheduleMatchValues>({
    resolver: zodResolver(scheduleMatchSchema),
    defaultValues: {
      scheduled_at: "",
    },
  });

  return (
    <form className="flex flex-wrap items-end gap-3" onSubmit={form.handleSubmit(onSubmit)}>
      <FormField label="scheduled_at">
        <Input type="datetime-local" {...form.register("scheduled_at")} />
      </FormField>
      <Button type="submit" disabled={isBusy}>
        Сохранить
      </Button>
    </form>
  );
}

function MatchReasonInline({
  label,
  onSubmit,
  isBusy,
}: {
  label: string;
  onSubmit: (values: RequestReasonValues) => void;
  isBusy?: boolean;
}) {
  const form = useForm<RequestReasonValues>({
    resolver: zodResolver(requestReasonSchema),
    defaultValues: {
      reason: "",
    },
  });

  return (
    <form className="grid gap-3" onSubmit={form.handleSubmit(onSubmit)}>
      <FormField label={label}>
        <Textarea {...form.register("reason")} />
      </FormField>
      <Button type="submit" disabled={isBusy}>
        Отправить
      </Button>
    </form>
  );
}

function MatchResultInline({
  onSubmit,
  isBusy,
}: {
  onSubmit: (values: SubmitResultValues) => void;
  isBusy?: boolean;
}) {
  const form = useForm<SubmitResultValues>({
    resolver: zodResolver(submitResultSchema),
    defaultValues: {
      winner_team_id: "",
      score_text: "",
    },
  });

  return (
    <form className="grid gap-3 md:grid-cols-[1fr_1fr_auto]" onSubmit={form.handleSubmit(onSubmit)}>
      <FormField label="winner_team_id">
        <Input {...form.register("winner_team_id")} />
      </FormField>
      <FormField label="score_text">
        <Input {...form.register("score_text")} placeholder="2:0" />
      </FormField>
      <div className="pt-7">
        <Button type="submit" disabled={isBusy}>
          Отправить
        </Button>
      </div>
    </form>
  );
}

export function TournamentAdminPage() {
  const { id = "" } = useParams();
  const { user } = useAuth();

  const tournamentQuery = useTournament(id);
  const access = useTournamentAdminAccess(id, tournamentQuery.data);

  const importsQuery = useTournamentImports(id, access.canAccessAdmin);
  const teamsQuery = useTournamentAdminTeams(id, access.canAccessAdmin);
  const bracketQuery = useTournamentBracket(id);
  const matchesQuery = useTournamentAdminMatches(id, access.canAccessAdmin);
  const auditQuery = useTournamentAudit(id, access.canAccessAdmin);

  const updateTournamentMutation = useUpdateTournament(id);
  const changeStatusMutation = useChangeTournamentStatus(id);
  const addManagerMutation = useAddManager(id);
  const removeManagerMutation = useRemoveManager(id);

  const connectSheetMutation = useConnectGoogleSheet(id);
  const validateSheetMutation = useValidateGoogleSheet(id);
  const previewImportMutation = usePreviewImport(id);
  const confirmImportMutation = useConfirmImport(id);

  const approveTeamMutation = useApproveTeam(id);
  const rejectTeamMutation = useRejectTeam(id);
  const removeMemberMutation = useRemoveMember(id);

  const generateBracketMutation = useGenerateBracket(id);
  const regenerateBracketMutation = useRegenerateBracket(id);
  const reseedBracketMutation = useReseedBracket(id);

  const scheduleMatchMutation = useScheduleMatch(id);
  const confirmReadyMutation = useConfirmReady(id);
  const requestRescheduleMutation = useRequestReschedule(id);
  const reportIssueMutation = useReportIssue(id);
  const submitResultMutation = useSubmitResult(id);
  const approveResultMutation = useApproveResult(id);
  const rejectResultMutation = useRejectResult(id);

  const [selectedTeamId, setSelectedTeamId] = useState<string | null>(null);
  const selectedTeamQuery = useTeam(selectedTeamId ?? undefined, Boolean(selectedTeamId) && access.canAccessAdmin);

  const [selectedBatchId, setSelectedBatchId] = useState<string | null>(null);
  const selectedBatchQuery = useImportBatch(selectedBatchId ?? undefined, Boolean(selectedBatchId) && access.canAccessAdmin);

  const [sheetDefaults, setSheetDefaults] = useState<{ sheet_url?: string; worksheet_name?: string }>({});
  const [reseedItems, setReseedItems] = useState<{ id: string; label: string }[]>([]);
  const [activeMatchId, setActiveMatchId] = useState<string | null>(null);
  const [matchAction, setMatchAction] = useState<"schedule" | "reschedule" | "issue" | "result" | null>(null);

  useEffect(() => {
    const fromMatches =
      bracketQuery.data?.matches?.length
        ? deriveSeedOrderFromMatches(bracketQuery.data.matches).map((id) => ({
            id,
            label: teamsQuery.data?.items.find((team) => team.id === id)?.name || id,
          }))
        : [];
    const fallback =
      teamsQuery.data?.items?.length
        ? deriveSeedOrderFromTeams(teamsQuery.data.items).map((teamId) => ({
            id: teamId,
            label: teamsQuery.data?.items.find((team) => team.id === teamId)?.name || teamId,
          }))
        : [];

    const next = fromMatches.length ? fromMatches : fallback;
    setReseedItems(next);
  }, [bracketQuery.data?.matches, teamsQuery.data?.items]);

  if (tournamentQuery.isLoading || access.isLoading) return <Spinner />;
  if (tournamentQuery.isError || !tournamentQuery.data) return <ErrorState />;
  if (!user || !access.canAccessAdmin) {
    return <Navigate to={`/tournaments/${id}`} replace />;
  }

  const tournament = tournamentQuery.data;
  const canReseed = tournament.status !== "in_progress" && tournament.status !== "finished" && tournament.status !== "cancelled";

  async function handleUpdateTournament(values: TournamentFormValues) {
    try {
      await updateTournamentMutation.mutateAsync(values);
      toast.success("Основные настройки обновлены");
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
      toast.success("Менеджер добавлен");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleRemoveManager(values: ManagerFormValues) {
    try {
      await removeManagerMutation.mutateAsync(values.user_id);
      toast.success("Менеджер удалён");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleConnectSheet(values: { sheet_url: string; worksheet_name: string }) {
    try {
      await connectSheetMutation.mutateAsync(values);
      setSheetDefaults(values);
      toast.success("Связь с таблицей сохранена");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleValidateSheet(values: { sheet_url: string; worksheet_name: string }) {
    try {
      const response = await validateSheetMutation.mutateAsync(values);
      setSheetDefaults(values);
      toast.success(`Проверка прошла: строк ${response.row_count}`);
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handlePreviewImport(values: { sheet_url: string; worksheet_name: string }) {
    try {
      const response = await previewImportMutation.mutateAsync(values);
      setSelectedBatchId(response.batch.id);
      setSheetDefaults(values);
      toast.success("Превью импорта готово");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleConfirmImport(batchId: string) {
    try {
      await confirmImportMutation.mutateAsync({ batch_id: batchId });
      toast.success("Импорт подтверждён");
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
    try {
      await rejectTeamMutation.mutateAsync(teamId);
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

  async function handleSaveReseed() {
    try {
      await reseedBracketMutation.mutateAsync({ ordered_team_ids: reseedItems.map((item) => item.id) });
      toast.success("Порядок посева сохранён");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleConfirmReady(matchId: string) {
    try {
      await confirmReadyMutation.mutateAsync(matchId);
      toast.success("Готовность подтверждена");
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

  const activeMatch = matchesQuery.data?.items.find((item) => item.id === activeMatchId) ?? null;

  return (
    <div className="grid gap-6">
      <PageHeader
        title={`Admin · ${tournament.title}`}
        description="Отдельная административная страница турнира. Публичный просмотр и управление разделены."
      />

      <SectionCard title="General settings" description="Редактирование турнира и смена статуса.">
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
            submitLabel="Сохранить изменения"
            onSubmit={handleUpdateTournament}
            isSubmitting={updateTournamentMutation.isPending}
          />
          <TournamentStatusForm
            currentStatus={tournament.status}
            onSubmit={handleStatusUpdate}
            isBusy={changeStatusMutation.isPending}
          />
        </div>
      </SectionCard>

      <SectionCard
        title="Managers"
        description="Backend не даёт списка менеджеров и не поддерживает поиск пользователей. Только manual user_id input."
      >
        <ManagerForm
          onAdd={handleAddManager}
          onRemove={handleRemoveManager}
          isBusy={addManagerMutation.isPending || removeManagerMutation.isPending}
        />
      </SectionCard>

      <SectionCard
        title="Public Google Sheet connect"
        description="Frontend отправляет только sheet_url и worksheet_name. OAuth и Google account connect отсутствуют намеренно."
      >
        <div className="grid gap-4">
          <GoogleSheetForm
            defaultValues={sheetDefaults}
            onConnect={handleConnectSheet}
            onValidate={handleValidateSheet}
            onPreview={handlePreviewImport}
            isBusy={
              connectSheetMutation.isPending ||
              validateSheetMutation.isPending ||
              previewImportMutation.isPending
            }
          />
          {validateSheetMutation.data ? (
            <div className="rounded-2xl border border-[#0a3575] bg-[#002366] p-4 text-sm text-[#90afd4]">
              <div>Spreadsheet ID: {validateSheetMutation.data.spreadsheet_id}</div>
              <div>Worksheet: {validateSheetMutation.data.worksheet_name}</div>
              <div>Rows: {validateSheetMutation.data.row_count}</div>
              <div>Sample row: {validateSheetMutation.data.sample_row.join(" | ")}</div>
            </div>
          ) : null}
        </div>
      </SectionCard>

      <SectionCard title="Import preview" description="Ошибки валидации и конфликты дубликатов показываются явно.">
        {previewImportMutation.data ? (
          <ImportPreviewTable
            preview={previewImportMutation.data}
            onConfirm={handleConfirmImport}
            isConfirming={confirmImportMutation.isPending}
          />
        ) : selectedBatchQuery.data ? (
          <ImportPreviewTable
            preview={selectedBatchQuery.data}
            onConfirm={handleConfirmImport}
            isConfirming={confirmImportMutation.isPending}
          />
        ) : (
          <EmptyState
            title="Нет превью импорта"
            description="Сначала выполните preview для выбранной Google Sheet."
          />
        )}
      </SectionCard>

      <SectionCard title="Import history" description="История import batches для текущего турнира.">
        {importsQuery.isLoading ? (
          <Spinner />
        ) : importsQuery.isError ? (
          <ErrorState />
        ) : importsQuery.data?.items.length ? (
          <ImportHistoryTable items={importsQuery.data.items} onOpen={setSelectedBatchId} />
        ) : (
          <EmptyState title="История пуста" description="Импортов пока не было." />
        )}
      </SectionCard>

      <SectionCard title="Teams and confirmations" description="Одобрение команды доступно после выполнения правил готовности.">
        <div className="grid gap-4">
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
            />
          ) : (
            <EmptyState title="Нет команд" description="После импорта команды появятся здесь." />
          )}

          {selectedTeamId && selectedTeamQuery.isLoading ? <Spinner /> : null}
          {selectedTeamQuery.data ? (
            <TeamDetailsCard
              data={selectedTeamQuery.data}
              allowAdminActions
              onRemoveMember={handleRemoveMember}
            />
          ) : null}
        </div>
      </SectionCard>

      <SectionCard
        title="Bracket generation"
        description="Single elimination. Базовый посев случайный. Ресидинг отдельным блоком ниже."
        actions={
          <>
            <Button onClick={handleGenerateBracket} disabled={generateBracketMutation.isPending}>
              Сгенерировать
            </Button>
            <Button variant="outline" onClick={handleRegenerateBracket} disabled={regenerateBracketMutation.isPending}>
              Пересоздать
            </Button>
          </>
        }
      >
        {bracketQuery.isLoading ? <Spinner /> : bracketQuery.isError ? <ErrorState /> : <BracketView matches={bracketQuery.data?.matches ?? []} adminMode />}
      </SectionCard>

      <SectionCard title="Drag-and-drop reseeding" description='POST /tournaments/:id/bracket/reseed с payload { "ordered_team_ids": string[] }'>
        <ReseedBoard
          items={reseedItems}
          onChange={setReseedItems}
          onSave={handleSaveReseed}
          disabled={!canReseed}
          saving={reseedBracketMutation.isPending}
        />
      </SectionCard>

      <SectionCard title="Match management" description="Расписание, готовность, переносы, проблемы и результаты.">
        <div className="grid gap-4">
          {matchesQuery.isLoading ? (
            <Spinner />
          ) : matchesQuery.isError ? (
            <ErrorState />
          ) : matchesQuery.data?.items.length ? (
            <MatchesTable
              matches={matchesQuery.data.items}
              adminMode
              onSchedule={(match) => {
                setActiveMatchId(match.id);
                setMatchAction("schedule");
              }}
              onConfirmReady={(match) => void handleConfirmReady(match.id)}
              onReschedule={(match) => {
                setActiveMatchId(match.id);
                setMatchAction("reschedule");
              }}
              onIssue={(match) => {
                setActiveMatchId(match.id);
                setMatchAction("issue");
              }}
              onSubmitResult={(match) => {
                setActiveMatchId(match.id);
                setMatchAction("result");
              }}
              onApprove={(match) => void handleApproveResult(match.id)}
              onReject={(match) => void handleRejectResult(match.id)}
            />
          ) : (
            <EmptyState title="Матчей нет" description="После генерации сетки матчи появятся здесь." />
          )}

          {activeMatch ? (
            <div className="rounded-2xl border border-[#0a3575] bg-[#002366] p-4">
              <div className="mb-4 text-sm font-medium text-white">
                Активный матч: {activeMatch.id}
              </div>
              {matchAction === "schedule" ? (
                <MatchScheduleInline
                  onSubmit={async (values) => {
                    try {
                      await scheduleMatchMutation.mutateAsync({ matchId: activeMatch.id, payload: values });
                      toast.success("Время матча обновлено");
                    } catch (error) {
                      toast.error(getErrorMessage(error));
                    }
                  }}
                  isBusy={scheduleMatchMutation.isPending}
                />
              ) : null}
              {matchAction === "reschedule" ? (
                <MatchReasonInline
                  label="Причина переноса"
                  onSubmit={async (values) => {
                    try {
                      await requestRescheduleMutation.mutateAsync({ matchId: activeMatch.id, payload: values });
                      toast.success("Запрос на перенос отправлен");
                    } catch (error) {
                      toast.error(getErrorMessage(error));
                    }
                  }}
                  isBusy={requestRescheduleMutation.isPending}
                />
              ) : null}
              {matchAction === "issue" ? (
                <MatchReasonInline
                  label="Описание проблемы"
                  onSubmit={async (values) => {
                    try {
                      await reportIssueMutation.mutateAsync({ matchId: activeMatch.id, payload: values });
                      toast.success("Проблема отправлена");
                    } catch (error) {
                      toast.error(getErrorMessage(error));
                    }
                  }}
                  isBusy={reportIssueMutation.isPending}
                />
              ) : null}
              {matchAction === "result" ? (
                <MatchResultInline
                  onSubmit={async (values) => {
                    try {
                      await submitResultMutation.mutateAsync({ matchId: activeMatch.id, payload: values });
                      toast.success("Результат отправлен");
                    } catch (error) {
                      toast.error(getErrorMessage(error));
                    }
                  }}
                  isBusy={submitResultMutation.isPending}
                />
              ) : null}
            </div>
          ) : null}
        </div>
      </SectionCard>

      <SectionCard title="Audit log" description="Журнал действий по турниру.">
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
                  <TableHead>Actor</TableHead>
                  <TableHead>Details</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {auditQuery.data.items.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>{formatDateTime(item.created_at)}</TableCell>
                    <TableCell>{item.action}</TableCell>
                    <TableCell>{item.actor_email || item.actor_user_id || "—"}</TableCell>
                    <TableCell className="max-w-[360px] whitespace-pre-wrap break-words text-xs">
                      {item.details ? JSON.stringify(item.details, null, 2) : "—"}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        ) : (
          <EmptyState title="Записей аудита нет" description="Audit log пока пуст." />
        )}
      </SectionCard>
    </div>
  );
}