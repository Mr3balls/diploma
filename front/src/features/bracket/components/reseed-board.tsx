import {
  DndContext,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  useSortable,
  verticalListSortingStrategy,
  arrayMove,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical } from "lucide-react";
import { Button } from "@/shared/ui/button";
import { Card, CardContent } from "@/shared/ui/card";

type SeedItem = {
  id: string;
  label: string;
};

function SortableRow({ item }: { item: SeedItem }) {
  const { attributes, listeners, setNodeRef, transform, transition } = useSortable({ id: item.id });

  return (
    <div
      ref={setNodeRef}
      style={{ transform: CSS.Transform.toString(transform), transition }}
      className="flex items-center justify-between rounded-xl border border-[#0a3575] bg-[#002366] px-4 py-3"
    >
      <div className="flex items-center gap-3">
        <button type="button" className="cursor-grab text-[#90afd4]" {...attributes} {...listeners}>
          <GripVertical className="h-4 w-4" />
        </button>
        <span className="text-sm font-medium text-white">{item.label}</span>
      </div>
      <span className="text-xs text-[#90afd4]">{item.id}</span>
    </div>
  );
}

export function ReseedBoard({
  items,
  onChange,
  onSave,
  disabled,
  saving,
}: {
  items: SeedItem[];
  onChange: (next: SeedItem[]) => void;
  onSave: () => void;
  disabled?: boolean;
  saving?: boolean;
}) {
  const sensors = useSensors(useSensor(PointerSensor));

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (!over || active.id === over.id) return;

    const oldIndex = items.findIndex((item) => item.id === active.id);
    const newIndex = items.findIndex((item) => item.id === over.id);
    onChange(arrayMove(items, oldIndex, newIndex));
  }

  return (
    <Card>
      <CardContent className="grid gap-4 pt-5">
        <p className="text-sm text-[#90afd4]">
          Перестановка доступна только до перехода турнира в статус <b>in_progress</b>.
        </p>
        <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
          <SortableContext items={items.map((item) => item.id)} strategy={verticalListSortingStrategy}>
            <div className="grid gap-3">
              {items.map((item) => (
                <SortableRow key={item.id} item={item} />
              ))}
            </div>
          </SortableContext>
        </DndContext>
        <div>
          <Button onClick={onSave} disabled={disabled || saving || !items.length}>
            Сохранить порядок посева
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}