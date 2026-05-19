import { useEffect, useRef } from "react";
import L from "leaflet";
import "leaflet/dist/leaflet.css";

// Fix Leaflet's broken default icon paths when bundled with Vite
import markerIcon2x from "leaflet/dist/images/marker-icon-2x.png";
import markerIcon from "leaflet/dist/images/marker-icon.png";
import markerShadow from "leaflet/dist/images/marker-shadow.png";

delete (L.Icon.Default.prototype as unknown as Record<string, unknown>)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconUrl: markerIcon,
  iconRetinaUrl: markerIcon2x,
  shadowUrl: markerShadow,
});

interface MapPickerProps {
  value: { lat: number; lng: number } | null;
  onChange: (coords: { lat: number; lng: number } | null) => void;
  height?: number;
}

export function MapPicker({ value, onChange, height = 320 }: MapPickerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<L.Map | null>(null);
  const markerRef = useRef<L.Marker | null>(null);

  useEffect(() => {
    if (!containerRef.current || mapRef.current) return;

    const initialCenter: L.LatLngExpression = value
      ? [value.lat, value.lng]
      : [51.505, 82.0]; // default: Kazakhstan center-ish

    const map = L.map(containerRef.current).setView(initialCenter, value ? 13 : 5);
    mapRef.current = map;

    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
    }).addTo(map);

    if (value) {
      markerRef.current = L.marker([value.lat, value.lng]).addTo(map);
    }

    map.on("click", (e: L.LeafletMouseEvent) => {
      const { lat, lng } = e.latlng;
      if (markerRef.current) {
        markerRef.current.setLatLng([lat, lng]);
      } else {
        markerRef.current = L.marker([lat, lng]).addTo(map);
      }
      onChange({ lat, lng });
    });

    return () => {
      map.remove();
      mapRef.current = null;
      markerRef.current = null;
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Sync external value changes (e.g. clear button)
  useEffect(() => {
    if (!mapRef.current) return;
    if (!value) {
      markerRef.current?.remove();
      markerRef.current = null;
    } else if (markerRef.current) {
      markerRef.current.setLatLng([value.lat, value.lng]);
    } else {
      markerRef.current = L.marker([value.lat, value.lng]).addTo(mapRef.current);
    }
  }, [value]);

  return (
    <div className="grid gap-2">
      <div
        ref={containerRef}
        style={{ height, borderRadius: "0.75rem", overflow: "hidden", border: "1px solid #2d2d2d" }}
      />
      {value ? (
        <div className="flex items-center justify-between text-xs text-[#666666]">
          <span>
            {value.lat.toFixed(5)}, {value.lng.toFixed(5)}
          </span>
          <button
            type="button"
            className="text-[#ff5500] hover:underline"
            onClick={() => onChange(null)}
          >
            Убрать метку
          </button>
        </div>
      ) : (
        <p className="text-xs text-[#666666]">Нажмите на карту чтобы поставить метку</p>
      )}
    </div>
  );
}
