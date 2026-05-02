"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { apiJson, ApiError } from "@/lib/api";
import { getToken } from "@/lib/session";
import type { Booking } from "@/lib/apitypes";
import { PaginationBar } from "@/components/PaginationBar";
import { CarRoadLoader, PageLoader } from "@/components/loaders";

const PER_PAGE = 20;

export default function CustomerBookingsPage() {
  const router = useRouter();
  const [rows, setRows] = useState<Booking[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [error, setError] = useState<string | null>(null);
  const [listBusy, setListBusy] = useState(true);

  const load = useCallback(async () => {
    setListBusy(true);
    setError(null);
    try {
      const res = await apiJson<{ bookings: Booking[]; total?: number }>(
        `/bookings/mine?page=${page}&per_page=${PER_PAGE}`,
      );
      const t = res.total ?? res.bookings?.length ?? 0;
      setTotal(t);
      const totalPages = Math.max(1, Math.ceil(t / PER_PAGE));
      if (page > totalPages) {
        setPage(totalPages);
        return;
      }
      setRows(res.bookings || []);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not load bookings");
    } finally {
      setListBusy(false);
    }
  }, [page]);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    void load();
  }, [router, load]);

  return (
    <main className="page-shell max-w-3xl space-y-4">
      <h1 className="text-xl font-semibold sm:text-2xl">My bookings</h1>
      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
      )}
      {listBusy && rows.length > 0 && (
        <div className="flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs text-slate-600 shadow-sm">
          <CarRoadLoader size="sm" className="!w-14" />
          <span>Updating bookings…</span>
        </div>
      )}
      {listBusy && rows.length === 0 ? (
        <PageLoader title="Loading your bookings…" subtitle="Trips, status, and pay links." className="min-h-[280px] py-8" />
      ) : null}
      {!listBusy || rows.length > 0 ? (
      <ul className="space-y-3">
        {rows.map((b) => (
          <li key={b.id} className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
              <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
              <div className="min-w-0">
                <p className="font-medium break-words">
                  {b.car.car_name} · {b.car.car_model} ({b.car.car_number})
                </p>
                <p className="text-sm text-slate-600">
                  Status: <span className="font-medium text-slate-900">{b.status}</span>
                  {b.final_booking_price && (
                    <span className="ml-2">· Agreed ₹{b.final_booking_price}</span>
                  )}
                  {b.payment?.payment_status === "PAID" && (
                    <span className="ml-2 font-medium text-emerald-800">· Paid</span>
                  )}
                </p>
              </div>
              <div className="flex w-full shrink-0 flex-col gap-2 sm:w-auto sm:flex-row">
                {b.status === "CONFIRMED" && b.payment?.payment_status === "UNPAID" && (
                  <Link
                    className="inline-flex min-h-[44px] w-full items-center justify-center rounded-md bg-emerald-700 px-3 py-2 text-sm font-medium text-white hover:bg-emerald-800 sm:w-auto sm:py-1.5"
                    href={`/customer/bookings/${b.id}/pay`}
                  >
                    Pay
                  </Link>
                )}
                <Link
                  className="inline-flex min-h-[44px] w-full shrink-0 items-center justify-center rounded-md bg-slate-900 px-3 py-2 text-sm text-white hover:bg-slate-800 sm:w-auto sm:py-1.5"
                  href={`/bookings/${b.id}`}
                >
                  Open chat
                </Link>
              </div>
            </div>
          </li>
        ))}
        {!rows.length && !error && !listBusy && (
          <p className="text-sm text-slate-600">You have no bookings yet. Start from Search cars.</p>
        )}
      </ul>
      ) : null}
      {!(listBusy && rows.length === 0) ? (
        <PaginationBar page={page} perPage={PER_PAGE} total={total} onPageChange={setPage} noun="bookings" />
      ) : null}
    </main>
  );
}
