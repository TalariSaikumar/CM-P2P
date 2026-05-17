"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { apiJson, ApiError } from "@/lib/api";
import { getToken, getUser } from "@/lib/session";
import type { Booking, Message } from "@/lib/apitypes";
import { PageLoader } from "@/components/loaders";
import {
  HandoverPhotoField,
  uploadPendingHandoverPhotos,
  type HandoverPhotoStep,
} from "@/components/HandoverPhotoField";
import { BookingSummaryCard } from "@/components/BookingSummaryCard";
import { PostTripChargesSummary } from "@/components/PostTripChargesSummary";
import { isTripActive, isTripCompleted } from "@/lib/bookingStatus";
import { NegotiationPricePreview } from "@/components/NegotiationPricePreview";
import { endDateAfterStartChange, localDateInputValue, tripDaysInclusive } from "@/lib/rentalDates";

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

  const [cancelReason, setCancelReason] = useState("");
  const [pickupOdom, setPickupOdom] = useState("");
  const [pickupFuel, setPickupFuel] = useState("");
  const [pickupNotesH, setPickupNotesH] = useState("");
  const [returnOdom, setReturnOdom] = useState("");
  const [returnFuel, setReturnFuel] = useState("");
  const [returnNotesH, setReturnNotesH] = useState("");
  const [reviewRating, setReviewRating] = useState(5);
  const [reviewComment, setReviewComment] = useState("");
  const [chargeDrafts, setChargeDrafts] = useState<{ label: string; amount: string }[]>([{ label: "", amount: "" }]);
  const [settlementBusy, setSettlementBusy] = useState(false);

  const ownerPickupPhotosRef = useRef<File[]>([]);
  const customerPickupPhotosRef = useRef<File[]>([]);
  const customerReturnPhotosRef = useRef<File[]>([]);
  const ownerReturnPhotosRef = useRef<File[]>([]);

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

  async function submitPostTripCharges() {
    if (!id) return;
    setError(null);
    setInfo(null);
    const items = chargeDrafts
      .filter((r) => r.label.trim() && r.amount.trim())
      .map((r) => ({ label: r.label.trim(), amount_inr: r.amount.trim() }));
    setSettlementBusy(true);
    try {
      const res = await apiJson<{ booking: Booking }>(`/bookings/${id}/post-trip-charges`, {
        method: "PUT",
        body: JSON.stringify({ items }),
      });
      setBooking(res.booking);
      const next = res.booking.payment?.post_trip_items;
      if (next?.length) {
        setChargeDrafts(next.map((i) => ({ label: i.label, amount: i.amount_inr })));
      } else {
        setChargeDrafts([{ label: "", amount: "" }]);
      }
      setInfo("Post-trip charges saved. The customer can pay the final balance.");
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not save charges");
    } finally {
      setSettlementBusy(false);
    }
  }

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

  const lastPostTripItemsKey = useRef("");

  useEffect(() => {
    const items = booking?.payment?.post_trip_items;
    if (!items) return;
    const key = items.map((i) => `${i.label}\0${i.amount_inr}`).join("\n");
    if (key === lastPostTripItemsKey.current) return;
    lastPostTripItemsKey.current = key;
    if (items.length > 0) {
      setChargeDrafts(items.map((i) => ({ label: i.label, amount: i.amount_inr })));
    }
  }, [booking?.payment?.post_trip_items]);

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
      setInfo(
        "Quoted price saved. The customer must accept it in chat before you can confirm the booking.",
      );
      await loadBooking();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not update price");
    }
  }

  async function acceptQuotedPrice() {
    if (!id) return;
    setError(null);
    setInfo(null);
    try {
      const res = await apiJson<{ booking: Booking }>(`/bookings/${id}/accept-price`, { method: "POST" });
      setBooking(res.booking);
      setInfo("You accepted the quoted price. The owner can now confirm the booking.");
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not accept price");
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

  async function cancelBooking() {
    if (!id) return;
    setError(null);
    setInfo(null);
    try {
      const res = await apiJson<{ booking: Booking }>(`/bookings/${id}/cancel`, {
        method: "POST",
        body: JSON.stringify({ reason: cancelReason.trim() || undefined }),
      });
      setBooking(res.booking);
      setInfo("Booking cancelled.");
      setCancelReason("");
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not cancel booking");
    }
  }

  async function flushHandoverPhotos(step: HandoverPhotoStep, filesRef: React.MutableRefObject<File[]>) {
    if (!id || filesRef.current.length === 0) return;
    const files = [...filesRef.current];
    filesRef.current = [];
    try {
      const updated = await uploadPendingHandoverPhotos(id, step, files);
      if (updated) setBooking(updated);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not upload photos");
    }
  }

  async function submitHandover(phase: "pickup" | "return") {
    if (!id || !booking) return;
    const odom = phase === "pickup" ? pickupOdom.trim() : returnOdom.trim();
    const n = parseInt(odom, 10);
    if (!Number.isFinite(n) || n <= 0) {
      setError("Enter a valid odometer reading (km).");
      return;
    }
    let fuelPct: number | undefined;
    const fuelStr = phase === "pickup" ? pickupFuel.trim() : returnFuel.trim();
    if (fuelStr) {
      const f = parseInt(fuelStr, 10);
      if (!Number.isFinite(f) || f < 0 || f > 100) {
        setError("Fuel percent must be between 0 and 100 if provided.");
        return;
      }
      fuelPct = f;
    }
    setError(null);
    setInfo(null);
    try {
      const res = await apiJson<{ booking: Booking }>(`/bookings/${id}/handover`, {
        method: "PATCH",
        body: JSON.stringify({
          phase,
          odometer_km: n,
          fuel_percent: fuelPct ?? null,
          notes: phase === "pickup" ? pickupNotesH.trim() : returnNotesH.trim(),
        }),
      });
      setBooking(res.booking);
      const me = getUser();
      const actingOwner = me?.id === booking.owner_id;
      const actingCustomer = me?.id === booking.customer_id;
      if (phase === "pickup") {
        if (actingOwner) await flushHandoverPhotos("owner_pickup", ownerPickupPhotosRef);
        if (actingCustomer) await flushHandoverPhotos("customer_pickup", customerPickupPhotosRef);
        const ownerDone = !!res.booking.handover?.owner_pickup_recorded_at;
        const customerDone = !!res.booking.handover?.pickup_recorded_at;
        if (customerDone) setInfo("Pickup check-in saved.");
        else if (ownerDone) setInfo("Vehicle handover recorded.");
        else setInfo("Pickup handover saved.");
      } else {
        if (actingCustomer) await flushHandoverPhotos("customer_return", customerReturnPhotosRef);
        if (res.booking.handover?.owner_return_accepted_at) {
          setInfo("Return accepted — trip complete.");
        } else {
          setInfo("Return check-in saved.");
        }
      }
      if (phase === "pickup") {
        setPickupOdom("");
        setPickupFuel("");
        setPickupNotesH("");
      } else {
        setReturnOdom("");
        setReturnFuel("");
        setReturnNotesH("");
      }
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not save handover");
    }
  }

  async function submitOwnerAcceptReturn() {
    if (!id || !booking?.handover?.return_odometer_km) return;
    setError(null);
    setInfo(null);
    try {
      const res = await apiJson<{ booking: Booking }>(`/bookings/${id}/handover`, {
        method: "PATCH",
        body: JSON.stringify({
          phase: "return",
          odometer_km: booking.handover.return_odometer_km,
          fuel_percent: booking.handover.return_fuel_percent ?? null,
          notes: "",
        }),
      });
      setBooking(res.booking);
      await flushHandoverPhotos("owner_return", ownerReturnPhotosRef);
      setInfo("Return accepted — trip complete.");
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not accept return");
    }
  }

  async function submitReview() {
    if (!id) return;
    setError(null);
    setInfo(null);
    try {
      const res = await apiJson<{ booking: Booking }>(`/bookings/${id}/reviews`, {
        method: "POST",
        body: JSON.stringify({ rating: reviewRating, comment: reviewComment.trim() }),
      });
      setBooking(res.booking);
      setInfo("Thanks — your review was posted.");
      setReviewComment("");
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not post review");
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

  const payStatus = booking.payment?.payment_status ?? "";
  const payPhase = booking.payment?.payment_phase ?? "";
  const isFullyPaid = payStatus === "PAID";
  const rentalEnded = new Date(booking.rental_to).getTime() < Date.now();
  const canCancelBooking =
    booking.status !== "CANCELLED" &&
    (booking.status === "PENDING" ||
      booking.status === "NEGOTIATING" ||
      (booking.status === "CONFIRMED" && payStatus === "UNPAID"));
  const myReviewParty = isCustomer ? "CUSTOMER" : isOwner ? "OWNER" : "";
  const hasMyReview = booking.reviews?.some((r) => r.party === myReviewParty) ?? false;
  const canPostReview =
    isTripActive(booking.status) &&
    isFullyPaid &&
    rentalEnded &&
    (isCustomer || isOwner) &&
    !hasMyReview;
  const depositPaidForHandover =
    payStatus === "DEPOSIT_PAID" || payStatus === "FINAL_DUE" || payStatus === "PAID";
  const showHandover = isTripActive(booking.status) && depositPaidForHandover;
  const tripStageLabel = booking.trip_stage_label?.trim() ?? "";
  const ownerHandoverDone = !!booking.handover?.owner_pickup_recorded_at;
  const customerPickupDone = !!booking.handover?.pickup_recorded_at;
  const pickupComplete = ownerHandoverDone && customerPickupDone;
  const customerReturnDone = !!booking.handover?.return_recorded_at;
  const ownerReturnAccepted = !!booking.handover?.owner_return_accepted_at;
  const returnComplete = customerReturnDone && ownerReturnAccepted;
  const showOwnerHandoverForm = showHandover && !ownerHandoverDone && isOwner;
  const showCustomerAwaitingOwner =
    showHandover && !ownerHandoverDone && isCustomer;
  const showHandoverCard = showHandover && ownerHandoverDone;
  const handoverPhotos = booking?.handover?.photos ?? [];
  const tripDays = tripDaysInclusive(booking.rental_from, booking.rental_to);
  const postTripInr = Number(booking.payment?.post_trip_charges_inr ?? 0);
  const customerTripTotalInr = booking.payment
    ? (Number(booking.payment.customer_total_inr) + (Number.isFinite(postTripInr) ? postTripInr : 0)).toFixed(2)
    : null;
  const showWithdrawBlock = isCustomer && canWithdraw;
  const showCancelBlock = canCancelBooking && (isOwner || isCustomer) && !showWithdrawBlock;
  const customerNeedsDeposit =
    isCustomer && booking.status === "CONFIRMED" && payStatus === "UNPAID" && booking.payment;
  const customerNeedsFinal =
    isCustomer && booking.status === "CONFIRMED" && payStatus === "FINAL_DUE" && booking.payment;
  const showOwnerSettlement =
    isOwner &&
    booking.status === "CONFIRMED" &&
    (payStatus === "DEPOSIT_PAID" || payStatus === "FINAL_DUE") &&
    customerReturnDone;
  const customerPriceAccepted = !!booking.customer_price_accepted;
  const canCustomerAcceptPrice =
    isCustomer &&
    (booking.status === "PENDING" || booking.status === "NEGOTIATING") &&
    !!booking.final_booking_price &&
    !customerPriceAccepted;
  const ownerCanConfirm =
    isOwner &&
    !!booking.final_booking_price &&
    customerPriceAccepted &&
    booking.status !== "CONFIRMED" &&
    booking.status !== "CANCELLED";

  function cancelTripEdit() {
    if (!booking) return;
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
            <BookingSummaryCard
              booking={booking}
              tripDays={tripDays}
              customerTripTotalInr={customerTripTotalInr}
              postTripInr={postTripInr}
              viewerRole={isOwner ? "owner" : isCustomer ? "customer" : null}
            />
            {customerNeedsDeposit && booking.payment && (
              <p className="mt-2">
                <Link
                  href={`/customer/bookings/${booking.id}/pay`}
                  className="inline-flex min-h-[44px] items-center rounded-md bg-emerald-700 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-800"
                >
                  Pay deposit ₹{booking.payment.deposit_due_inr ?? booking.payment.customer_total_inr}
                </Link>
              </p>
            )}
            {customerNeedsFinal && booking.payment && (
              <p className="mt-2">
                <Link
                  href={`/customer/bookings/${booking.id}/pay`}
                  className="inline-flex min-h-[44px] items-center rounded-md bg-emerald-700 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-800"
                >
                  Pay final balance ₹{booking.payment.final_due_inr}
                </Link>
              </p>
            )}
            {isCustomer &&
              booking.status === "CONFIRMED" &&
              payStatus === "DEPOSIT_PAID" &&
              payPhase === "awaiting_settlement" &&
              customerReturnDone && (
                <p className="mt-2 rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">
                  Deposit paid and return logged. Waiting for the owner to submit post-trip charges (tolls, fines,
                  damage, etc.). You will get a <strong>Pay final balance</strong> button when the bill is ready.
                </p>
              )}
            {isOwner &&
              booking.status === "CONFIRMED" &&
              payStatus === "UNPAID" &&
              booking.payment && (
                <p className="mt-2 rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">
                  <span className="font-medium text-slate-900">Next for you:</span> Wait for the customer to pay the trip
                  deposit. Pickup &amp; return handover will appear here after that.
                </p>
              )}
            {isCustomer &&
              booking.status === "CONFIRMED" &&
              payStatus === "UNPAID" &&
              booking.payment && (
                <p className="mt-2 rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">
                  <span className="font-medium text-slate-900">Next for you:</span> Pay the deposit below to lock the
                  booking. Pickup &amp; return handover unlocks after payment.
                </p>
              )}
            {isTripActive(booking.status) && booking.payment && (isCustomer || isOwner) && (
              <PostTripChargesSummary
                payment={booking.payment}
                variant={isCustomer ? "customer" : "owner"}
              />
            )}
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
                      min={localDateInputValue()}
                      value={tripFrom}
                      onChange={(e) => {
                        const next = e.target.value;
                        setTripFrom(next);
                        setTripTo((prev) => endDateAfterStartChange(next, prev));
                      }}
                    />
                  </div>
                  <div>
                    <label className="text-sm text-slate-700">To date</label>
                    <input
                      type="date"
                      className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                      min={tripFrom || localDateInputValue()}
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
            ) : canEditTrip ? (
              <div className="mt-4">
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
              </div>
            ) : null}
          </div>

          {error && (
            <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
          )}
          {info && (
            <div className="rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-900">
              {info}
            </div>
          )}

          {canCustomerAcceptPrice && (
            <div className="rounded-lg border border-emerald-200 bg-emerald-50/80 p-4 shadow-sm">
              <h2 className="font-medium text-emerald-950">Accept quoted price</h2>
              <p className="mt-1 text-sm text-emerald-900/90">
                The owner quoted <strong>₹{booking.final_booking_price} per day</strong> (exclusive of platform fee and
                GST). Accept when you are happy with this amount so they can confirm the trip. You can still chat to
                negotiate before accepting.
              </p>
              <NegotiationPricePreview booking={booking} tripDays={tripDays} variant="customer" />
              <button
                type="button"
                onClick={() => void acceptQuotedPrice()}
                className="mt-3 min-h-[44px] rounded-md bg-emerald-700 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-800"
              >
                I accept ₹{booking.final_booking_price}/day
              </button>
            </div>
          )}
          {isCustomer &&
            customerPriceAccepted &&
            (booking.status === "PENDING" || booking.status === "NEGOTIATING") && (
              <div className="rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-900">
                <p>
                  You accepted the quoted price of <strong>₹{booking.final_booking_price} per day</strong>
                  {booking.customer_accepted_price_at
                    ? ` on ${new Date(booking.customer_accepted_price_at).toLocaleString()}`
                    : ""}
                  . Waiting for the owner to confirm the booking.
                </p>
                <NegotiationPricePreview booking={booking} tripDays={tripDays} variant="customer" />
              </div>
            )}

          {isOwner && booking.status !== "CONFIRMED" && booking.status !== "CANCELLED" && (
            <div className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
              <h2 className="font-medium">Final price & confirmation</h2>
              <p className="text-sm text-slate-600">
                Set your quoted amount <strong>per day</strong>, exclusive of platform fee and GST. The customer must
                accept it before you can confirm. Only you can confirm the booking.
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
              {booking.final_booking_price && (
                <NegotiationPricePreview booking={booking} tripDays={tripDays} variant="owner" />
              )}
              {booking.final_booking_price && !customerPriceAccepted && (
                <p className="mt-4 rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-950">
                  Waiting for the customer to accept your quote of <strong>₹{booking.final_booking_price} per day</strong>
                  .
                  Confirm will unlock after they accept (chat is still open if they want changes).
                </p>
              )}
              {customerPriceAccepted && (
                <p className="mt-4 rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-900">
                  Customer accepted <strong>₹{booking.final_booking_price} per day</strong>
                  {booking.customer_accepted_price_at
                    ? ` on ${new Date(booking.customer_accepted_price_at).toLocaleString()}`
                    : ""}
                  . You can confirm the booking now.
                </p>
              )}
              {ownerCanConfirm && (
                <button
                  type="button"
                  onClick={() => void confirm()}
                  className="mt-4 min-h-[44px] w-full rounded-md bg-emerald-700 px-4 py-2 text-sm text-white hover:bg-emerald-800 sm:w-auto sm:min-h-0"
                >
                  Confirm booking
                </button>
              )}
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

          {booking.cancellation && (
            <div className="rounded-lg border border-slate-200 bg-slate-50 p-4 text-sm text-slate-700">
              <p className="font-medium text-slate-900">Booking cancelled</p>
              <p className="mt-1">
                By {booking.cancellation.cancelled_by_role} on{" "}
                {new Date(booking.cancellation.cancelled_at).toLocaleString()}
              </p>
              {booking.cancellation.reason ? (
                <p className="mt-2 whitespace-pre-wrap text-slate-600">{booking.cancellation.reason}</p>
              ) : null}
            </div>
          )}

          {showCancelBlock && (
            <div className="rounded-lg border border-red-200 bg-red-50/80 p-4 shadow-sm">
              <h2 className="font-medium text-red-950">Cancel booking</h2>
              <p className="mt-1 text-sm text-red-900/90">
                {booking.status === "CONFIRMED" && payStatus === "UNPAID"
                  ? "Cancels this confirmed trip before any deposit is paid. Use only if plans changed."
                  : "Ends this inquiry or negotiation. The other party will see this thread as cancelled."}
              </p>
              <label className="mt-3 block text-sm text-red-950">
                Reason (optional)
                <textarea
                  className="mt-1 w-full rounded-md border border-red-200 bg-white px-3 py-2 text-sm text-slate-900"
                  rows={2}
                  value={cancelReason}
                  onChange={(e) => setCancelReason(e.target.value)}
                  placeholder="e.g. dates no longer work"
                />
              </label>
              <button
                type="button"
                onClick={() => void cancelBooking()}
                className="mt-3 min-h-[44px] rounded-md bg-red-800 px-4 py-2 text-sm font-medium text-white hover:bg-red-900"
              >
                Cancel booking
              </button>
            </div>
          )}

          {showCustomerAwaitingOwner && (
            <div className="rounded-lg border border-slate-200 bg-slate-50 p-4 shadow-sm">
              <h2 className="font-medium text-slate-900">Awaiting vehicle handover</h2>
              <p className="mt-2 text-sm text-slate-700">
                Your deposit is paid. The owner will record odometer and condition when they hand you the keys. You will
                then confirm pickup check-in here before driving off.
              </p>
            </div>
          )}

          {showOwnerHandoverForm && (
            <div className="rounded-lg border border-sky-200 bg-white p-4 shadow-sm">
              <h2 className="font-medium text-slate-900">Vehicle handover (owner)</h2>
              <p className="mt-1 text-sm text-slate-600">
                When you hand over the keys, log odometer and optional fuel level. The customer must accept this record
                before the trip starts.
              </p>
              <div className="mt-3 grid gap-2 sm:grid-cols-2">
                <label className="text-sm text-slate-700">
                  Odometer (km)
                  <input
                    className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                    inputMode="numeric"
                    value={pickupOdom}
                    onChange={(e) => setPickupOdom(e.target.value)}
                  />
                </label>
                <label className="text-sm text-slate-700">
                  Fuel % (optional)
                  <input
                    className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                    inputMode="numeric"
                    value={pickupFuel}
                    onChange={(e) => setPickupFuel(e.target.value)}
                  />
                </label>
                <label className="sm:col-span-2 text-sm text-slate-700">
                  Notes
                  <input
                    className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                    value={pickupNotesH}
                    onChange={(e) => setPickupNotesH(e.target.value)}
                  />
                </label>
                {id ? (
                  <HandoverPhotoField
                    bookingId={id}
                    step="owner_pickup"
                    photos={handoverPhotos}
                    pendingRef={ownerPickupPhotosRef}
                  />
                ) : null}
                <button
                  type="button"
                  onClick={() => void submitHandover("pickup")}
                  className="sm:col-span-2 min-h-[44px] rounded-md bg-sky-900 px-4 py-2 text-sm text-white hover:bg-sky-800"
                >
                  Record vehicle handover
                </button>
              </div>
            </div>
          )}

          {showHandoverCard && (
            <div className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
              <h2 className="font-medium text-slate-900">Pickup & return handover</h2>
              {isCustomer && (
                <div className="mt-3 rounded-md border border-emerald-200 bg-emerald-50/80 px-3 py-3 text-sm text-emerald-950">
                  <p className="font-medium">Your next step (customer)</p>
                  {!customerPickupDone && (
                    <p className="mt-1 text-emerald-900/90">
                      The owner logged handover at{" "}
                      {booking.handover?.owner_pickup_recorded_at
                        ? new Date(booking.handover.owner_pickup_recorded_at).toLocaleString()
                        : "pickup"}
                      . Verify odometer and condition, then submit your pickup check-in before you drive off.
                    </p>
                  )}
                  {pickupComplete && !customerReturnDone && (
                    <p className="mt-1 text-emerald-900/90">
                      When you return the car, log return odometer (and fuel if you want). The owner will bill any extras
                      and must confirm return after you pay the final balance.
                    </p>
                  )}
                  {customerReturnDone && !isFullyPaid && (
                    <p className="mt-1 text-emerald-900/90">
                      Return is logged. After the rental ends, the owner may add post-trip charges; pay the final balance
                      when the bill is ready.
                    </p>
                  )}
                  {customerReturnDone && isFullyPaid && !ownerReturnAccepted && (
                    <p className="mt-1 text-emerald-900/90">
                      Final payment received. Waiting for the owner to accept your return and close the trip.
                    </p>
                  )}
                  {returnComplete && (
                    <p className="mt-1 text-emerald-900/90">
                      Trip is complete. You can leave a review after the rental end date.
                    </p>
                  )}
                </div>
              )}
              {isOwner && (
                <div className="mt-3 rounded-md border border-sky-200 bg-sky-50/80 px-3 py-3 text-sm text-sky-950">
                  <p className="font-medium">Your next step (owner)</p>
                  {!customerPickupDone && (
                    <p className="mt-1 text-sky-900/90">
                      Waiting for the customer to confirm pickup check-in against your handover record.
                    </p>
                  )}
                  {pickupComplete && !customerReturnDone && (
                    <p className="mt-1 text-sky-900/90">
                      Waiting for the customer to log return check-in when they bring the car back.
                    </p>
                  )}
                  {customerReturnDone && !isFullyPaid && !ownerReturnAccepted && payStatus !== "FINAL_DUE" && (
                    <p className="mt-1 text-sky-900/90">
                      Customer return is logged. Add any tolls, fines, or damage in <strong>Post-trip charges</strong> below
                      (use an empty list if there are none). The customer can pay the final balance once you submit.
                    </p>
                  )}
                  {customerReturnDone && payStatus === "FINAL_DUE" && !isFullyPaid && (
                    <p className="mt-1 text-sky-900/90">
                      Post-trip bill sent. You can revise charges below until the customer pays the final balance.
                    </p>
                  )}
                  {customerReturnDone && isFullyPaid && !ownerReturnAccepted && (
                    <p className="mt-1 text-sky-900/90">
                      Final payment received. Review the customer&apos;s return record below and accept it to complete the
                      trip.
                    </p>
                  )}
                  {returnComplete && (
                    <p className="mt-1 text-sky-900/90">Return accepted — this trip is complete.</p>
                  )}
                </div>
              )}
              {booking.handover?.owner_pickup_recorded_at && (
                <p className="mt-3 text-sm text-slate-700">
                  Owner handover{" "}
                  {booking.handover.owner_pickup_odometer_km != null
                    ? `· ${booking.handover.owner_pickup_odometer_km} km`
                    : ""}
                  {booking.handover.owner_pickup_fuel_percent != null
                    ? ` · fuel ${booking.handover.owner_pickup_fuel_percent}%`
                    : ""}{" "}
                  · {new Date(booking.handover.owner_pickup_recorded_at).toLocaleString()}
                </p>
              )}
              {customerPickupDone ? (
                <p className="mt-2 text-sm text-slate-700">
                  Customer pickup check-in{" "}
                  {booking.handover?.pickup_odometer_km != null ? `· ${booking.handover.pickup_odometer_km} km` : ""}
                  {booking.handover?.pickup_fuel_percent != null
                    ? ` · fuel ${booking.handover.pickup_fuel_percent}%`
                    : ""}{" "}
                  · {new Date(booking.handover!.pickup_recorded_at!).toLocaleString()}
                </p>
              ) : isCustomer ? (
                <div className="mt-3 grid gap-2 sm:grid-cols-2">
                  <label className="text-sm text-slate-700">
                    Odometer (km)
                    <input
                      className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                      inputMode="numeric"
                      value={pickupOdom}
                      onChange={(e) => setPickupOdom(e.target.value)}
                    />
                  </label>
                  <label className="text-sm text-slate-700">
                    Fuel % (optional)
                    <input
                      className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                      inputMode="numeric"
                      value={pickupFuel}
                      onChange={(e) => setPickupFuel(e.target.value)}
                    />
                  </label>
                  <label className="sm:col-span-2 text-sm text-slate-700">
                    Notes
                    <input
                      className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                      value={pickupNotesH}
                      onChange={(e) => setPickupNotesH(e.target.value)}
                    />
                  </label>
                  {id ? (
                    <HandoverPhotoField
                      bookingId={id}
                      step="customer_pickup"
                      photos={handoverPhotos}
                      pendingRef={customerPickupPhotosRef}
                    />
                  ) : null}
                  <button
                    type="button"
                    onClick={() => void submitHandover("pickup")}
                    className="sm:col-span-2 min-h-[44px] rounded-md bg-emerald-800 px-4 py-2 text-sm text-white hover:bg-emerald-900"
                  >
                    Confirm pickup check-in
                  </button>
                </div>
              ) : null}
              {pickupComplete && (
              <>
              {customerReturnDone ? (
                <p className="mt-4 border-t border-slate-100 pt-4 text-sm text-slate-700">
                  Customer return{" "}
                  {booking.handover!.return_odometer_km != null ? `· ${booking.handover!.return_odometer_km} km` : ""}
                  {booking.handover!.return_fuel_percent != null
                    ? ` · fuel ${booking.handover!.return_fuel_percent}%`
                    : ""}{" "}
                  · {new Date(booking.handover!.return_recorded_at!).toLocaleString()}
                </p>
              ) : isCustomer ? (
                <div className="mt-4 border-t border-slate-100 pt-4">
                  <p className="text-sm font-medium text-slate-800">Return check-in (customer)</p>
                  <div className="mt-2 grid gap-2 sm:grid-cols-2">
                    <label className="text-sm text-slate-700">
                      Odometer (km)
                      <input
                        className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                        inputMode="numeric"
                        value={returnOdom}
                        onChange={(e) => setReturnOdom(e.target.value)}
                      />
                    </label>
                    <label className="text-sm text-slate-700">
                      Fuel % (optional)
                      <input
                        className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                        inputMode="numeric"
                        value={returnFuel}
                        onChange={(e) => setReturnFuel(e.target.value)}
                      />
                    </label>
                    <label className="sm:col-span-2 text-sm text-slate-700">
                      Notes
                      <input
                        className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                        value={returnNotesH}
                        onChange={(e) => setReturnNotesH(e.target.value)}
                      />
                    </label>
                    {id ? (
                      <HandoverPhotoField
                        bookingId={id}
                        step="customer_return"
                        photos={handoverPhotos}
                        pendingRef={customerReturnPhotosRef}
                      />
                    ) : null}
                    <button
                      type="button"
                      onClick={() => void submitHandover("return")}
                      className="sm:col-span-2 min-h-[44px] rounded-md bg-emerald-800 px-4 py-2 text-sm text-white hover:bg-emerald-900"
                    >
                      Save return check-in
                    </button>
                  </div>
                </div>
              ) : null}
              {ownerReturnAccepted && booking.handover?.owner_return_accepted_at && (
                <p className="mt-2 text-sm text-slate-700">
                  Owner accepted return · {new Date(booking.handover.owner_return_accepted_at).toLocaleString()}
                </p>
              )}
              {isOwner && customerReturnDone && isFullyPaid && !ownerReturnAccepted && (
                <div className="mt-4 border-t border-slate-100 pt-4">
                  {id ? (
                    <HandoverPhotoField
                      bookingId={id}
                      step="owner_return"
                      photos={handoverPhotos}
                      pendingRef={ownerReturnPhotosRef}
                    />
                  ) : null}
                  <button
                    type="button"
                    onClick={() => void submitOwnerAcceptReturn()}
                    className="mt-3 min-h-[44px] rounded-md bg-sky-900 px-4 py-2 text-sm font-medium text-white hover:bg-sky-800"
                  >
                    Accept return &amp; complete trip
                  </button>
                </div>
              )}
              </>
              )}
            </div>
          )}

          {showOwnerSettlement && (
            <div className="rounded-lg border border-amber-200 bg-amber-50/50 p-4 shadow-sm">
              <h2 className="font-medium text-amber-950">Post-trip charges</h2>
              <p className="mt-1 text-sm text-amber-900/90">
                The customer logged return. Add documented costs (for example Fastag /
                tolls, traffic fines, scratches or other damage). The customer&apos;s final bill is the remaining trip
                balance plus these lines. You can revise this list until they pay the final balance.
              </p>
              <div className="mt-3 space-y-3">
                {chargeDrafts.map((row, idx) => (
                  <div key={idx} className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_140px_auto] sm:items-end">
                    <label className="text-sm text-slate-800">
                      Description
                      <input
                        className="mt-1 w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm"
                        value={row.label}
                        placeholder="e.g. Fastag / scratch repair"
                        onChange={(e) => {
                          const next = [...chargeDrafts];
                          next[idx] = { ...next[idx], label: e.target.value };
                          setChargeDrafts(next);
                        }}
                      />
                    </label>
                    <label className="text-sm text-slate-800">
                      Amount (INR)
                      <input
                        className="mt-1 w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm"
                        inputMode="decimal"
                        value={row.amount}
                        placeholder="0"
                        onChange={(e) => {
                          const next = [...chargeDrafts];
                          next[idx] = { ...next[idx], amount: e.target.value };
                          setChargeDrafts(next);
                        }}
                      />
                    </label>
                    <div className="flex items-end pb-1">
                      {chargeDrafts.length > 1 ? (
                        <button
                          type="button"
                          className="text-sm text-red-800 hover:underline"
                          onClick={() => setChargeDrafts(chargeDrafts.filter((_, i) => i !== idx))}
                        >
                          Remove
                        </button>
                      ) : (
                        <span className="text-xs text-slate-500 sm:pl-1"> </span>
                      )}
                    </div>
                  </div>
                ))}
              </div>
              <div className="mt-4 flex flex-wrap gap-2">
                <button
                  type="button"
                  className="min-h-[40px] rounded-md border border-amber-800/40 bg-white px-3 py-2 text-sm text-amber-950 hover:bg-amber-100"
                  onClick={() => setChargeDrafts([...chargeDrafts, { label: "", amount: "" }])}
                >
                  Add line
                </button>
                <button
                  type="button"
                  disabled={settlementBusy}
                  onClick={() => void submitPostTripCharges()}
                  className="min-h-[40px] rounded-md bg-amber-900 px-4 py-2 text-sm font-medium text-white hover:bg-amber-950 disabled:opacity-60"
                >
                  {settlementBusy ? "Saving…" : "Save charges & notify customer"}
                </button>
              </div>
            </div>
          )}

          {!!booking.reviews?.length && (
            <div className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
              <h2 className="font-medium text-slate-900">Reviews</h2>
              <ul className="mt-2 space-y-3 text-sm">
                {booking.reviews.map((r) => (
                  <li key={`${r.party}-${r.created_at}`} className="rounded-md border border-slate-100 bg-slate-50/80 px-3 py-2">
                    <p className="font-medium text-slate-900">
                      {r.party === "CUSTOMER" ? "Customer" : "Owner"} · {r.rating}/5 · {r.reviewer.full_name}
                    </p>
                    {r.comment ? <p className="mt-1 text-slate-700">{r.comment}</p> : null}
                    <p className="mt-1 text-xs text-slate-500">{new Date(r.created_at).toLocaleString()}</p>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {canPostReview && (
            <div className="rounded-lg border border-emerald-200 bg-emerald-50/60 p-4 shadow-sm">
              <h2 className="font-medium text-emerald-950">Rate this trip</h2>
              <p className="mt-1 text-sm text-emerald-900/90">Available after rental end, once the final payment is complete.</p>
              <label className="mt-3 block text-sm text-emerald-950">
                Rating (1–5)
                <select
                  className="mt-1 w-full max-w-xs rounded-md border border-emerald-200 bg-white px-3 py-2 text-sm"
                  value={reviewRating}
                  onChange={(e) => setReviewRating(Number(e.target.value))}
                >
                  {[5, 4, 3, 2, 1].map((n) => (
                    <option key={n} value={n}>
                      {n}
                    </option>
                  ))}
                </select>
              </label>
              <label className="mt-3 block text-sm text-emerald-950">
                Comment (optional)
                <textarea
                  className="mt-1 w-full rounded-md border border-emerald-200 bg-white px-3 py-2 text-sm text-slate-900"
                  rows={3}
                  value={reviewComment}
                  onChange={(e) => setReviewComment(e.target.value)}
                />
              </label>
              <button
                type="button"
                onClick={() => void submitReview()}
                className="mt-3 min-h-[44px] rounded-md bg-emerald-800 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-900"
              >
                Submit review
              </button>
            </div>
          )}
        </div>

        <aside className="min-w-0 lg:sticky lg:top-20 lg:self-start">
          <div className="flex flex-col rounded-lg border border-slate-200 bg-white p-4 shadow-sm lg:max-h-[calc(100dvh-5.5rem)] lg:min-h-[min(30rem,calc(100dvh-5.5rem))]">
            <div className="shrink-0">
              <h2 className="font-medium">{tripStageLabel || "Chat"}</h2>
              {tripStageLabel ? (
                <p className="mt-0.5 text-xs text-slate-500">Messages with your booking partner</p>
              ) : null}
            </div>
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
                  className="min-h-[44px] min-w-0 flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm disabled:bg-slate-100 disabled:text-slate-500"
                  placeholder={
                    booking.status === "CANCELLED"
                      ? "Chat closed (booking cancelled)"
                      : isTripCompleted(booking.status)
                        ? "Chat closed (trip completed)"
                        : "Type a message"
                  }
                  value={text}
                  disabled={booking.status === "CANCELLED" || isTripCompleted(booking.status)}
                  onChange={(e) => setText(e.target.value)}
                  onKeyDown={(e) => {
                    if (booking.status === "CANCELLED" || isTripCompleted(booking.status)) return;
                    if (e.key === "Enter" && !e.shiftKey) {
                      e.preventDefault();
                      void sendMessage();
                    }
                  }}
                />
                <button
                  type="button"
                  disabled={booking.status === "CANCELLED" || isTripCompleted(booking.status)}
                  onClick={() => void sendMessage()}
                  className="min-h-[44px] shrink-0 rounded-md bg-slate-900 px-4 py-2 text-sm text-white hover:bg-slate-800 disabled:opacity-50"
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
