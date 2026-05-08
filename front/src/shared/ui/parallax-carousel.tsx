import { useEffect, useRef } from "react";

const ROW_1 = [
  "/images/23Aoxfn80QeSWSV_XPcVNL.jpg",
  "/images/2AsFhvkf0Yo-fTWyENf58C.jpg",
  "/images/C7ZOAhVOtAfgskwCn1MkSy.jpg",
  "/images/H6qy9fnrVmEgMbJfptnNPa.jpg",
  "/images/KZ1ckZTzaodlc1bd-QuXMd.jpg",
  "/images/23Aoxfn80QeSWSV_XPcVNL.jpg",
  "/images/2AsFhvkf0Yo-fTWyENf58C.jpg",
  "/images/C7ZOAhVOtAfgskwCn1MkSy.jpg",
  "/images/H6qy9fnrVmEgMbJfptnNPa.jpg",
  "/images/KZ1ckZTzaodlc1bd-QuXMd.jpg",
  "/images/gettyimages-1340355203-2048x2048.jpg",
];

const ROW_2 = [
  "/images/ma1rKCERws__L8z6nKa3N1.jpg",
  "/images/n3Wx8xNyLfwN5ZXH1WV_VG.jpg",
  "/images/qHTKZ0Zb7pvFEOSYiIZuQa.jpg",
  "/images/UptqzfekwXbnOhGNQ92UUo.jpg",
  "/images/Zhs0xBBR7x1v2HmIL4srKs.jpg",
  "/images/ma1rKCERws__L8z6nKa3N1.jpg",
  "/images/n3Wx8xNyLfwN5ZXH1WV_VG.jpg",
  "/images/qHTKZ0Zb7pvFEOSYiIZuQa.jpg",
  "/images/UptqzfekwXbnOhGNQ92UUo.jpg",
  "/images/Zhs0xBBR7x1v2HmIL4srKs.jpg",
  "/images/gettyimages-1624124840-2048x2048.jpg",
];

export function ParallaxCarousel() {
  const sectionRef = useRef<HTMLElement>(null);
  const row1Ref = useRef<HTMLDivElement>(null);
  const row2Ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let rafId: number;

    const update = () => {
      const section = sectionRef.current;
      const row1 = row1Ref.current;
      const row2 = row2Ref.current;
      if (!section || !row1 || !row2) return;

      const rect = section.getBoundingClientRect();
      const progress = (window.innerHeight - rect.top) / (window.innerHeight + rect.height);
      const offset = (progress - 0.5) * 400;
      const BASE = -500;

      row1.style.transform = `translateX(${BASE - offset}px)`;
      row2.style.transform = `translateX(${BASE + offset}px)`;
    };

    const onScroll = () => {
      cancelAnimationFrame(rafId);
      rafId = requestAnimationFrame(update);
    };

    window.addEventListener("scroll", onScroll, { passive: true });
    update();

    return () => {
      window.removeEventListener("scroll", onScroll);
      cancelAnimationFrame(rafId);
    };
  }, []);

  return (
    <section
      ref={sectionRef}
      style={{
        background: "#001538",
        width: "100vw",
        marginLeft: "calc(50% - 50vw)",
        overflow: "hidden",
        padding: "80px 0 88px",
      }}
    >
      <div style={{ textAlign: "center", marginBottom: 52, padding: "0 16px" }}>
        <p
          style={{
            color: "#90b8ff",
            fontSize: 12,
            fontWeight: 600,
            letterSpacing: "0.35em",
            textTransform: "uppercase",
            marginBottom: 14,
          }}
        >
          Platform
        </p>
        <h2
          style={{
            color: "#ffffff",
            fontSize: "clamp(2.8rem, 7vw, 5.5rem)",
            fontWeight: 800,
            letterSpacing: "-0.03em",
            lineHeight: 1,
            textTransform: "uppercase",
            margin: 0,
          }}
        >
          СТАНЬ ВЕЛИКИМ
        </h2>
        <div
          style={{
            width: 48,
            height: 3,
            background: "#2255ff",
            margin: "20px auto 0",
            borderRadius: 2,
          }}
        />
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
        <div
          ref={row1Ref}
          style={{
            display: "flex",
            gap: 10,
            willChange: "transform",
          }}
        >
          {ROW_1.map((src, i) => (
            <CarouselImage key={i} src={src} />
          ))}
        </div>

        <div
          ref={row2Ref}
          style={{
            display: "flex",
            gap: 10,
            willChange: "transform",
          }}
        >
          {ROW_2.map((src, i) => (
            <CarouselImage key={i} src={src} />
          ))}
        </div>
      </div>
    </section>
  );
}

function CarouselImage({ src }: { src: string }) {
  return (
    <div
      style={{
        flexShrink: 0,
        width: 320,
        aspectRatio: "16 / 9",
        borderRadius: 8,
        overflow: "hidden",
        background: "#002366",
        border: "1px solid #0a3575",
      }}
    >
      <img
        src={src}
        alt=""
        loading="lazy"
        style={{ width: "100%", height: "100%", objectFit: "cover", display: "block" }}
      />
    </div>
  );
}
