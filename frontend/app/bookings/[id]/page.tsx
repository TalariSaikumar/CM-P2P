"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { apiJson, ApiError } from "@/lib/api";
import { getToken, getUser } from "@/lib/session";
import type { Booking, Message } from "@/lib/apitypes";
import { PageLoader } from "@/components/loaders";

export default function BookingChatPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const id = params.id;
  const [booking, setBooking] = useState<Booking | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [text, setText] = useState("");
  const [price, setPrice] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [info, setInfo] = useState<string | null>(null);
  const [bootError, setBootError] = useState<string | null>(null);
  /** First booking GET finished (success or error); avoids infinite “loading” on failed load. */
  const [bookingInitDone, setBookingInitDone] = useState(false);

  const [tripFrom, setTripFrom] = useState("");
  const [tripTo, setTripTo] = useState("");
  const [tripPickup, setTripPickup] = useState("");
  const [tripDrop, setTripDrop] = useState("");
  /** When trip details are already on file, customer must open the editor explicitly. */
  const [tripEditorUserOpen, setTripEditorUserOpen] = useState(false);

  const loadBooking = useCallback(async () => {
    if (!getToken() || !id) return;
    try {
      const res = await apiJson<{ booking: Booking }>(`/bookings/${id}`);
      setBooking(res.booking);
      setBootError(null);
    } catch (e) {
      if (e instanceof ApiError) {
        setBootError((prev) => prev ?? e.message);
      }
    } finally {
      setBookingInitDone(true);
    }
  }, [id]);

  const loadMessages = useCallback(async () => {
    if (!getToken() || !id) return;
    try {
      const res = await apiJson<{ messages: Message[] }>(`/bookings/${id}/messages`);
      setMessages(res.messages || []);
    } catch {
      /* ignore */
    }
  }, [id]);

  useEffect(() => {
    setBookingInitDone(false);
    setBooking(null);
    setBootError(null);
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    if (!id) return;
    void loadBooking();
    void loadMessages();
    const t = setInterval(() => {
      void loadBooking();
      void loadMessages();
    }, 2000);
    return () => clearInterval(t);
  }, [router, id, loadBooking, loadMessages]);

  useEffect(() => {
    setTripEditorUserOpen(false);
  }, [id]);

  const lastTripInitBookingId = useRef<string | null>(null);

  useEffect(() => {
    if (!booking) return;
    const meId = getUser()?.id;
    if (!meId || meId !== booking.customer_id) return;
    if (booking.status !== "PENDING" && booking.status !== "NEGOTIATING") return;
    if (lastTripInitBookingId.current === booking.id) return;
    lastTripInitBookingId.current = booking.id;
    setTripFrom(booking.rental_from.slice(0, 10));
    setTripTo(booking.rental_to.slice(0, 10));
    setTripPickup(booking.pickup_point);
    setTripDrop(booking.drop_point);
  }, [booking]);

  function toStartOfDayUTC(dateStr: string) {
    return `${dateStr}T00:00:00.000Z`;
  }

  function toEndOfDayUTC(dateStr: string) {
    return `${dateStr}T23:59:59.999Z`;
  }

  async function saveTripDetails() {
    if (!id) return;
    if (!tripFrom || !tripTo) {
      setError("Please set rental start and end dates.");
      return;
    }
    if (tripTo < tripFrom) {
      setError("End date must be on or after the start date.");
      return;
    }
    if (!tripPickup.trim() || !tripDrop.trim()) {
      setError("Pickup and return points are required.");
      return;
    }
    setError(null);
    setInfo(null);
    try {
      await apiJson(`/bookings/${id}/trip`, {
        method: "PATCH",
        body: JSON.stringify({
          rental_from: toStartOfDayUTC(tripFrom),
          rental_to: toEndOfDayUTC(tripTo),
          pickup_point: tripPickup.trim(),
          drop_point: tripDrop.trim(),
        }),
      });
      setInfo("Trip details updated.");
      setTripEditorUserOpen(false);
      await loadBooking();
    } catch (e) {
      if (e instanceof ApiError && e.code === "CAR_ALREADY_BOOKED") {
        setError("Already booked for those dates. Pick other dates.");
      } else {
        setError(e instanceof ApiError ? e.message : "Could not update trip details");
      }
    }
  }

  async function sendMessage() {
    if (!text.trim() || !id) return;
    setError(null);
    try {
      await apiJson(`/bookings/${id}/messages`, {
        method: "POST",
        body: JSON.stringify({ body: text.trim() }),
      });
      setText("");
      await loadMessages();
      await loadBooking();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not send message");
    }
  }

  async function updatePrice() {
    if (!price.trim() || !id) return;
    setError(null);
    setInfo(null);
    try {
      await apiJson(`/bookings/${id}/price`, {
        method: "PATCH",
        body: JSON.stringify({ final_booking_price: price.trim() }),
      });
      setInfo("Final price updated. The customer will see it within a few seconds. Set the price, then confirm the booking when you are ready.");
      await loadBooking();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not update price");
    }
  }

  async function confirm() {
    if (!id) return;
    setError(null);
    setInfo(null);
    try {
      await apiJson(`/bookings/${id}/confirm`, { method: "POST" });
      setInfo("Booking confirmed. The customer may receive an SMS if notifications are configured.");
      await loadBooking();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not confirm booking");
    }
  }

  async function withdraw() {
    if (!id) return;
    setError(null);
    setInfo(null);
    try {
      await apiJson(`/bookings/${id}/withdraw`, { method: "POST" });
      setInfo("Booking withdrawn.");
      await loadBooking();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not withdraw booking");
    }
  }

  if (!bookingInitDone) {
    return (
      <main className="page-shell w-full max-w-7xl">
        <PageLoader title="Loading booking…" subtitle="Trip details, agreed price, and chat." />
      </main>
    );
  }

  if (!booking) {
    return (
      <main className="page-shell w-full max-w-7xl space-y-4">
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {bootError ??
            "We could not load this booking. Check that you are signed in and allowed to open this thread."}
        </div>
        <Link href="/customer/search" className="text-sm font-medium text-emerald-800 hover:text-emerald-900">
          ← Search cars
        </Link>
      </main>
    );
  }

  const me = getUser();
  const isOwner = me?.id === booking.owner_id;
  const isCustomer = me?.id === booking.customer_id;
  const canWithdraw =
    isCustomer &&
    (booking.status === "PENDING" || booking.status === "NEGOTIATING") &&
    !booking.final_booking_price;
  const canEditTrip =
    isCustomer && (booking.status === "PENDING" || booking.status === "NEGOTIATING");
  const needsTripDetails =
    !booking.pickup_point.trim() || !booking.drop_point.trim();
  const showTripEditor = canEditTrip && (needsTripDetails || tripEditorUserOpen);

  function cancelTripEdit() {
    setTripEditorUserOpen(false);
    setTripFrom(booking.rental_from.slice(0, 10));
    setTripTo(booking.rental_to.slice(0, 10));
    setTripPickup(booking.pickup_point);
    setTripDrop(booking.drop_point);
    setError(null);
  }

  return (
    <main className="page-shell w-full max-w-7xl">
      <div className="flex flex-col gap-4 lg:grid lg:grid-cols-[minmax(0,1fr)_minmax(17.5rem,22rem)] xl:grid-cols-[minmax(0,1fr)_minmax(20rem,26rem)] lg:items-start lg:gap-8">
        <div className="min-w-0 space-y-4">
          <div>
            <h1 className="text-xl font-semibold sm:text-2xl">Booking</h1>
            <p className="text-sm break-words text-slate-600">
              {booking.car.car_name} · {booking.car.car_model} ({booking.car.car_number}) ·{" "}
              <span className="font-medium text-slate-900">{booking.status}</span>
            </p>
            {booking.final_booking_price && (
              <p className="text-sm text-slate-700">
                Final agreed price: <span className="font-semibold">₹{booking.final_booking_price}</span>
              </p>
            )}
            {isCustomer && booking.status === "CONFIRMED" && booking.payment?.payment_status === "UNPAID" && (
              <p className="mt-2">
                <Link
                  href={`/customer/bookings/${booking.id}/pay`}
                  className="inline-flex min-h-[44px] items-center rounded-md bg-emerald-700 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-800"
                >
                  Pay ₹{booking.payment.customer_total_inr}
                </Link>
              </p>
            )}
            {isOwner && booking.status === "CONFIRMED" && booking.payment && (
              <div className="mt-3 rounded-md border border-slate-200 bg-slate-50 px-3 py-3 text-sm text-slate-700">
                <p className="font-medium text-slate-900">Your payout</p>
                <ul className="mt-2 space-y-2">
                  <li className="flex justify-between gap-4">
                    <span>Negotiated price (rental)</span>
                    <span className="shrink-0 font-medium tabular-nums text-slate-900">
                      ₹{booking.payment.agreed_base_inr}
                    </span>
                  </li>
                  <li className="flex justify-between gap-4">
                    <span>Platform fee ({booking.payment.owner_commission_percent}%)</span>
                    <span className="shrink-0 tabular-nums text-slate-800">
                      − ₹{booking.payment.owner_commission_inr}
                    </span>
                  </li>
                  <li className="flex justify-between gap-4">
                    <span>GST ({booking.payment.gst_percent_on_commission}% on negotiated rental)</span>
                    <span className="shrink-0 tabular-nums text-slate-800">
                      − ₹{booking.payment.owner_gst_inr}
                    </span>
                  </li>
                  <li className="flex justify-between gap-4 border-t border-slate-200 pt-2 text-base font-semibold text-slate-900">
                    <span>Total you get</span>
                    <span className="shrink-0 tabular-nums text-emerald-900">₹{booking.payment.owner_net_inr}</span>
                  </li>
                </ul>
                {booking.payment.payment_status === "PAID" && (
                  <p className="mt-2 text-xs text-slate-600">
                    Customer has completed payment
                    {booking.payment.payment_method ? ` (${booking.payment.payment_method})` : ""}.
                  </p>
                )}
              </div>
            )}
            <p className="text-xs text-slate-500">This page refreshes booking and chat every 2 seconds.</p>
            {showTripEditor ? (
              <div className="mt-4 rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
                <h2 className="font-medium text-slate-900">
                  {needsTripDetails ? "Add trip details" : "Update trip details"}
                </h2>
                <p className="mt-1 text-sm text-slate-600">
                  {needsTripDetails
                    ? "Enter rental dates and pickup and return points so the owner knows the trip."
                    : "You can change dates and pickup/return points while the booking is pending or negotiating."}
                </p>
                <div className="mt-3 grid gap-3 sm:grid-cols-2">
                  <div>
                    <label className="text-sm text-slate-700">From date</label>
                    <input
                      type="date"
                      className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                      value={tripFrom}
                      onChange={(e) => setTripFrom(e.target.value)}
                    />
                  </div>
                  <div>
                    <label className="text-sm text-slate-700">To date</label>
                    <input
                      type="date"
                      className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                      value={tripTo}
                      onChange={(e) => setTripTo(e.target.value)}
                    />
                  </div>
                </div>
                <div className="mt-3">
                  <label className="text-sm text-slate-700">Pickup point</label>
                  <input
                    className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                    value={tripPickup}
                    onChange={(e) => setTripPickup(e.target.value)}
                  />
                </div>
                <div className="mt-3">
                  <label className="text-sm text-slate-700">Return / drop-off point</label>
                  <input
                    className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                    value={tripDrop}
                    onChange={(e) => setTripDrop(e.target.value)}
                  />
                </div>
                <div className="mt-4 flex flex-wrap gap-2">
                  <button
                    type="button"
                    onClick={() => void saveTripDetails()}
                    className="min-h-[44px] rounded-md bg-slate-900 px-4 py-2 text-sm text-white hover:bg-slate-800"
                  >
                    Save trip details
                  </button>
                  {!needsTripDetails && (
                    <button
                      type="button"
                      onClick={cancelTripEdit}
                      className="min-h-[44px] rounded-md border border-slate-300 bg-white px-4 py-2 text-sm text-slate-800 hover:bg-slate-50"
                    >
                      Cancel
                    </button>
                  )}
                </div>
              </div>
            ) : (
              <div className="mt-4 space-y-3">
                <div className="rounded-lg border border-slate-200 bg-slate-50/80 p-3 text-sm text-slate-700">
                  <p>
                    <span className="font-medium text-slate-900">Rental:</span>{" "}
                    {new Date(booking.rental_from).toLocaleString()} → {new Date(booking.rental_to).toLocaleString()}
                  </p>
                  <p className="mt-1">
                    <span className="font-medium text-slate-900">Pickup:</span>{" "}
                    {booking.pickup_point.trim() ? (
                      booking.pickup_point
                    ) : (
                      <span className="text-slate-500">Not set</span>
                    )}
                  </p>
                  <p className="mt-1">
                    <span className="font-medium text-slate-900">Return:</span>{" "}
                    {booking.drop_point.trim() ? booking.drop_point : <span className="text-slate-500">Not set</span>}
                  </p>
                </div>
                {canEditTrip && (
                  <button
                    type="button"
                    onClick={() => {
                      setTripEditorUserOpen(true);
                      setError(null);
                    }}
                    className="min-h-[44px] rounded-md border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-900 shadow-sm hover:bg-slate-50"
                  >
                    Edit trip details
                  </button>
                )}
              </div>
            )}
          </div>

          {error && (
            <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
          )}
          {info && (
            <div className="rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-900">
              {info}
            </div>
          )}

          {isOwner && booking.status !== "CONFIRMED" && booking.status !== "CANCELLED" && (
            <div className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
              <h2 className="font-medium">Final price & confirmation</h2>
              <p className="text-sm text-slate-600">
                Set the agreed amount, then confirm the booking. Only you can confirm; the customer cannot confirm on
                your behalf.
              </p>
              <div className="mt-3 flex flex-col gap-2 sm:flex-row sm:flex-wrap">
                <input
                  className="min-h-[44px] min-w-0 flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm sm:min-w-[160px]"
                  placeholder="Amount in INR"
                  value={price}
                  onChange={(e) => setPrice(e.target.value)}
                />
                <button
                  type="button"
                  onClick={() => void updatePrice()}
                  className="min-h-[44px] shrink-0 rounded-md bg-slate-900 px-4 py-2 text-sm text-white hover:bg-slate-800 sm:min-h-0"
                >
                  Save price
                </button>
              </div>
              <button
                type="button"
                disabled={!booking.final_booking_price}
                onClick={() => void confirm()}
                className="mt-4 min-h-[44px] w-full rounded-md bg-emerald-700 px-4 py-2 text-sm text-white hover:bg-emerald-800 disabled:cursor-not-allowed disabled:opacity-50 sm:w-auto sm:min-h-0"
              >
                Confirm booking
              </button>
            </div>
          )}

          {isCustomer && canWithdraw && (
            <div className="rounded-lg border border-amber-200 bg-amber-50 p-4 shadow-sm">
              <h2 className="font-medium text-amber-950">Withdraw inquiry</h2>
              <p className="text-sm text-amber-900">
                You can withdraw while the owner has not set a final price yet. After a price is saved, withdraw is no
                longer available.
              </p>
              <button
                type="button"
                onClick={() => void withdraw()}
                className="mt-3 min-h-[44px] rounded-md border border-amber-800 bg-white px-4 py-2 text-sm font-medium text-amber-950 hover:bg-amber-100"
              >
                Withdraw booking
              </button>
            </div>
          )}
        </div>

        <aside className="min-w-0 lg:sticky lg:top-20 lg:self-start">
          <div className="flex flex-col rounded-lg border border-slate-200 bg-white p-4 shadow-sm lg:max-h-[calc(100dvh-5.5rem)] lg:min-h-[min(30rem,calc(100dvh-5.5rem))]">
            <h2 className="shrink-0 font-medium">Chat</h2>
            <div className="mt-3 flex min-h-0 flex-1 flex-col">
              <div className="min-h-0 flex-1 space-y-3 overflow-y-auto rounded-md bg-slate-50 p-3 text-sm max-h-[min(22rem,55dvh)] lg:max-h-full">
                {messages.map((m) => {
                  const isMine = me?.id === m.sender_id;
                  return (
                    <div key={m.id} className={`flex w-full ${isMine ? "justify-end" : "justify-start"}`}>
                      <div
                        className={`max-w-[min(85%,20rem)] rounded-2xl px-3 py-2 shadow-sm lg:max-w-[min(92%,28rem)] ${
                          isMine
                            ? "rounded-br-md bg-slate-900 text-white"
                            : "rounded-bl-md border border-slate-200 bg-white text-slate-900"
                        }`}
                      >
                        <p className={`text-xs ${isMine ? "text-slate-400" : "text-slate-500"}`}>
                          {isMine ? "You" : m.sender.full_name} · {new Date(m.created_at).toLocaleString()}
                        </p>
                        <p
                          className={`mt-0.5 whitespace-pre-wrap break-words ${isMine ? "text-white" : "text-slate-900"}`}
                        >
                          {m.body}
                        </p>
                      </div>
                    </div>
                  );
                })}
                {!messages.length && <p className="text-slate-500">No messages yet. Say hello.</p>}
              </div>
              <div className="mt-3 flex min-w-0 shrink-0 gap-2">
                <input
                  className="min-h-[44px] min-w-0 flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm"
                  placeholder="Type a message"
                  value={text}
                  onChange={(e) => setText(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && !e.shiftKey) {
                      e.preventDefault();
                      void sendMessage();
                    }
                  }}
                />
                <button
                  type="button"
                  onClick={() => void sendMessage()}
                  className="min-h-[44px] shrink-0 rounded-md bg-slate-900 px-4 py-2 text-sm text-white hover:bg-slate-800"
                >
                  Send
                </button>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </main>
  );
}
