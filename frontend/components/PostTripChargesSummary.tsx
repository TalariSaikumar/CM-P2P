import type { BookingPaymentBreakdown } from "@/lib/apitypes";

type Props = {
  payment: BookingPaymentBreakdown;
  variant: "customer" | "owner";
};

/** Read-only list of owner post-trip charges once submitted or when line items exist. */
export function PostTripChargesSummary({ payment, variant }: Props) {
  const items = payment.post_trip_items ?? [];
  const postTripTotal = Number(payment.post_trip_charges_inr ?? 0);
  const hasLines = items.length > 0;
  const billReady =
    payment.payment_phase === "final_due" ||
    payment.payment_phase === "paid" ||
    hasLines ||
    postTripTotal > 0;

  if (!billReady) return null;

  return (
    <div className="rounded-lg border border-amber-200 bg-amber-50/40 px-4 py-3 text-sm text-amber-950 shadow-sm">
      <p className="font-medium text-amber-950">
        {variant === "customer" ? "Post-trip charges (from owner)" : "Post-trip charges you added"}
      </p>
      {hasLines ? (
        <ul className="mt-2 space-y-1.5">
          {items.map((it, i) => (
            <li key={`${it.label}-${i}`} className="flex justify-between gap-3 text-amber-900/90">
              <span className="min-w-0">{it.label}</span>
              <span className="shrink-0 tabular-nums font-medium">₹{it.amount_inr}</span>
            </li>
          ))}
        </ul>
      ) : (
        <p className="mt-1 text-amber-900/80">No additional post-trip charges on this booking.</p>
      )}
      {postTripTotal > 0 && (
        <p className="mt-2 flex justify-between gap-3 border-t border-amber-200/80 pt-2 font-medium text-amber-950">
          <span>Extra charges total</span>
          <span className="tabular-nums">₹{payment.post_trip_charges_inr}</span>
        </p>
      )}
      {variant === "customer" && payment.payment_phase === "final_due" && (
        <ul className="mt-3 space-y-1 border-t border-amber-200/80 pt-2 text-amber-900/90">
          {Number(payment.deposit_paid_inr) > 0 && (
            <li className="flex justify-between gap-3">
              <span>Deposit already paid</span>
              <span className="tabular-nums">− ₹{payment.deposit_paid_inr}</span>
            </li>
          )}
          {Number(payment.trip_balance_inr) > 0 && (
            <li className="flex justify-between gap-3">
              <span>Remaining trip balance</span>
              <span className="tabular-nums">₹{payment.trip_balance_inr}</span>
            </li>
          )}
          <li className="flex justify-between gap-3 text-base font-semibold text-emerald-900">
            <span>Final balance due</span>
            <span className="tabular-nums">₹{payment.final_due_inr}</span>
          </li>
        </ul>
      )}
      {variant === "owner" && payment.payment_phase === "final_due" && (
        <p className="mt-2 text-xs text-amber-900/80">
          Bill sent to the customer. They can pay the remaining balance from their booking page.
        </p>
      )}
      {variant === "owner" && payment.payment_phase === "paid" && (
        <p className="mt-2 text-xs text-amber-900/80">Included in the customer&apos;s completed payment.</p>
      )}
    </div>
  );
}
