"use client";

import { useEffect, useState } from "react";

const generateData = () => {
  const data = [];
  for (let i = 0; i < 24; i++) {
    data.push({
      time: `${i.toString().padStart(2, "0")}:00`,
      inbound: Math.floor(Math.random() * 80) + 20,
      outbound: Math.floor(Math.random() * 60) + 10,
    });
  }
  return data;
};

export function NetworkChart() {
  const [data, setData] = useState(generateData());
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  useEffect(() => {
    const interval = setInterval(() => {
      setData((prev) => {
        const newData = [...prev.slice(1)];
        const lastTime = parseInt(prev[prev.length - 1].time.split(":")[0]);
        newData.push({
          time: `${((lastTime + 1) % 24).toString().padStart(2, "0")}:00`,
          inbound: Math.floor(Math.random() * 80) + 20,
          outbound: Math.floor(Math.random() * 60) + 10,
        });
        return newData;
      });
    }, 3000);

    return () => clearInterval(interval);
  }, []);

  const maxValue = Math.max(...data.flatMap((d) => [d.inbound, d.outbound]));

  return (
    <div className="relative h-64">
      {/* Y-axis labels */}
      <div className="absolute left-0 top-0 bottom-8 w-12 flex flex-col justify-between text-xs text-muted-foreground">
        <span>{maxValue} Mbps</span>
        <span>{Math.floor(maxValue / 2)} Mbps</span>
        <span>0 Mbps</span>
      </div>

      {/* Chart area */}
      <div className="ml-14 h-full flex items-end gap-1 pb-8">
        {data.map((d, i) => (
          <div
            key={i}
            className="flex-1 flex flex-col items-center gap-1 group cursor-pointer"
            onMouseEnter={() => setHoveredIndex(i)}
            onMouseLeave={() => setHoveredIndex(null)}
          >
            {/* Tooltip */}
            {hoveredIndex === i && (
              <div className="absolute bottom-full mb-2 glass-card px-3 py-2 rounded-lg text-xs z-10">
                <p className="font-medium">{d.time}</p>
                <p className="text-primary">In: {d.inbound} Mbps</p>
                <p className="text-emerald-400">Out: {d.outbound} Mbps</p>
              </div>
            )}

            {/* Bars */}
            <div className="w-full flex gap-0.5 items-end h-48">
              <div
                className="flex-1 rounded-t-sm bg-primary/80 transition-all duration-300 hover:bg-primary"
                style={{ height: `${(d.inbound / maxValue) * 100}%` }}
              />
              <div
                className="flex-1 rounded-t-sm bg-emerald-400/60 transition-all duration-300 hover:bg-emerald-400"
                style={{ height: `${(d.outbound / maxValue) * 100}%` }}
              />
            </div>

            {/* X-axis label */}
            {i % 4 === 0 && (
              <span className="text-xs text-muted-foreground mt-2">{d.time}</span>
            )}
          </div>
        ))}
      </div>

      {/* Legend */}
      <div className="absolute top-0 right-0 flex items-center gap-4 text-xs">
        <div className="flex items-center gap-2">
          <div className="h-3 w-3 rounded-sm bg-primary" />
          <span className="text-muted-foreground">Inbound</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="h-3 w-3 rounded-sm bg-emerald-400" />
          <span className="text-muted-foreground">Outbound</span>
        </div>
      </div>
    </div>
  );
}
