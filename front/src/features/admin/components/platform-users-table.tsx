import type { User } from "@/shared/types/api";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";

export function PlatformUsersTable({
  users,
  onBlock,
  onUnblock,
}: {
  users: User[];
  onBlock: (id: string) => void;
  onUnblock: (id: string) => void;
}) {
  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>ID</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Ник</TableHead>
            <TableHead>Роль</TableHead>
            <TableHead>Статус</TableHead>
            <TableHead>Действия</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {users.map((user) => (
            <TableRow key={user.id}>
              <TableCell className="font-medium text-slate-900">{user.id}</TableCell>
              <TableCell>{user.email}</TableCell>
              <TableCell>{user.nickname || "—"}</TableCell>
              <TableCell>{user.role || "player"}</TableCell>
              <TableCell>
                {user.is_blocked ? <Badge tone="danger">Заблокирован</Badge> : <Badge tone="success">Активен</Badge>}
              </TableCell>
              <TableCell>
                {user.is_blocked ? (
                  <Button variant="outline" size="sm" onClick={() => onUnblock(user.id)}>
                    Разблокировать
                  </Button>
                ) : (
                  <Button variant="destructive" size="sm" onClick={() => onBlock(user.id)}>
                    Заблокировать
                  </Button>
                )}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}