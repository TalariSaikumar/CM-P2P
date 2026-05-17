"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { apiJson, ApiError } from "@/lib/api";
import { getToken, getUser } from "@/lib/session";
import type { Booking, BookingPaymentBreakdown } from "@/lib/apitypes";
import { ButtonCarSpinner, OverlayLoader, PageLoader } from "@/components/loaders";
import { PricingCalculationTip } from "@/components/PricingCalculationTip";
import { ratesFromPayment } from "@/lib/pricingCalculationExample";
import { isMobilePaymentDevice, openRazorpayCheckout } from "@/lib/razorpay";
import { type UpiAppId, UpiAppPicker } from "@/components/UpiAppPicker";

const METHODS = [
  {
    value: "UPI",
    label: "UPI",
    hint: "Pick Google Pay, PhonePe, Paytm, or another app on your phone",
  },
  { value: "CARD", label: "Card", hint: "Visa, Mastercard, RuPay" },
  { value: "NET_BANKING", label: "Net banking", hint: "Your bank’s login" },
  { value: "QR_CODE", label: "QR code", hint: "Scan with any UPI app" },
] as const;

function digitsOnly(s: string): string {
  return s.replace(/\D/g, "");
}

function validateMethodFields(method: string, fields: Record<string, string>): string | null {
  switch (method) {
    case "UPI": {
      if (!fields.upiApp.trim()) return "Select a UPI app installed on your device.";
      return null;
    }
    case "CARD": {
      const num = digitsOnly(fields.cardNumber);
      if (num.length < 12 || num.length > 19) return "Enter a valid card number (12–19 digits).";
      const exp = fields.cardExpiry.trim();
      if (!/^\d{2}\/\d{2}$/.test(exp)) return "Expiry must be MM/YY.";
      const cvv = digitsOnly(fields.cardCvv);
      if (cvv.length < 3 || cvv.length > 4) return "Enter a valid CVV (3 or 4 digits).";
      if (!fields.cardName.trim()) return "Enter the name on the card.";
      return null;
    }
    case "NET_BANKING": {
      if (!fields.netbankUser.trim()) return "Enter your bank customer ID or login.";
      if (!fields.netbankPass.trim()) return "Enter your password or secure key.";
      return null;
    }
    case "QR_CODE": {
      if (fields.qrConfirmed !== "yes") return "Confirm you have scanned the QR code and completed payment.";
      return null;
    }
    default:
      return "Select a payment method.";
  }
}

