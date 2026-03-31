import React from "react";

/** Robô de transição — asset em `public/robo-transic.png` */
export const MESTRE_IA_ROBO_SRC = "/robo-transic.png";

type Props = {
  /** Cobre o container pai (use com `position: relative` no pai) */
  overlay?: boolean;
  minHeight?: number | string;
  caption?: string;
  className?: string;
  size?: "sm" | "md" | "lg";
  /** Imagem e legenda lado a lado (ex.: linha em card) */
  row?: boolean;
};

const sizes = { sm: 112, md: 176, lg: 220 };

/**
 * Loading / transição visual do Mestre IA (robô flutuando).
 */
export function MestreIaTransitionLoader({
  overlay = false,
  minHeight = 220,
  caption,
  className = "",
  size = "md",
  row = false,
}: Props) {
  const px = sizes[size];

  const body = (
    <div
      className={
        row
          ? "flex flex-row align-items-center gap-3"
          : "flex flex-column align-items-center"
      }
    >
      <img
        src={MESTRE_IA_ROBO_SRC}
        alt=""
        width={px}
        height={px}
        className="mestre-ia-transition-robo-img"
        style={{ objectFit: "contain", display: "block", flexShrink: 0 }}
      />
      {caption ? (
        <div
          className={`text-600 text-sm ${row ? "text-left" : "mt-2 text-center"}`}
          style={{ maxWidth: 280 }}
        >
          {caption}
        </div>
      ) : null}
    </div>
  );

  return (
    <>
      {overlay ? (
        <div
          className={`mestre-ia-transition-overlay flex align-items-center justify-content-center ${className}`}
          role="status"
          aria-live="polite"
        >
          <div className="flex flex-column align-items-center justify-content-center px-3">{body}</div>
        </div>
      ) : (
        <div
          className={`flex flex-column align-items-center justify-content-center ${className}`}
          style={minHeight === 0 || minHeight === "0" ? undefined : { minHeight }}
          role="status"
          aria-live="polite"
        >
          {body}
        </div>
      )}

      <style>{`
        .mestre-ia-transition-overlay {
          position: absolute;
          inset: 0;
          z-index: 10;
          background: rgba(255, 255, 255, 0.88);
          backdrop-filter: blur(3px);
          -webkit-backdrop-filter: blur(3px);
          border-radius: inherit;
        }
        .mestre-ia-transition-robo-img {
          animation: mestre-ia-robo-bob 1.55s ease-in-out infinite;
          filter: drop-shadow(0 6px 14px rgba(37, 99, 235, 0.18));
        }
        @keyframes mestre-ia-robo-bob {
          0%, 100% { transform: translateY(0); }
          50% { transform: translateY(-11px); }
        }
      `}</style>
    </>
  );
}
