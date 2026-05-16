import type { Booking } from "@/lib/apitypes";
import { DepositAmountLines } from "@/components/DepositAmountLines";
import { PricingCalculationTip } from "@/components/PricingCalculationTip";
import { ratesFromPayment } from "@/lib/pricingCalculationExample";
import { isNegotiating } from "@/lib/bookingStatus";
import { formatTripDateShort } from "@/lib/rentalDates";

type Props = {
  booking: Booking;
  tripDays: number;
  customerTripTotalInr: string | null;
  postTripInr: number;
  /** Who is viewing the booking page (pricing column differs by role). */
  viewerRole: "owner" | "customer" | null;
};

function ownerEarningsInr(
  payment: NonNullable<Booking["payment"]>,
  postTripInr: number
): string {
  if (payment.payment_status === "PAID") {
    const net = Number(payment.owner_net_inr);
    const post = Number.isFinite(postTripInr) ? postTripInr : 0;
    return (net + post).toFixed(2);
  }
  return payment.owner_projected_payout_inr ?? payment.owner_net_inr;
}

function statusBadgeClass(status: string): string {
  switch (status) {
    case "CONFIRMED":
      return "bg-emerald-100 text-emerald-800 ring-emerald-600/20";
    case "COMPLETED":
      return "bg-slate-200 text-slate-800 ring-slate-600/20";
    case "CANCELLED":
      return "bg-red-100 text-red-800 ring-red-600/20";
    case "NEGOTIATING":
      return "bg-amber-100 text-amber-900 ring-amber-600/20";
    default:
      return "bg-slate-100 text-slate-700 ring-slate-600/20";
  }
}

