"use client";

import { useEffect, useState, useRef } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { apiForm, apiJson, ApiError } from "@/lib/api";
import { getToken } from "@/lib/session";
import type { Car } from "@/lib/apitypes";

export default function OwnerFleetPage() {
  const router = useRouter();
  const [cars, setCars] = useState<Car[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [msg, setMsg] = useState<string | null>(null);
  const fileRefs = useRef<Record<string, HTMLInputElement | null>>({});

  async function load() {
    setError(null);
    try {
      const res = await apiJson<{ cars: Car[] }>("/cars/mine");
      setCars(res.cars || []);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not load fleet");
    }
  }

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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
      <ul className="space-y-4">
        {cars.map((c) => (
          <li key={c.id} className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
            <div className="flex flex-col gap-3 sm:flex-row sm:justify-between">
              <div className="min-w-0 flex-1">
                <p className="font-medium break-words">
                  {c.car_name} · {c.car_model}
                </p>
                <p className="text-sm text-slate-600">
                  {c.location} · {c.car_number} · {c.is_active ? "Active" : "Inactive"}
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
        {!cars.length && !error && (
          <p className="text-sm text-slate-600">No vehicles yet. Add your first listing.</p>
        )}
      </ul>
    </main>
  );
}
