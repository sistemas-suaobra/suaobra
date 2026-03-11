export function percent(numerator: number, denominator: number) {
  if (!denominator || denominator <= 0) return 0;
  return Math.round((numerator / denominator) * 100);
}