"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { apiJson, ApiError } from "@/lib/api";
import { getToken } from "@/lib/session";
import type { Booking } from "@/lib/apitypes";

export default function CustomerBookingsPage() {
  const router = useRouter();
  const [rows, setRows] = useState<Booking[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function load() {
    setError(null);
    try {
      const res = await apiJson<{ bookings: Booking[] }>("/bookings/mine");
      setRows(res.bookings || []);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not load bookings");
    }
  }

  return (
    <main className="mx-auto max-w-3xl space-y-4 p-8">
      <h1 className="text-2xl font-semibold">My bookings</h1>
      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
      )}
      <ul className="space-y-3">
        {rows.map((b) => (
          <li key={b.id} className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div>
                <p className="font-medium">
                  {b.car.car_name} · {b.car.car_model} ({b.car.car_number})
                </p>
                <p className="text-sm text-slate-600">
                  Status: <span className="font-medium text-slate-900">{b.status}</span>
                  {b.final_booking_price && (
                    <span className="ml-2">· Agreed ₹{b.final_booking_price}</span>
                  )}
                </p>
              </div>
              <Link
                className="rounded-md bg-slate-900 px-3 py-1.5 text-sm text-white hover:bg-slate-800"
                href={`/bookings/${b.id}`}
              >
                Open chat
              </Link>
            </div>
          </li>
        ))}
        {!rows.length && !error && (
          <p className="text-sm text-slate-600">You have no bookings yet. Start from Search cars.</p>
        )}
      </ul>
    </main>
  );
}
