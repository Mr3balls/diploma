import type { ImportPreviewResponse } from "@/shared/types/api";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";
import { importRowStatusLabel } from "@/shared/lib/enums";

function rowTone(status: string) {
  if (status === "confirmed" || status === "valid") return "success";
  if (status === "duplicate" || status === "rejected") return "danger";
  if (status === "needs_review") return "warning";
  return "muted";
}

export function ImportPreviewTable({
  preview,
  onConfirm,
  isConfirming,
}: {
  preview: ImportPreviewResponse;
  onConfirm: (batchId: string) => void;
  isConfirming?: boolean;
}) {
  return (
    <div className="grid gap-4">
      <div className="flex items-center justify-between gap-4 rounded-2xl border border-[#0a3575] bg-[#002366] p-4">
        <div className="text-sm text-[#90afd4]">
          Batch ID: <span className="font-medium text-white">{preview.batch.id}</span>
        </div>
        <Button onClick={() => onConfirm(preview.batch.id)} disabled={isConfirming}>
          Подтвердить импорт
        </Button>
      </div>

      <div className="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Статус</TableHead>
              <TableHead>Команда</TableHead>
              <TableHead>Капитан</TableHead>
              <TableHead>Игроки</TableHead>
              <TableHead>Ошибки</TableHead>
              <TableHead>Дубликаты</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {preview.rows.map((row) => (
              <TableRow key={row.id}>
                <TableCell>
                  <Badge tone={rowTone(row.status)}>{importRowStatusLabel[row.status]}</Badge>
                </TableCell>
                <TableCell>{row.team_name || "—"}</TableCell>
                <TableCell>{row.captain_nickname || "—"}</TableCell>
                <TableCell>{row.player_nicknames?.join(", ") || "—"}</TableCell>
                <TableCell>
                  <div className="flex flex-col gap-1">
                    {row.validation_errors?.length
                      ? row.validation_errors.map((error) => (
                          <span key={error} className="text-xs text-red-600">
                            {error}
                          </span>
                        ))
                      : "—"}
                  </div>
                </TableCell>
                <TableCell>
                  <div className="flex flex-col gap-1">
                    {row.duplicate_conflicts?.length
                      ? row.duplicate_conflicts.map((conflict) => (
                          <span key={conflict} className="text-xs text-amber-700">
                            {conflict}
                          </span>
                        ))
                      : "—"}
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}