export default function CustomerPayBookingPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const id = params?.id ?? "";

  const [booking, setBooking] = useState<Booking | null>(null);
  const [breakdown, setBreakdown] = useState<BookingPaymentBreakdown | null>(null);
  const [method, setMethod] = useState<string>("UPI");
  const [loadErr, setLoadErr] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [paying, setPaying] = useState(false);
  const [loadSettled, setLoadSettled] = useState(false);

  const [selectedUpiApp, setSelectedUpiApp] = useState<UpiAppId | null>(null);
  const [cardNumber, setCardNumber] = useState("");
  const [cardExpiry, setCardExpiry] = useState("");
  const [cardCvv, setCardCvv] = useState("");
  const [cardName, setCardName] = useState("");
  const [netbankUser, setNetbankUser] = useState("");
  const [netbankPass, setNetbankPass] = useState("");
  const [qrConfirmed, setQrConfirmed] = useState(false);

  const load = useCallback(async () => {
    if (!id) {
      setLoadSettled(true);
      return;
    }
    setLoadErr(null);
    try {
      const res = await apiJson<{ booking: Booking }>(`/bookings/${id}`);
      if (getUser()?.id !== res.booking.customer_id) {
        setBooking(null);
        setBreakdown(null);
        setLoadErr("Only the customer on this booking can complete payment here.");
        return;
      }
      if (!res.booking.payment) {
        setBooking(null);
        setBreakdown(null);
        setLoadErr("Payment opens after the owner confirms this booking and an agreed price is set.");
        return;
      }
      setBooking(res.booking);
      setBreakdown(res.booking.payment);
    } catch (e) {
      setLoadErr(e instanceof ApiError ? e.message : "Could not load payment");
      setBooking(null);
      setBreakdown(null);
    } finally {
      setLoadSettled(true);
    }
  }, [id]);

  useEffect(() => {
    setLoadSettled(false);
  }, [id]);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    const u = getUser();
    if (u?.role !== "CUSTOMER") {
      router.replace("/owner/fleet");
      return;
    }
    void load();
  }, [router, load]);

  const field =
    "mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-500 focus:outline-none focus:ring-1 focus:ring-slate-500";

  const useRazorpay =
    breakdown?.checkout_provider === "razorpay" && Boolean(breakdown?.razorpay_key_id?.trim());

  async function pay() {
    if (!id || !breakdown) return;

    if (!useRazorpay) {
      const v = validateMethodFields(method, {
        upiApp: selectedUpiApp ?? "",
        cardNumber,
        cardExpiry,
        cardCvv,
        cardName,
        netbankUser,
        netbankPass,
        qrConfirmed: qrConfirmed ? "yes" : "",
      });
      if (v) {
        setError(v);
        return;
      }
    }

    setError(null);
    setPaying(true);
    try {
      if (useRazorpay) {
        const keyId = breakdown.razorpay_key_id!.trim();
        const orderRes = await apiJson<{
          order_id: string;
          amount_paise: number;
          currency: string;
          key_id: string;
        }>(`/bookings/${id}/payment-order`, { method: "POST" });

        const user = getUser();
        const rz = await openRazorpayCheckout({
          keyId: orderRes.key_id || keyId,
          orderId: orderRes.order_id,
          amountPaise: orderRes.amount_paise,
          currency: orderRes.currency,
          paymentMethod: method,
          description:
            breakdown.payment_phase === "final_due"
              ? "CarManage — final trip balance"
              : "CarManage — trip deposit",
          prefillEmail: user?.email,
          prefillContact: user?.phone_number,
        });

        await apiJson(`/bookings/${id}/pay`, {
          method: "POST",
          body: JSON.stringify({
            payment_method: method,
            razorpay_order_id: rz.razorpay_order_id,
            razorpay_payment_id: rz.razorpay_payment_id,
            razorpay_signature: rz.razorpay_signature,
          }),
        });
      } else {
        await apiJson(`/bookings/${id}/pay`, {
          method: "POST",
          body: JSON.stringify({ payment_method: method }),
        });
      }
      router.push(`/bookings/${id}`);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : e instanceof Error ? e.message : "Payment failed");
    } finally {
      setPaying(false);
    }
  }

  if (!loadSettled) {
    return (
      <main className="page-shell w-full max-w-6xl">
        <PageLoader title="Loading checkout…" subtitle="Fetching your booking and payment summary." />
      </main>
    );
  }

  if (loadErr) {
    return (
      <main className="page-shell w-full max-w-3xl space-y-4">
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{loadErr}</div>
        <Link href="/customer/bookings" className="text-sm font-medium text-emerald-800 hover:text-emerald-900">
          ← My bookings
        </Link>
      </main>
    );
  }

  if (!booking || !breakdown) {
    return (
      <main className="page-shell w-full max-w-6xl space-y-4">
        <p className="text-sm text-slate-600">Payment details could not be loaded.</p>
        <Link href="/customer/bookings" className="text-sm font-medium text-emerald-800 hover:text-emerald-900">
          ← My bookings
        </Link>
      </main>
    );
  }

  const phase = breakdown.payment_phase ?? "";
  const payStatus = booking.payment?.payment_status ?? "";

  if (payStatus === "PAID") {
    return (
      <main className="page-shell w-full max-w-3xl space-y-4">
        <h1 className="text-xl font-semibold">Already paid</h1>
        <p className="text-sm text-slate-600">This booking is fully settled.</p>
        <Link href={`/bookings/${id}`} className="text-sm font-medium text-emerald-800 hover:text-emerald-900">
          ← Back to booking
        </Link>
      </main>
    );
  }

  if (payStatus === "DEPOSIT_PAID" && phase === "awaiting_settlement") {
    return (
      <main className="page-shell w-full max-w-3xl space-y-4">
        <h1 className="text-xl font-semibold">Waiting for final bill</h1>
        <p className="mt-2 text-sm text-slate-600">
          Your <strong>75% deposit</strong> is recorded. After the trip, the owner will verify the vehicle and submit
          any tolls, fines, or damage charges if applicable. You will return here to pay the remaining balance.
        </p>
        <Link href={`/bookings/${id}`} className="text-sm font-medium text-emerald-800 hover:text-emerald-900">
          ← Back to booking
        </Link>
      </main>
    );
  }

  const dueNowInr =
    phase === "final_due" && breakdown.final_due_inr
      ? breakdown.final_due_inr
      : breakdown.deposit_due_inr || breakdown.customer_total_inr;

  return (
    <>
      {paying ? <OverlayLoader message="Completing payment…" /> : null}
      <main className="page-shell w-full max-w-6xl pb-12">
      <div className="border-b border-slate-200/80 pb-6">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-900 sm:text-3xl">
          {phase === "final_due" ? "Pay final balance" : "Pay trip deposit"}
        </h1>
        <p className="mt-2 max-w-3xl text-sm text-slate-600 sm:text-base">
          {booking.car.car_name} · {booking.car.car_model}
          {phase === "final_due" ? (
            <>
              {" "}
              — you are paying the <strong>remaining trip balance</strong> plus any post-trip charges the owner
              submitted
              {useRazorpay ? " via Razorpay." : " (demo checkout; no real charge)."}
            </>
          ) : (
            <>
              {" "}
              — you pay <strong>{breakdown.deposit_percent ?? 75}%</strong> of your trip total now as a deposit. The
              rest is due after the trip when the owner confirms any extra charges
              {useRazorpay ? " via Razorpay." : " (demo checkout; no real charge)."}
            </>
          )}
        </p>
      </div>

      <div className="mt-8 grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(300px,400px)] lg:items-start">
        {/* Summary first on mobile so users see totals; sticky on desktop right */}
        <aside className="order-1 space-y-4 lg:order-2 lg:sticky lg:top-6">
          <div className="rounded-xl border border-slate-200 bg-gradient-to-b from-slate-50 to-white p-5 shadow-sm ring-1 ring-slate-100">
            <div className="flex items-center justify-between gap-2">
              <h2 className="text-xs font-semibold uppercase tracking-wide text-slate-500">Order summary</h2>
              <PricingCalculationTip variant="customer" rates={ratesFromPayment(breakdown)} />
            </div>
            <ul className="mt-4 space-y-3 text-sm text-slate-700">
              <li className="flex justify-between gap-4">
                <span>Agreed rental per day</span>
                <span className="shrink-0 font-medium tabular-nums text-slate-900">
                  ₹{booking.final_booking_price}
                </span>
              </li>
              <li className="flex justify-between gap-4">
                <span>Trip rental (incl. days)</span>
                <span className="shrink-0 font-medium tabular-nums text-slate-900">₹{breakdown.agreed_base_inr}</span>
              </li>
              <li className="flex justify-between gap-4">
                <span>Platform fee ({breakdown.customer_commission_percent}%)</span>
                <span className="shrink-0 tabular-nums text-slate-800">+ ₹{breakdown.customer_commission_inr}</span>
              </li>
              <li className="flex justify-between gap-4">
                <span>GST ({breakdown.gst_percent_on_commission}% on rental + platform fee)</span>
                <span className="shrink-0 tabular-nums text-slate-800">+ ₹{breakdown.customer_gst_inr}</span>
              </li>
              <li className="flex justify-between gap-4 border-t border-slate-200 pt-3 text-base font-semibold text-slate-900">
                <span>Trip total (incl. fees)</span>
                <span className="shrink-0 text-lg tabular-nums text-slate-900">₹{breakdown.customer_total_inr}</span>
              </li>
              {phase === "final_due" ? (
                <>
                  {Number(breakdown.deposit_paid_inr) > 0 && (
                    <li className="flex justify-between gap-4 text-sm text-slate-700">
                      <span>Deposit paid ({breakdown.deposit_percent ?? 75}%)</span>
                      <span className="shrink-0 tabular-nums text-slate-800">− ₹{breakdown.deposit_paid_inr}</span>
                    </li>
                  )}
                  {Number(breakdown.post_trip_charges_inr) > 0 && (
                    <li className="flex justify-between gap-4 text-sm text-slate-700">
                      <span>Post-trip charges (owner)</span>
                      <span className="shrink-0 tabular-nums text-slate-800">+ ₹{breakdown.post_trip_charges_inr}</span>
                    </li>
                  )}
                  {breakdown.post_trip_items && breakdown.post_trip_items.length > 0 && (
                    <li className="border-t border-slate-100 pt-2 text-xs text-slate-600">
                      <p className="font-medium text-slate-700">Charge lines</p>
                      <ul className="mt-1 space-y-1">
                        {breakdown.post_trip_items.map((it, i) => (
                          <li key={`${it.label}-${i}`} className="flex justify-between gap-2">
                            <span className="min-w-0 truncate">{it.label}</span>
                            <span className="shrink-0 tabular-nums">₹{it.amount_inr}</span>
                          </li>
                        ))}
                      </ul>
                    </li>
                  )}
                  <li className="flex justify-between gap-4 border-t border-slate-200 pt-3 text-base font-semibold text-emerald-950">
                    <span>Due now</span>
                    <span className="shrink-0 text-lg tabular-nums text-emerald-900">₹{dueNowInr}</span>
                  </li>
                </>
              ) : (
                <li className="flex justify-between gap-4 border-t border-slate-200 pt-3 text-base font-semibold text-emerald-950">
                  <span>Due now ({breakdown.deposit_percent ?? 75}% deposit)</span>
                  <span className="shrink-0 text-lg tabular-nums text-emerald-900">₹{dueNowInr}</span>
                </li>
              )}
            </ul>
          </div>
        </aside>

        <div className="order-2 space-y-6 lg:order-1">
          <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm sm:p-6">
            <h2 className="text-base font-semibold text-slate-900">Payment method</h2>
            <p className="mt-1 text-sm text-slate-500">
              {useRazorpay
                ? method === "UPI" && isMobilePaymentDevice()
                  ? "Pay with UPI — you will pick an app installed on your phone (Google Pay, PhonePe, Paytm, etc.) in the next step."
                  : "Choose how you want to pay — Razorpay Checkout opens when you continue."
                : "Choose how you want to pay — fields appear for the option you select."}
            </p>

            <div className="mt-5 grid gap-3 sm:grid-cols-2">
              {METHODS.map((m) => (
                <label
                  key={m.value}
                  className={`flex cursor-pointer flex-col rounded-lg border-2 p-4 transition-colors ${
                    method === m.value
                      ? "border-slate-900 bg-slate-50 ring-1 ring-slate-900/10"
                      : "border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50/50"
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <input
                      type="radio"
                      name="pm"
                      value={m.value}
                      checked={method === m.value}
                      onChange={() => {
                        setMethod(m.value);
                        setError(null);
                      }}
                      className="h-4 w-4 shrink-0 border-slate-300 text-slate-900"
                    />
                    <span className="font-medium text-slate-900">{m.label}</span>
                  </div>
                  <span className="mt-1 pl-7 text-xs text-slate-500">{m.hint}</span>
                </label>
              ))}
            </div>

            {useRazorpay && method === "UPI" && (
              <div className="mt-6 rounded-lg border border-emerald-200 bg-emerald-50/80 px-4 py-3 text-sm text-emerald-950">
                <p className="font-medium">UPI apps on your device</p>
                <p className="mt-1 text-emerald-900/90">
                  {isMobilePaymentDevice()
                    ? "After you tap Pay, Razorpay will list UPI apps on your phone — choose Google Pay, PhonePe, Paytm, or another installed app."
                    : "On desktop, Razorpay will show a QR code — scan it with any UPI app on your phone to pay."}
                </p>
              </div>
            )}

            {!useRazorpay && (
            <div className="mt-6 border-t border-slate-100 pt-6">
              {method === "UPI" && (
                <div>
                  <h3 className="text-sm font-medium text-slate-800">Choose your UPI app</h3>
                  <div className="mt-3">
                    <UpiAppPicker
                      selected={selectedUpiApp}
                      onSelect={(appId) => {
                        setSelectedUpiApp(appId);
                        setError(null);
                      }}
                      disabled={paying}
                    />
                  </div>
                </div>
              )}

              {method === "CARD" && (
                <div className="space-y-4">
                  <h3 className="text-sm font-medium text-slate-800">Card details</h3>
                  <label className="block text-sm text-slate-700">
                    Name on card
                    <input
                      className={field}
                      value={cardName}
                      onChange={(e) => setCardName(e.target.value)}
                      placeholder="As printed on card"
                      autoComplete="cc-name"
                    />
                  </label>
                  <label className="block text-sm text-slate-700">
                    Card number
                    <input
                      className={field}
                      inputMode="numeric"
                      value={cardNumber}
                      onChange={(e) => setCardNumber(e.target.value)}
                      placeholder="1234 5678 9012 3456"
                      autoComplete="cc-number"
                    />
                  </label>
                  <div className="grid gap-4 sm:grid-cols-2">
                    <label className="block text-sm text-slate-700">
                      Expiry (MM/YY)
                      <input
                        className={field}
                        value={cardExpiry}
                        onChange={(e) => {
                          let v = e.target.value.replace(/\D/g, "").slice(0, 4);
                          if (v.length >= 2) v = `${v.slice(0, 2)}/${v.slice(2)}`;
                          setCardExpiry(v);
                        }}
                        placeholder="MM/YY"
                        autoComplete="cc-exp"
                      />
                    </label>
                    <label className="block text-sm text-slate-700">
                      CVV
                      <input
                        className={field}
                        type="password"
                        inputMode="numeric"
                        maxLength={4}
                        value={cardCvv}
                        onChange={(e) => setCardCvv(e.target.value.replace(/\D/g, "").slice(0, 4))}
                        placeholder="•••"
                        autoComplete="cc-csc"
                      />
                    </label>
                  </div>
                </div>
              )}

              {method === "NET_BANKING" && (
                <div className="space-y-4">
                  <h3 className="text-sm font-medium text-slate-800">Net banking</h3>
                  <label className="block text-sm text-slate-700">
                    Customer ID / login
                    <input
                      className={field}
                      value={netbankUser}
                      onChange={(e) => setNetbankUser(e.target.value)}
                      placeholder="Your bank login ID"
                      autoComplete="username"
                    />
                  </label>
                  <label className="block text-sm text-slate-700">
                    Password / PIN
                    <input
                      className={field}
                      type="password"
                      value={netbankPass}
                      onChange={(e) => setNetbankPass(e.target.value)}
                      placeholder="••••••••"
                      autoComplete="current-password"
                    />
                  </label>
                  <p className="text-xs text-slate-500">
                    Demo only — nothing is sent to a bank. Do not enter real credentials.
                  </p>
                </div>
              )}

              {method === "QR_CODE" && (
                <div>
                  <h3 className="text-sm font-medium text-slate-800">Scan to pay</h3>
                  <p className="mt-1 text-sm text-slate-600">
                    Open Google Pay, PhonePe, Paytm, or any UPI app and scan this code for ₹{dueNowInr}.
                  </p>
                  <div
                    className="mx-auto mt-4 flex h-44 w-44 items-center justify-center rounded-lg border-2 border-dashed border-slate-300 bg-white p-3"
                    aria-hidden
                  >
                    <svg viewBox="0 0 120 120" className="h-full w-full text-slate-800">
                      <rect width="120" height="120" fill="#fff" />
                      <rect x="8" y="8" width="28" height="28" fill="currentColor" />
                      <rect x="84" y="8" width="28" height="28" fill="currentColor" />
                      <rect x="8" y="84" width="28" height="28" fill="currentColor" />
                      <rect x="44" y="44" width="10" height="10" fill="currentColor" />
                      <rect x="58" y="44" width="8" height="8" fill="currentColor" />
                      <rect x="44" y="58" width="8" height="8" fill="currentColor" />
                      <rect x="66" y="58" width="12" height="12" fill="currentColor" />
                      <rect x="52" y="72" width="14" height="14" fill="currentColor" />
                      <rect x="72" y="44" width="8" height="24" fill="currentColor" />
                      <rect x="44" y="72" width="6" height="20" fill="currentColor" />
                    </svg>
                  </div>
                  <p className="mt-3 text-center text-xs text-slate-500">
                    Demo placeholder — with Razorpay enabled, the live QR opens in checkout.
                  </p>
                  <label className="mt-4 flex cursor-pointer items-start gap-2 text-sm text-slate-700">
                    <input
                      type="checkbox"
                      checked={qrConfirmed}
                      onChange={(e) => setQrConfirmed(e.target.checked)}
                      className="mt-0.5 h-4 w-4 rounded border-slate-300 text-slate-900"
                    />
                    <span>I have scanned the QR code and paid ₹{dueNowInr}</span>
                  </label>
                </div>
              )}
            </div>
            )}
          </div>

          {error && (
            <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
          )}

          <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
            <button
              type="button"
              disabled={paying}
              onClick={() => void pay()}
              className="inline-flex min-h-[48px] min-w-[200px] items-center justify-center gap-2 rounded-lg bg-emerald-700 px-6 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-emerald-800 disabled:opacity-60"
            >
              {paying ? (
                <>
                  <ButtonCarSpinner className="text-emerald-100" />
                  Processing…
                </>
              ) : (
                `Pay ₹${dueNowInr}`
              )}
            </button>
            <Link
              href={`/bookings/${id}`}
              className="min-h-[48px] inline-flex items-center justify-center rounded-lg border border-slate-300 px-6 py-2.5 text-center text-sm font-medium text-slate-800 hover:bg-slate-50"
            >
              Cancel
            </Link>
          </div>
        </div>
      </div>
    </main>
    </>
  );
}