export function BookingSummaryCard({
  booking,
  tripDays,
  customerTripTotalInr,
  postTripInr,
  viewerRole,
}: Props) {
  const finalPrice = customerTripTotalInr ?? booking.final_booking_price;
  const fromLabel = formatTripDateShort(booking.rental_from);
  const toLabel = formatTripDateShort(booking.rental_to);
  const payment = booking.payment;
  const ownerEarnings =
    viewerRole === "owner" && payment ? ownerEarningsInr(payment, postTripInr) : null;
  const negotiating = isNegotiating(booking.status);

  return (
    <div className="mt-3 overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm ring-1 ring-slate-100">
      <div className="flex flex-wrap items-start justify-between gap-2 border-b border-slate-100 bg-slate-50/80 px-4 py-3">
        <div className="min-w-0">
          <p className="font-medium text-slate-900">
            {booking.car.car_name} · {booking.car.car_model}
          </p>
          <p className="text-sm text-slate-600">{booking.car.car_number}</p>
          <p className="mt-2 flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-slate-600">
            <span>
              <span className="font-medium text-slate-700">Pickup · </span>
              {booking.pickup_point.trim() ? booking.pickup_point : <span className="text-slate-400">Not set</span>}
            </span>
            <span className="hidden text-slate-300 sm:inline" aria-hidden>
              →
            </span>
            <span>
              <span className="font-medium text-slate-700">Return · </span>
              {booking.drop_point.trim() ? booking.drop_point : <span className="text-slate-400">Not set</span>}
            </span>
          </p>
        </div>
        <span
          className={`inline-flex shrink-0 rounded-full px-2.5 py-1 text-xs font-semibold ring-1 ring-inset ${statusBadgeClass(booking.status)}`}
        >
          {booking.status}
        </span>
      </div>

      <dl className="grid sm:grid-cols-2">
        <div className="border-b border-slate-100 px-4 py-3 sm:border-b-0 sm:border-r">
          <dt className="text-xs font-medium uppercase tracking-wide text-slate-500">Trip</dt>
          <dd className="mt-1 text-base font-semibold text-slate-900">
            {payment?.trip_days && payment.trip_days > 0 ? payment.trip_days : tripDays}{" "}
            {(payment?.trip_days && payment.trip_days > 0 ? payment.trip_days : tripDays) === 1
              ? "day"
              : "days"}
          </dd>
          <dd className="mt-2 space-y-0.5 text-sm text-slate-600">
            <p>{fromLabel}</p>
            <p className="text-slate-400">to</p>
            <p>{toLabel}</p>
          </dd>
        </div>

        <div className="px-4 py-3">
          <dt className="flex items-center justify-between gap-2 text-xs font-medium uppercase tracking-wide text-slate-500">
            <span>Pricing</span>
            {viewerRole && payment ? (
              <PricingCalculationTip
                variant={viewerRole}
                rates={ratesFromPayment(payment)}
                className="normal-case"
              />
            ) : null}
          </dt>
          {negotiating && booking.final_booking_price ? (
            <dd className="mt-1 text-xs text-slate-500">
              Quoted rate is exclusive of platform fee and GST.
            </dd>
          ) : null}
          {booking.final_booking_price ? (
            <dd className="mt-2 flex justify-between gap-3 text-sm">
              <span className="text-slate-600">Agreed rental per day</span>
              <span className="shrink-0 tabular-nums font-medium text-slate-900">
                ₹{booking.final_booking_price}
              </span>
            </dd>
          ) : null}
          {viewerRole === "customer" && payment && !negotiating ? (
            <>
              <dd className="mt-2 flex justify-between gap-3 text-sm">
                <span className="text-slate-600">
                  Platform fee ({payment.customer_commission_percent}%)
                </span>
                <span className="shrink-0 tabular-nums font-medium text-slate-900">
                  + ₹{payment.customer_commission_inr}
                </span>
              </dd>
              <dd className="mt-1 flex justify-between gap-3 text-sm">
                <span className="text-slate-600">
                  GST ({payment.gst_percent_on_commission}% on rental + platform fee)
                </span>
                <span className="shrink-0 tabular-nums font-medium text-slate-900">
                  + ₹{payment.customer_gst_inr}
                </span>
              </dd>
            </>
          ) : null}
          {viewerRole === "owner" && payment && !negotiating ? (
            <div className={booking.final_booking_price ? "mt-2 space-y-2 border-t border-slate-100 pt-2" : "mt-2 space-y-2"}>
              <dd className="flex justify-between gap-3 text-sm">
                <span className="text-slate-600">
                  Trip rental
                  {(payment.trip_days > 0 ? payment.trip_days : tripDays) > 1
                    ? ` (${payment.trip_days > 0 ? payment.trip_days : tripDays} days)`
                    : ""}
                </span>
                <span className="shrink-0 tabular-nums font-medium text-slate-900">
                  ₹{payment.agreed_base_inr}
                </span>
              </dd>
              <dd className="flex justify-between gap-3 text-sm">
                <span className="text-slate-600">
                  Platform fee ({payment.owner_commission_percent}%)
                </span>
                <span className="shrink-0 tabular-nums text-slate-800">
                  − ₹{payment.owner_commission_inr}
                </span>
              </dd>
              <dd className="flex justify-between gap-3 text-sm">
                <span className="text-slate-600">
                  GST ({payment.gst_percent_on_commission}% on trip rental)
                </span>
                <span className="shrink-0 tabular-nums text-slate-800">
                  − ₹{payment.owner_gst_inr}
                </span>
              </dd>
              <dd className="flex justify-between gap-3 border-t border-slate-100 pt-2 text-sm font-semibold text-slate-900">
                <span>Rental payout</span>
                <span className="shrink-0 tabular-nums text-emerald-900">₹{payment.owner_net_inr}</span>
              </dd>
              {postTripInr > 0 ? (
                <dd className="flex justify-between gap-3 text-sm">
                  <span className="text-slate-600">Post-trip charges</span>
                  <span className="shrink-0 tabular-nums text-slate-800">+ ₹{postTripInr.toFixed(2)}</span>
                </dd>
              ) : null}
              {ownerEarnings ? (
                <dd className="flex justify-between gap-3 border-t border-slate-200 pt-2">
                  <span className="text-sm font-medium text-slate-900">Your earnings</span>
                  <span className="shrink-0 text-base font-semibold tabular-nums text-emerald-800">
                    ₹{ownerEarnings}
                  </span>
                </dd>
              ) : null}
              {payment.payment_status !== "PAID" && payment.owner_projected_payout_inr ? (
                <p className="text-xs text-slate-500">
                  Projected total after final pay
                  {postTripInr > 0 ? " (incl. post-trip charges)" : ""}: ₹
                  {payment.owner_projected_payout_inr}
                </p>
              ) : null}
              {payment.payment_status === "PAID" ? (
                <p className="text-xs text-slate-500">
                  Customer has completed payment
                  {payment.payment_method ? ` (${payment.payment_method})` : ""}.
                </p>
              ) : null}
            </div>
          ) : null}
          {viewerRole === "customer" && finalPrice && !negotiating ? (
            <dd
              className={`flex justify-between gap-3 ${booking.final_booking_price || payment ? "mt-1.5 border-t border-slate-100 pt-2" : "mt-2"}`}
            >
              <span className="text-sm text-slate-600">Final trip price</span>
              <span className="shrink-0 text-base font-semibold tabular-nums text-slate-900">
                ₹{finalPrice}
              </span>
            </dd>
          ) : null}
          {viewerRole === "customer" && payment && negotiating && payment.customer_total_inr ? (
            <dd className="mt-2 flex justify-between gap-3 border-t border-slate-100 pt-2 text-sm font-semibold text-slate-900">
              <span>Trip total (incl. fees)</span>
              <span className="shrink-0 tabular-nums">₹{payment.customer_total_inr}</span>
            </dd>
          ) : null}
          {payment && viewerRole === "customer" && !negotiating ? (
            <dd className="mt-0 block">
              <DepositAmountLines payment={payment} variant="customer" />
            </dd>
          ) : null}
        </div>
      </dl>
    </div>
  );
}
