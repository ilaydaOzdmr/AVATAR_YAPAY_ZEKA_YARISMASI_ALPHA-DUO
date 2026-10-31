import { useState } from "react";

interface ChartRow { code: string; predicted: number; actual: number }
const format = (value: number) => value.toLocaleString("tr-TR", { maximumFractionDigits: 1 });

export default function ForecastChart({ rows }: { rows: ChartRow[] }) {
  const [activeCode, setActiveCode] = useState<string | null>(null);
  const active = rows.find((row) => row.code === activeCode);
  const width = Math.max(640, rows.length * 112 + 72);
  const height = 300;
  const left = 52, top = 32, bottom = 258;
  const plotHeight = bottom - top;
  const maximum = Math.max(...rows.flatMap((row) => [row.actual, row.predicted]), 1);
  const magnitude = 10 ** Math.floor(Math.log10(maximum / 4));
  const step = Math.ceil(maximum / 4 / magnitude) * magnitude;
  const limit = step * 4;
  const groupWidth = (width - left - 20) / rows.length;
  const barWidth = 26;

  return (
    <div className="forecast-chart">
      <div className="forecast-chart__legend" aria-label="Grafik açıklaması">
        <span><i className="forecast-chart__key forecast-chart__key--predicted" aria-hidden="true" />Model tahmini</span>
        <span><i className="forecast-chart__key forecast-chart__key--actual" aria-hidden="true" />Gerçekleşen</span>
        <span className="forecast-chart__unit">Miktar (kg)</span>
      </div>
      <div className="forecast-chart__scroll" role="region" aria-label="Üretim miktarı grafiği; dar ekranda yatay kaydırılabilir" tabIndex={0}>
        <svg viewBox={`0 0 ${width} ${height}`} style={{ minWidth: width }} role="group" aria-label="Üretim emirlerine göre model tahmini ve gerçekleşen miktar karşılaştırması">
          <title>Model tahmini ve gerçekleşen üretim miktarı</title>
          <desc>Mavi sütun model tahminini, yeşil sütun gerçekleşen miktarı kilogram olarak gösterir. Her üretim emrine odaklanarak değerleri okuyabilirsiniz.</desc>
          {[0, 1, 2, 3, 4].map((tick) => {
            const y = bottom - tick * plotHeight / 4;
            return (
              <g key={tick} aria-hidden="true">
                <line x1={left} y1={y} x2={width - 20} y2={y} className={tick ? "forecast-chart__grid" : "forecast-chart__axis"} />
                <text x={left - 10} y={y + 4} textAnchor="end" className="forecast-chart__tick">{format(tick * step)}</text>
              </g>
            );
          })}
          {rows.map((row, index) => {
            const center = left + (index + .5) * groupWidth;
            const predictedHeight = row.predicted / limit * plotHeight;
            const actualHeight = row.actual / limit * plotHeight;
            const detail = `${row.code}: model tahmini ${format(row.predicted)} kg, gerçekleşen ${format(row.actual)} kg`;
            return (
              <g key={row.code} className="forecast-chart__group" role="img" aria-label={detail} tabIndex={0}
                onMouseEnter={() => setActiveCode(row.code)} onMouseLeave={() => setActiveCode(null)}
                onFocus={() => setActiveCode(row.code)} onBlur={() => setActiveCode(null)}>
                <title>{detail}</title>
                <rect x={center - groupWidth / 2 + 4} y={12} width={groupWidth - 8} height={height - 16} rx={8}
                  fill={activeCode === row.code ? "#f0f5f3" : "transparent"} />
                <rect x={center - barWidth - 4} y={bottom - predictedHeight} width={barWidth} height={predictedHeight} rx={4} className="forecast-chart__bar--predicted" />
                <rect x={center + 4} y={bottom - actualHeight} width={barWidth} height={actualHeight} rx={4} className="forecast-chart__bar--actual" />
                <text x={center - barWidth / 2 - 4} y={bottom - predictedHeight - 8} textAnchor="middle" className="forecast-chart__value forecast-chart__value--predicted">{format(row.predicted)}</text>
                <text x={center + barWidth / 2 + 4} y={bottom - actualHeight - 8} textAnchor="middle" className="forecast-chart__value forecast-chart__value--actual">{format(row.actual)}</text>
                <text x={center} y={bottom + 24} textAnchor="middle" className="forecast-chart__tick">{row.code.replace(/^URT-/, "")}</text>
              </g>
            );
          })}
        </svg>
      </div>
      <div className="forecast-chart__detail" aria-live="polite" aria-atomic="true">
        {active ? <><b>{active.code}</b><span>Tahmin: {format(active.predicted)} kg</span><span>Gerçekleşen: {format(active.actual)} kg</span><span>Sapma: {active.predicted >= active.actual ? "+" : ""}{format(active.predicted - active.actual)} kg</span></>
          : <span>Değerleri incelemek için sütunların üzerine gelin veya klavyeyle odaklanın.</span>}
      </div>
    </div>
  );
}
