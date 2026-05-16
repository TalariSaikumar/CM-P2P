/** True while the trip is live or finished (not cancelled / pre-confirm). */
export function isTripActive(status: string): boolean {
  return status === "CONFIRMED" || status === "COMPLETED";
}

/** True while owner and customer are still agreeing price / trip details. */
export function isNegotiating(status: string): boolean {
  return status === "PENDING" || status === "NEGOTIATING";
}

export function isTripCompleted(status: string): boolean {
  return status === "COMPLETED";
}
