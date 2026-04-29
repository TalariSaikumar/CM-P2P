"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { apiJson, ApiError } from "@/lib/api";
import { getToken, getUser } from "@/lib/session";
import type { Car } from "@/lib/apitypes";

export default function CustomerSearchPage() {
  const router = useRouter();
  const [location, setLocation] = useState("");
  const [model, setModel] = useState("");
  const [cars, setCars] = useState<Car[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function search() {
    setError(null);
    setLoading(true);
    try {
      const qs = new URLSearchParams();
      if (location.trim()) qs.set("location", location.trim());
      if (model.trim()) qs.set("model", model.trim());
      const res = await apiJson<{ cars: Car[] }>(`/cars?${qs.toString()}`);
      setCars(res.cars || []);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Search failed");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void search();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function inquire(carId: string) {
    const token = getToken();
    const user = getUser();
    if (!token || !user) {
      router.push("/login");
      return;
    }
    if (user.role !== "CUSTOMER") {
      setError("Only customers can send booking inquiries.");
      return;
    }
    if (!user.is_kyc_verified) {
      setError("Complete KYC on the Account page before booking.");
      return;
    }
    if (!user.driving_license_number) {
      setError("Add your driving license number on the Account page before booking.");
      return;
    }
    setError(null);
    setLoading(true);
    try {
      const res = await apiJson<{ booking: { id: string } }>("/bookings", {
        method: "POST",
        body: JSON.stringify({ car_id: carId, customer_note: "" }),
      });
      router.push(`/bookings/${res.booking.id}`);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not create inquiry");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="page-shell max-w-4xl space-y-6">
      <div>
        <h1 className="text-xl font-semibold sm:text-2xl">Search cars</h1>
        <p className="text-sm text-slate-600">Filter by location and model, then send a booking inquiry.</p>
      </div>
      <div className="flex flex-col gap-3 rounded-lg border border-slate-200 bg-white p-4 shadow-sm sm:flex-row sm:flex-wrap sm:items-end">
        <input
          placeholder="Location (city or area)"
          className="min-h-[44px] min-w-0 flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm sm:min-w-[140px]"
          value={location}
          onChange={(e) => setLocation(e.target.value)}
        />
        <input
          placeholder="Model keyword"
          className="min-h-[44px] min-w-0 flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm sm:min-w-[120px]"
          value={model}
          onChange={(e) => setModel(e.target.value)}
        />
        <button
          type="button"
          onClick={() => void search()}
          disabled={loading}
          className="min-h-[44px] w-full shrink-0 rounded-md bg-slate-900 px-4 py-2 text-sm text-white hover:bg-slate-800 disabled:opacity-60 sm:w-auto"
        >
          {loading ? "Searching…" : "Search"}
        </button>
      </div>
      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
      )}
      <ul className="space-y-3">
        {cars.map((c) => (
          <li
            key={c.id}
            className="flex flex-col gap-3 rounded-lg border border-slate-200 bg-white p-4 shadow-sm sm:flex-row sm:items-center sm:justify-between"
          >
            <div>
              <p className="font-medium text-slate-900">
                {c.car_name} · {c.car_model}
              </p>
              <p className="text-sm text-slate-600">
                {c.location} · Plate {c.car_number}
              </p>
              <p className="text-sm text-slate-600">
                From ₹{c.price_per_day}/day · ₹{c.price_per_hour}/hr · ₹{c.price_per_km}/km
              </p>
            </div>
            <div className="flex w-full shrink-0 sm:w-auto">
              <button
                type="button"
                onClick={() => void inquire(c.id)}
                className="min-h-[44px] w-full rounded-md border border-slate-300 px-3 py-2 text-sm hover:bg-slate-50 sm:min-h-0 sm:w-auto sm:py-1.5"
              >
                Booking inquiry
              </button>
            </div>
          </li>
        ))}
        {!cars.length && !loading && (
          <p className="text-sm text-slate-600">No cars match your filters yet.</p>
        )}
      </ul>
    </main>
  );
}
