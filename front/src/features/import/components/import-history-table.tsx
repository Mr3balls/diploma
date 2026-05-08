import type { ImportBatch } from "@/shared/types/api";
import { Button } from "@/shared/ui/button";
import { Badge } from "@/shared/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";
import { formatDateTime } from "@/shared/lib/date";
import { importBatchStatusLabel } from "@/shared/lib/enums";

function tone(status: ImportBatch["status"]) {
  if (status === "confirmed") return "success";
  if (status === "failed") return "danger";
  if (status === "preview_ready") return "warning";
  return "muted";
}

export function ImportHistoryTable({
  items,
  onOpen,
}: {
  items: ImportBatch[];
  onOpen: (batchId: string) => void;
}) {
  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>ID</TableHead>
            <TableHead>Статус</TableHead>
            <TableHead>Лист</TableHead>
            <TableHead>Строк</TableHead>
            <TableHead>Создан</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((item) => (
            <TableRow key={item.id}>
              <TableCell className="font-medium text-white">{item.id}</TableCell>
              <TableCell>
                <Badge tone={tone(item.status)}>{importBatchStatusLabel[item.status]}</Badge>
              </TableCell>
              <TableCell>{item.worksheet_name || "—"}</TableCell>
              <TableCell>{item.row_count ?? "—"}</TableCell>
              <TableCell>{formatDateTime(item.created_at)}</TableCell>
              <TableCell>
                <Button variant="outline" size="sm" onClick={() => onOpen(item.id)}>
                  Открыть
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}