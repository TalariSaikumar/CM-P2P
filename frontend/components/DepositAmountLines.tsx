import type { BookingPaymentBreakdown } from "@/lib/apitypes";

type Props = {
  payment: BookingPaymentBreakdown;
  variant: "owner" | "customer";
  className?: string;
};

/** Deposit due, paid, and remaining trip balance (75% deposit flow). */
export function DepositAmountLines({ payment, variant, className = "" }: Props) {
  const pct = payment.deposit_percent ?? 75;
  const status = payment.payment_status;
  const due = payment.deposit_due_inr;
  const paid = payment.deposit_paid_inr;
  const balance = payment.trip_balance_inr;
  const hasPaid = paid != null && Number(paid) > 0;

  if (status === "UNPAID" && due) {
    return (
      <div className={`space-y-2 border-t border-slate-200 pt-2 text-sm ${className}`}>
        <div className="flex justify-between gap-4">
          <span className="text-slate-600">
            {variant === "customer" ? `Deposit due (${pct}%)` : `Customer deposit due (${pct}%)`}
          </span>
          <span className="shrink-0 font-semibold tabular-nums text-slate-900">₹{due}</span>
        </div>
        <p className="text-xs text-slate-500">
          {variant === "customer"
            ? "Pay this after the owner confirms the booking to lock the trip."
            : "The customer pays this upfront after you confirm the booking."}
        </p>
      </div>
    );
  }

  if (!hasPaid) return null;

  return (
    <div className={`space-y-2 border-t border-slate-200 pt-2 text-sm ${className}`}>
      <div className="flex justify-between gap-4">
        <span className="text-slate-600">
          {variant === "customer" ? `Deposit paid (${pct}%)` : `Customer deposit received (${pct}%)`}
        </span>
        <span className="shrink-0 font-medium tabular-nums text-slate-900">₹{paid}</span>
      </div>
      {status !== "PAID" && balance && Number(balance) > 0 ? (
        <div className="flex justify-between gap-4">
          <span className="text-slate-600">
            {variant === "customer" ? "Balance after deposit" : "Customer balance after deposit"}
          </span>
          <span className="shrink-0 tabular-nums text-slate-800">₹{balance}</span>
        </div>
      ) : null}
      {status === "FINAL_DUE" && payment.final_due_inr && Number(payment.final_due_inr) > 0 ? (
        <div className="flex justify-between gap-4 font-medium text-slate-900">
          <span>{variant === "customer" ? "Final balance due" : "Customer final balance due"}</span>
          <span className="shrink-0 tabular-nums">₹{payment.final_due_inr}</span>
        </div>
      ) : null}
    </div>
  );
}
