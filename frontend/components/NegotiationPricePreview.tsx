import type { Booking } from "@/lib/apitypes";
import { DepositAmountLines } from "@/components/DepositAmountLines";
import { PricingCalculationTip } from "@/components/PricingCalculationTip";
import { ratesFromPayment } from "@/lib/pricingCalculationExample";

type Props = {
  booking: Booking;
  tripDays: number;
  variant: "owner" | "customer";
};

/** Fee breakdown while price is being negotiated (quoted rate excludes platform fee & GST). */
export function NegotiationPricePreview({ booking, tripDays, variant }: Props) {
  const payment = booking.payment;
  const perDay = booking.final_booking_price;
  if (!perDay || !payment) return null;
  const days = payment.trip_days && payment.trip_days > 0 ? payment.trip_days : tripDays;

  return (
    <div className="mt-3 rounded-md border border-slate-200 bg-slate-50 px-3 py-3 text-sm text-slate-700">
      <div className="flex items-start justify-between gap-2">
        <p className="text-xs text-slate-600">
          Quoted rate is <span className="font-medium text-slate-800">exclusive</span> of platform fee and GST.
        </p>
        <PricingCalculationTip variant={variant} rates={ratesFromPayment(payment)} />
      </div>
      <ul className="mt-2 space-y-2">
        <li className="flex justify-between gap-4">
          <span className="text-slate-600">Agreed rental per day</span>
          <span className="shrink-0 tabular-nums font-medium text-slate-900">₹{perDay}</span>
        </li>
        <li className="flex justify-between gap-4">
          <span className="text-slate-600">
            Trip rental{days > 1 ? ` (${days} days)` : ""}
          </span>
          <span className="shrink-0 tabular-nums font-medium text-slate-900">₹{payment.agreed_base_inr}</span>
        </li>
        {variant === "customer" ? (
          <>
            <li className="flex justify-between gap-4">
              <span className="text-slate-600">
                Platform fee ({payment.customer_commission_percent}%)
              </span>
              <span className="shrink-0 tabular-nums text-slate-800">
                + ₹{payment.customer_commission_inr}
              </span>
            </li>
            <li className="flex justify-between gap-4">
              <span className="text-slate-600">
                GST ({payment.gst_percent_on_commission}% on rental + platform fee)
              </span>
              <span className="shrink-0 tabular-nums text-slate-800">+ ₹{payment.customer_gst_inr}</span>
            </li>
            <li className="flex justify-between gap-4 border-t border-slate-200 pt-2 font-semibold text-slate-900">
              <span>Trip total (incl. fees)</span>
              <span className="shrink-0 tabular-nums">₹{payment.customer_total_inr}</span>
            </li>
          </>
        ) : (
          <>
            <li className="flex justify-between gap-4">
              <span className="text-slate-600">
                Platform fee ({payment.owner_commission_percent}%)
              </span>
              <span className="shrink-0 tabular-nums text-slate-800">
                − ₹{payment.owner_commission_inr}
              </span>
            </li>
            <li className="flex justify-between gap-4">
              <span className="text-slate-600">
                GST ({payment.gst_percent_on_commission}% on trip rental)
              </span>
              <span className="shrink-0 tabular-nums text-slate-800">− ₹{payment.owner_gst_inr}</span>
            </li>
            <li className="flex justify-between gap-4 border-t border-slate-200 pt-2 font-semibold text-emerald-950">
              <span>Your earnings (rental)</span>
              <span className="shrink-0 tabular-nums text-emerald-900">₹{payment.owner_net_inr}</span>
            </li>
            <p className="text-xs text-slate-500">Post-trip charges, if any, are added to your payout after the trip.</p>
          </>
        )}
      </ul>
      {variant === "customer" ? (
        <DepositAmountLines payment={payment} variant="customer" className="mt-3" />
      ) : null}
    </div>
  );
}
