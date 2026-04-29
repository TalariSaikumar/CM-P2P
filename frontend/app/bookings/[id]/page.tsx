"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { apiJson, ApiError } from "@/lib/api";
import { getToken, getUser } from "@/lib/session";
import type { Booking, Message } from "@/lib/apitypes";

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
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    void loadBooking();
    void loadMessages();
    const t = setInterval(() => {
      void loadBooking();
      void loadMessages();
    }, 2000);
    return () => clearInterval(t);
  }, [router, loadBooking, loadMessages]);

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
      setInfo("Final price updated. The customer will see it within a few seconds.");
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
      setInfo("Booking confirmed. You should receive an SMS shortly if notifications are configured.");
      await loadBooking();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not confirm booking");
    }
  }

  if (!booking) {
    return (
      <main className="p-8">
        <p className="text-slate-600">Loading booking…</p>
        {bootError && (
          <p className="mt-3 text-sm text-red-700">
            {bootError} — check that you are signed in and allowed to view this booking.
          </p>
        )}
      </main>
    );
  }

  const me = getUser();
  const isOwner = me?.id === booking.owner_id;
  const isCustomer = me?.id === booking.customer_id;

  return (
    <main className="mx-auto max-w-3xl space-y-4 p-8">
      <div>
        <h1 className="text-2xl font-semibold">Booking</h1>
        <p className="text-sm text-slate-600">
          {booking.car.car_name} · {booking.car.car_model} ({booking.car.car_number}) ·{" "}
          <span className="font-medium text-slate-900">{booking.status}</span>
        </p>
        {booking.final_booking_price && (
          <p className="text-sm text-slate-700">
            Final agreed price: <span className="font-semibold">₹{booking.final_booking_price}</span>
          </p>
        )}
        <p className="text-xs text-slate-500">This page refreshes booking and chat every 2 seconds.</p>
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
          <h2 className="font-medium">Update final price</h2>
          <p className="text-sm text-slate-600">Only you can set the agreed amount the customer will confirm.</p>
          <div className="mt-3 flex flex-wrap gap-2">
            <input
              className="min-w-[160px] flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm"
              placeholder="Amount in INR"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
            />
            <button
              type="button"
              onClick={() => void updatePrice()}
              className="rounded-md bg-slate-900 px-4 py-2 text-sm text-white hover:bg-slate-800"
            >
              Save price
            </button>
          </div>
        </div>
      )}

      {isCustomer && booking.status !== "CONFIRMED" && booking.status !== "CANCELLED" && (
        <div className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
          <h2 className="font-medium">Confirm booking</h2>
          <p className="text-sm text-slate-600">
            Once you are happy with the final price shown above, confirm to lock the booking.
          </p>
          <button
            type="button"
            disabled={!booking.final_booking_price}
            onClick={() => void confirm()}
            className="mt-3 rounded-md bg-emerald-700 px-4 py-2 text-sm text-white hover:bg-emerald-800 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Confirm booking
          </button>
        </div>
      )}

      <div className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
        <h2 className="font-medium">Chat</h2>
        <div className="mt-3 max-h-80 space-y-3 overflow-y-auto rounded-md bg-slate-50 p-3 text-sm">
          {messages.map((m) => {
            const isMine = me?.id === m.sender_id;
            return (
              <div key={m.id} className={`flex w-full ${isMine ? "justify-end" : "justify-start"}`}>
                <div
                  className={`max-w-[min(85%,20rem)] rounded-2xl px-3 py-2 shadow-sm ${
                    isMine
                      ? "rounded-br-md bg-slate-900 text-white"
                      : "rounded-bl-md border border-slate-200 bg-white text-slate-900"
                  }`}
                >
                  <p className={`text-xs ${isMine ? "text-slate-400" : "text-slate-500"}`}>
                    {isMine ? "You" : m.sender.full_name} · {new Date(m.created_at).toLocaleString()}
                  </p>
                  <p className={`mt-0.5 whitespace-pre-wrap break-words ${isMine ? "text-white" : "text-slate-900"}`}>
                    {m.body}
                  </p>
                </div>
              </div>
            );
          })}
          {!messages.length && <p className="text-slate-500">No messages yet. Say hello.</p>}
        </div>
        <div className="mt-3 flex gap-2">
          <input
            className="flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm"
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
            className="rounded-md bg-slate-900 px-4 py-2 text-sm text-white hover:bg-slate-800"
          >
            Send
          </button>
        </div>
      </div>
    </main>
  );
}
