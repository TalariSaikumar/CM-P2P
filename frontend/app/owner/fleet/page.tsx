"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import { apiForm, apiJson, ApiError } from "@/lib/api";
import { getToken } from "@/lib/session";
import Link from "next/link";
import type { CarMineRow } from "@/lib/apitypes";
import { PaginationBar } from "@/components/PaginationBar";
import { CarRoadLoader, PageLoader } from "@/components/loaders";

const PER_PAGE = 20;

export default function OwnerFleetPage() {
  const router = useRouter();
  const [cars, setCars] = useState<CarMineRow[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [error, setError] = useState<string | null>(null);
  const [msg, setMsg] = useState<string | null>(null);
  const [listBusy, setListBusy] = useState(true);
  const fileRefs = useRef<Record<string, HTMLInputElement | null>>({});

  const load = useCallback(async () => {
    setListBusy(true);
    setError(null);
    try {
      const res = await apiJson<{ cars: CarMineRow[]; total?: number }>(
        `/cars/mine?page=${page}&per_page=${PER_PAGE}`,
      );
      const t = res.total ?? res.cars?.length ?? 0;
      setTotal(t);
      const totalPages = Math.max(1, Math.ceil(t / PER_PAGE));
      if (page > totalPages) {
        setPage(totalPages);
        return;
      }
      setCars(res.cars || []);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not load fleet");
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

  async function upload(carId: string, file: File | null) {
    if (!file) return;
    setMsg(null);
    setError(null);
    const fd = new FormData();
    fd.append("file", file);
    try {
      await apiForm(`/cars/${carId}/images`, fd);
      setMsg("Image uploaded.");
      await load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Upload failed");
    }
  }

  return (
    <main className="page-shell max-w-4xl space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
        <h1 className="text-xl font-semibold sm:text-2xl">My fleet</h1>
        <Link
          href="/owner/cars/new"
          className="inline-flex min-h-[44px] w-full shrink-0 items-center justify-center rounded-md bg-slate-900 px-3 py-2 text-sm text-white hover:bg-slate-800 sm:w-auto sm:py-1.5"
        >
          Add car
        </Link>
      </div>
      {msg && (
        <div className="rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-900">
          {msg}
        </div>
      )}
      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
      )}
      {listBusy && cars.length > 0 && (
        <div className="flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs text-slate-600 shadow-sm">
          <CarRoadLoader size="sm" className="!w-14" />
          <span>Refreshing fleet…</span>
        </div>
      )}
      {listBusy && cars.length === 0 ? (
        <PageLoader title="Loading your fleet…" subtitle="Listings, pricing, and photos." className="min-h-[280px] py-8" />
      ) : null}
      {!listBusy || cars.length > 0 ? (
      <ul className="space-y-4">
        {cars.map((c) => (
          <li key={c.id} className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
            <div className="flex flex-col gap-3 sm:flex-row sm:justify-between">
              <div className="min-w-0 flex-1">
                <p className="font-medium break-words">
                  {c.car_name} · {c.car_model}
                </p>
                <p className="text-sm text-slate-600">
                  {c.model_year} · {c.color} · {c.fuel_type} · {c.transmission} · {c.mileage_km.toLocaleString()} km ·{" "}
                  {c.num_seats} seats
                </p>
                <p className="text-sm text-slate-600">
                  {c.location} · {c.car_number} · {c.is_active ? "Active" : "Inactive"}
                  {c.booked_for_current_date ? (
                    <span className="ml-2 inline-block rounded bg-amber-100 px-1.5 py-0.5 text-xs font-medium text-amber-900">
                      Booked today (UTC)
                    </span>
                  ) : null}
                </p>
                <p className="text-sm text-slate-600">
                  ₹{c.price_per_day}/day · ₹{c.price_per_hour}/hr · ₹{c.price_per_km}/km
                </p>
                {!!c.images?.length && (
                  <div className="mt-2 flex flex-wrap gap-2">
                    {c.images.map((im) => (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img
                        key={im.id}
                        src={im.url}
                        alt=""
                        className="h-20 w-28 rounded-md object-cover"
                      />
                    ))}
                  </div>
                )}
              </div>
              <div className="flex w-full min-w-0 flex-col items-stretch gap-2 sm:w-auto sm:items-start">
                <Link
                  href={`/owner/cars/${c.id}/edit`}
                  className="inline-flex min-h-[44px] items-center justify-center rounded-md border border-slate-300 px-3 py-2 text-center text-sm font-medium text-slate-800 hover:bg-slate-50 sm:min-h-0 sm:py-1.5"
                >
                  Edit details
                </Link>
                <input
                  type="file"
                  accept="image/*"
                  ref={(el) => {
                    fileRefs.current[c.id] = el;
                  }}
                  className="max-w-full min-w-0 text-sm file:mr-2 file:rounded file:border-0 file:bg-slate-100 file:px-2 file:py-1.5 file:text-sm"
                />
                <button
                  type="button"
                  className="min-h-[44px] rounded-md border border-slate-300 px-3 py-2 text-sm hover:bg-slate-50 sm:min-h-0 sm:py-1.5"
                  onClick={() => upload(c.id, fileRefs.current[c.id]?.files?.[0] || null)}
                >
                  Upload photo
                </button>
              </div>
            </div>
          </li>
        ))}
        {!cars.length && !error && !listBusy && (
          <p className="text-sm text-slate-600">No vehicles yet. Add your first listing.</p>
        )}
      </ul>
      ) : null}
      {!(listBusy && cars.length === 0) && (
        <PaginationBar page={page} perPage={PER_PAGE} total={total} onPageChange={setPage} noun="vehicles" />
      )}
    </main>
  );
}
