/** YYYY-MM-DD in the user's local timezone (for `<input type="date">`). */
export function localDateInputValue(d: Date = new Date()): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

/** Short label for trip dates, e.g. "08 Jan 26 (Mon)". */
export function formatTripDateShort(iso: string): string {
  const d = new Date(iso);
  const day = String(d.getUTCDate()).padStart(2, "0");
  const month = d.toLocaleString("en-US", { month: "short", timeZone: "UTC" });
  const yy = String(d.getUTCFullYear()).slice(-2);
  const weekday = d.toLocaleString("en-US", { weekday: "short", timeZone: "UTC" });
  return `${day} ${month} ${yy} (${weekday})`;
}

/** Range label for booking header, e.g. "08 Jan 26 (Mon) to 11 Jan 26 (Thu)". */
export function tripDateRangeLabel(rentalFrom: string, rentalTo: string): string {
  return `${formatTripDateShort(rentalFrom)} to ${formatTripDateShort(rentalTo)}`;
}

/** Inclusive calendar days between rental_from and rental_to (UTC calendar dates only). */
export function tripDaysInclusive(rentalFrom: string, rentalTo: string): number {
  const a = new Date(rentalFrom);
  const b = new Date(rentalTo);
  const start = Date.UTC(a.getUTCFullYear(), a.getUTCMonth(), a.getUTCDate());
  const end = Date.UTC(b.getUTCFullYear(), b.getUTCMonth(), b.getUTCDate());
  if (end < start) return 1;
  // Use floor on whole UTC days — round() wrongly adds a day when rental_to is 23:59:59.
  const wholeDays = Math.floor((end - start) / 86_400_000);
  return Math.max(1, wholeDays + 1);
}

/** After changing the start date, keep end on or after start (and not before today). */
export function endDateAfterStartChange(start: string, end: string, today = localDateInputValue()): string {
  if (!start) return end;
  const floor = start < today ? today : start;
  if (!end || end < floor) return floor;
  return end;
}
