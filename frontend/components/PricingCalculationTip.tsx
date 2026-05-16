"use client";

import {
  buildPricingExample,
  DEFAULT_PRICING_RATES,
  fmt,
  type PricingRates,
} from "@/lib/pricingCalculationExample";

type Props = {
  variant: "customer" | "owner";
  rates?: Partial<PricingRates>;
  className?: string;
};

export function PricingCalculationTip({ variant, rates: rateOverrides, className = "" }: Props) {
  const rates: PricingRates = { ...DEFAULT_PRICING_RATES, ...rateOverrides };
  const ex = buildPricingExample(rates);

  return (
    <details className={`relative inline-block text-left ${className}`}>
      <summary
        className="inline-flex h-5 w-5 cursor-pointer list-none items-center justify-center rounded-full border border-slate-300 bg-white text-xs font-semibold text-slate-600 hover:border-slate-400 hover:bg-slate-50 [&::-webkit-details-marker]:hidden"
        aria-label="How is this calculated? Example with ₹200 per day"
      >
        ?
      </summary>
      <div className="absolute right-0 top-full z-50 mt-1.5 w-[min(100vw-2rem,20rem)] rounded-lg border border-slate-200 bg-white p-3 text-left text-xs leading-relaxed text-slate-700 shadow-lg ring-1 ring-slate-100">
        <p className="font-semibold text-slate-900">Example (₹{ex.perDay}/day × {ex.days} days)</p>
        <p className="mt-1 text-slate-500">Your booking uses the same formulas with your agreed rate and trip length.</p>
        {variant === "customer" ? (
          <ul className="mt-2 space-y-1 tabular-nums">
            <li className="flex justify-between gap-2">
              <span>Trip rental</span>
              <span>₹{fmt(ex.tripRental)}</span>
            </li>
            <li className="flex justify-between gap-2">
              <span>+ Platform fee ({rates.customerCommissionPercent}%)</span>
              <span>₹{fmt(ex.customerPlatformFee)}</span>
            </li>
            <li className="flex justify-between gap-2">
              <span>+ GST ({rates.gstPercent}%)</span>
              <span>₹{fmt(ex.customerGst)}</span>
            </li>
            <li className="flex justify-between gap-2 border-t border-slate-100 pt-1 font-medium text-slate-900">
              <span>Trip total</span>
              <span>₹{fmt(ex.customerTotal)}</span>
            </li>
            <li className="flex justify-between gap-2 text-slate-600">
              <span>Deposit ({rates.depositPercent}%)</span>
              <span>₹{fmt(ex.depositDue)}</span>
            </li>
            <li className="flex justify-between gap-2 text-slate-600">
              <span>Balance after trip</span>
              <span>₹{fmt(ex.balanceAfterDeposit)}</span>
            </li>
          </ul>
        ) : (
          <ul className="mt-2 space-y-1 tabular-nums">
            <li className="flex justify-between gap-2">
              <span>Trip rental</span>
              <span>₹{fmt(ex.tripRental)}</span>
            </li>
            <li className="flex justify-between gap-2">
              <span>− Platform fee ({rates.ownerCommissionPercent}%)</span>
              <span>₹{fmt(ex.ownerPlatformFee)}</span>
            </li>
            <li className="flex justify-between gap-2">
              <span>− GST ({rates.gstPercent}%)</span>
              <span>₹{fmt(ex.ownerGst)}</span>
            </li>
            <li className="flex justify-between gap-2 border-t border-slate-100 pt-1 font-medium text-emerald-900">
              <span>Your earnings</span>
              <span>₹{fmt(ex.ownerEarnings)}</span>
            </li>
          </ul>
        )}
      </div>
    </details>
  );
}
